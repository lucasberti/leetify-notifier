package main

import (
	"context"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"leetify_notifier/config"
	"leetify_notifier/leetify"
	"leetify_notifier/notifiers"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CONFIG_PATH = "config.json"
	UPDATE_INTERVAL = 30 * time.Second
)

func checkGames(config *config.Config, profile *leetify.Profile, wg *sync.WaitGroup) {
	defer wg.Done()

	latestGame := profile.GetLatestGame()
	latestGameId := latestGame.GameId

	if slices.Contains(config.KnownMatchIds, latestGameId) {
		log.Print("Latest game is already known; skipping...")
		return
	}

	if len(config.KnownMatchIds) == 0 {
		log.Print("No knownMatchIds in config.json, saving latest one...")
		config.KnownMatchIds = []string{latestGameId}
	} else {
		config.KnownMatchIds = append(config.KnownMatchIds, latestGameId)
	}

	config.SaveConfig(CONFIG_PATH)

	notifiers.SendTelegramMessage(config, profile.GetLatestGame().GetGameLink())	
}

func checkHighlights(config *config.Config, profile *leetify.Profile, wg *sync.WaitGroup) {
	defer wg.Done()

	allFriendsIds := profile.GetFriendsSteamIds()

	friendsProfiles := leetify.GetFriendsProfiles(allFriendsIds)

	highlights := make([]leetify.Highlight, len(friendsProfiles))
	highlights = append(highlights, profile.Highlights...)

	for _, friendProfile := range friendsProfiles {
		highlights = append(highlights, friendProfile.Highlights...)
	}

	for _, highlight := range highlights {
		if slices.Contains(config.KnownHighlightIds, highlight.Id) {
			log.Print("Highlight is already known; skipping...")
			continue
		}

		if len(config.KnownHighlightIds) == 0 {
			log.Print("No knownHighlightIds in config.json, saving latest one...")
			config.KnownHighlightIds = []string{highlight.Id}
		} else {
			config.KnownHighlightIds = append(config.KnownHighlightIds, highlight.Id)
		}
		
		config.SaveConfig(CONFIG_PATH)

		notifiers.SendTelegramMessage(config, "NEW HIGHLIGHT: " + highlight.Description)
	}

}

func run(ctx context.Context, config *config.Config, profile *leetify.Profile) {
	ticker := time.NewTicker(UPDATE_INTERVAL)

	for {
		select {
		case <-ticker.C:
			log.Print("Checking for updates...")

			var wg sync.WaitGroup
			wg.Add(2)

			go checkGames(config, profile, &wg)
			go checkHighlights(config, profile, &wg)
		
			wg.Wait()

		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	ctx, cancel := context.WithCancel(context.Background())

	config, err := config.LoadConfig(CONFIG_PATH)

	if err != nil {
		log.Error().Err(err).Msg("Could not load config")
		return
	}

	if config.MainProfile == "" {
		log.Error().Msg("Please set mainProfile in config.json")
		return
	}

	profile, err := leetify.GetProfile(config.MainProfile)

	if err != nil {
		log.Error().Err(err).Msg("Could not get main profile")
		return
	}

	run(ctx, config, profile)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-shutdown
	cancel()
}

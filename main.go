package main

import (
	"bytes"
	"context"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
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

func replaceValues(input string, mapName string, score string, link string) string {
	input = strings.Replace(input, "%MAPNAME%", mapName, -1)
	input = strings.Replace(input, "%SCORE%", score, -1)
	input = strings.Replace(input, "%GAMELINK%", link, -1)

	return input
}

func generateGameMessage(profile *leetify.Profile, config *config.Config) string {
	game := profile.GetLatestGame()
	friends := profile.Teammates

	var message bytes.Buffer

	for _, player := range game.OwnTeamSteam64Ids {
		if mention, ok := config.TelegramUserNames[player]; ok {
			message.WriteString(mention + " ")

			for _, teammate := range friends {
				if teammate.Steam64Id == player {
					rank := strconv.FormatUint(uint64(teammate.Rank.SkillLevel), 10)
					message.WriteString(" (rank " + rank + ")")
				}
			}

			if player == profile.Meta.Steam64Id {
				rank := strconv.FormatUint(uint64(game.SkillLevel), 10)
				message.WriteString(" (rank " + rank + ")")
			}
			message.WriteString("\n")
		}
	}

	score := strconv.Itoa(game.Scores[0]) + " - " + strconv.Itoa(game.Scores[1])
	mapName := game.MapName
	gameLink := game.GetGameLink()

	switch game.MatchResult {
	case "win":
		message.WriteString(replaceValues(config.WinMsg, mapName, score, gameLink))
	
	case "loss":
		message.WriteString(replaceValues(config.LossMsg, mapName, score, gameLink))

	case "tie":
		message.WriteString(replaceValues(config.TieMsg, mapName, score, gameLink))

	default:
		message.WriteString("New match found!\n" + gameLink)
	}

	return message.String()
}

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
		if !slices.Contains(config.KnownMatchIds, latestGameId) && latestGameId != "" {
			config.KnownMatchIds = append(config.KnownMatchIds, latestGameId)
		}
	}

	config.SaveConfig(CONFIG_PATH)

	notifiers.SendTelegramMessage(config, generateGameMessage(profile, config))
}

func checkHighlights(config *config.Config, profile *leetify.Profile, wg *sync.WaitGroup) {
	defer wg.Done()

	allFriendsIds := profile.GetFriendsSteamIds()

	friendsProfiles := leetify.GetFriendsProfiles(allFriendsIds)

	highlights := make([]leetify.Highlight, len(friendsProfiles) + 1)
	highlights = append(highlights, profile.Highlights...)

	for _, friendProfile := range friendsProfiles {
		if len(friendProfile.Highlights) == 0 {
			continue
		}

		highlights = append(highlights, friendProfile.Highlights...)
	}

	for _, highlight := range highlights {
		if highlight.Id == "" {
			continue
		}

		if slices.Contains(config.KnownHighlightIds, highlight.Id) {
			log.Print("Highlight is already known; skipping...")
			continue
		}

		if len(config.KnownHighlightIds) == 0 {
			log.Print("No knownHighlightIds in config.json, saving latest one...")
			config.KnownHighlightIds = []string{highlight.Id}
		} else {
			if !slices.Contains(config.KnownHighlightIds, highlight.Id) {
				config.KnownHighlightIds = append(config.KnownHighlightIds, highlight.Id)
			}
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

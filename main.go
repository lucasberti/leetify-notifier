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

func generateGameMessage(cfg *config.Config, profile *leetify.Profile) string {
	game := profile.GetLatestGame()
	friends := profile.Teammates

	var message bytes.Buffer

	for _, player := range game.OwnTeamSteam64Ids {
		if mention, ok := cfg.TelegramUserNames[player]; ok {
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
		message.WriteString(replaceValues(cfg.WinMsg, mapName, score, gameLink))
	
	case "loss":
		message.WriteString(replaceValues(cfg.LossMsg, mapName, score, gameLink))

	case "tie":
		message.WriteString(replaceValues(cfg.TieMsg, mapName, score, gameLink))

	default:
		message.WriteString("New match found!\n" + gameLink)
	}

	return message.String()
}

func checkGames(cfg *config.Config, profile *leetify.Profile, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Print("Checking games...")
	defer log.Print("Finished checking games")

	latestGame := profile.GetLatestGame()

	if slices.Contains(cfg.KnownMatchIds, latestGame.GameId) {
		return
	}

	if latestGame.GameId != "" {
		cfg.KnownMatchIds = append(cfg.KnownMatchIds, latestGame.GameId)
	}

	cfg.SaveConfig(CONFIG_PATH)

	notifiers.SendTelegramMessage(cfg, generateGameMessage(cfg, profile))
}

func checkHighlights(cfg *config.Config, profile *leetify.Profile, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Print("Checking highlights...")
	defer log.Print("Finished checking highlights")

	allFriendsIDs := profile.GetFriendsSteamIds()
	friendsProfiles := leetify.GetFriendsProfiles(allFriendsIDs)

	highlights := append([]leetify.Highlight{}, profile.Highlights...)

	for _, friendProfile := range friendsProfiles {
		highlights = append(highlights, friendProfile.Highlights...)
	}

	for _, highlight := range highlights {
		if highlight.Id == "" || slices.Contains(cfg.KnownHighlightIds, highlight.Id){
			continue
		}

		cfg.KnownHighlightIds = append(cfg.KnownHighlightIds, highlight.Id)
		cfg.SaveConfig(CONFIG_PATH)

		notifiers.SendTelegramMessage(cfg, "NEW HIGHLIGHT: " + highlight.Description)
	}

}

func run(ctx context.Context, cfg *config.Config) {
	ticker := time.NewTicker(UPDATE_INTERVAL)

	for {
		select {
		case <-ticker.C:
			log.Print("Checking for updates...")

			profile, err := leetify.GetProfile(cfg.MainProfile)

			if err != nil {
				log.Error().Err(err).Msg("Could not get main profile")
				return
			}

			var wg sync.WaitGroup
			wg.Add(2)

			go checkGames(cfg, profile, &wg)
			go checkHighlights(cfg, profile, &wg)
		
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

	cfg, err := config.LoadConfig(CONFIG_PATH)

	if err != nil {
		log.Error().Err(err).Msg("Could not load config")
		return
	}

	if cfg.MainProfile == "" {
		log.Error().Msg("Please set mainProfile in config.json")
		return
	}

	run(ctx, cfg)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-shutdown
	cancel()
}

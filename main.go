package main

import (
	"os"
	"slices"
	"time"

	"leetify_notifier/config"
	"leetify_notifier/leetify"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CONFIG_PATH = "config.json"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

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
	
	allFriendsIds := profile.GetFriendsSteamIds()
	latestGameId := profile.GetLatestGame().GameId

	if slices.Contains(config.KnownMatchIds, latestGameId) {
		log.Print("Latest game is already known; skipping...")
		// return
	}

	if len(config.KnownMatchIds) == 0 {
		log.Print("No knownMatchIds in config.json, saving latest one...")
		config.KnownMatchIds = []string{latestGameId}
		config.SaveConfig(CONFIG_PATH)
	}

	friendsProfiles := leetify.GetFriendsProfiles(allFriendsIds)

	log.Print(friendsProfiles)
}

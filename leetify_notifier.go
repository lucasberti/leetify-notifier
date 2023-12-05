package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CONFIG_PATH = "config.json"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	config, err := LoadConfig(CONFIG_PATH)

	if err != nil {
		log.Error().Err(err).Msg("Could not load config")
		return
	}

	if config.MainProfile == "" {
		log.Error().Msg("Please set mainProfile in config.json")
		return
	}

	profile, err := GetProfile(config.MainProfile)

	if err != nil {
		log.Error().Err(err).Msg("Could not get main profile")
		return
	}
	
	allFriendsIds := profile.GetFriendsSteamIds()

	friendsProfiles := GetFriendsProfiles(allFriendsIds)

	log.Print(friendsProfiles)
}

package main

import (
	"slices"
	"sync"

	"github.com/lucasberti/leetify-notifier/config"
	"github.com/lucasberti/leetify-notifier/leetify"
	"github.com/lucasberti/leetify-notifier/notifiers"

	"github.com/rs/zerolog/log"
)

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
		if highlight.Id == "" || slices.Contains(cfg.KnownHighlightIds, highlight.Id) {
			continue
		}

		cfg.KnownHighlightIds = append(cfg.KnownHighlightIds, highlight.Id)
		cfg.SaveConfig(CONFIG_PATH)

		notifiers.SendTelegramMessage(cfg, "NEW HIGHLIGHT: "+highlight.Description)
	}

}
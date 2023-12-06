package main

import (
	"bytes"
	"slices"
	"strings"
	"sync"

	"github.com/lucasberti/leetify-notifier/config"
	"github.com/lucasberti/leetify-notifier/leetify"
	"github.com/lucasberti/leetify-notifier/notifiers"

	"github.com/rs/zerolog/log"
)

func replaceHighlightValues(input string, player string, description string, link string) string {
	input = strings.Replace(input, "%PLAYER%", player, -1)
	input = strings.Replace(input, "%DESCRIPTION%", description, -1)
	input = strings.Replace(input, "%HIGHLIGHTLINK%", link, -1)

	return input
}

func generateHighlightMessage(cfg *config.Config, highlight *leetify.Highlight) string {
	var message bytes.Buffer

	playerName := ""
	if mention, ok := cfg.TelegramUserNames[highlight.Steam64Id]; ok {
		playerName = mention
	}

	message.WriteString(replaceHighlightValues(cfg.HighlightMsg, playerName, highlight.Description, highlight.GetVideoURL()))

	return message.String()
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
		if highlight.Id == "" || slices.Contains(cfg.KnownHighlightIds, highlight.Id) {
			continue
		}

		cfg.KnownHighlightIds = append(cfg.KnownHighlightIds, highlight.Id)
		cfg.SaveConfig(CONFIG_PATH)

		notifiers.SendTelegramMessage(cfg, generateHighlightMessage(cfg, &highlight))
	}

}
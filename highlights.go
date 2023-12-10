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

func replaceHighlightValues(input, player, description, link string, skipLink bool) string {
	input = strings.Replace(input, "%PLAYER%", player, -1)
	input = strings.Replace(input, "%DESCRIPTION%", description, -1)

	if !skipLink {
		input = strings.Replace(input, "%HIGHLIGHTLINK%", link, -1)
	} else {
		input = strings.Replace(input, "%HIGHLIGHTLINK%", "", -1)
	}

	input = strings.Trim(input, "\n")

	return input
}

func generateHighlightMessage(cfg *config.Config, highlight *leetify.Highlight, skipLink bool) string {
	var message bytes.Buffer

	playerName := highlight.Username
	if mention, ok := cfg.TelegramUserNames[highlight.Steam64Id]; ok {
		playerName = mention
	}

	message.WriteString(replaceHighlightValues(cfg.HighlightMsg, playerName, highlight.Description, highlight.GetVideoURL(), skipLink))

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
		if friendProfile != nil {
			highlights = append(highlights, friendProfile.Highlights...)
		}
	}

	for _, highlight := range highlights {
		if highlight.Id == "" || slices.Contains(cfg.KnownHighlightIds, highlight.Id) {
			continue
		}

		log.Print("Downloading highlight: " + highlight.GetVideoURL())

		fileSize := highlight.GetVideoSize()

		if fileSize == 0 {
			log.Error().Msg("Could not get highlight file size")
			continue
		}

		if fileSize <= notifiers.MAX_FILE_SIZE {
			go func () {
				videoBody, err := highlight.DownloadHighlight()

				if err != nil {
					log.Error().Err(err).Msg("Could not download highlight")
					return
				}
	
				notifiers.SendTelegramVideo(cfg, videoBody, fileSize, generateHighlightMessage(cfg, &highlight, true))
			}()
		} else {
			go notifiers.SendTelegramMessage(cfg, generateHighlightMessage(cfg, &highlight, false))
		}	

		cfg.KnownHighlightIds = append(cfg.KnownHighlightIds, highlight.Id)
		cfg.SaveConfig(CONFIG_PATH)
	}

}
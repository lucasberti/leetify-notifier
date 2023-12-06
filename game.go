package main

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/lucasberti/leetify-notifier/config"
	"github.com/lucasberti/leetify-notifier/leetify"
	"github.com/lucasberti/leetify-notifier/notifiers"

	"github.com/rs/zerolog/log"
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
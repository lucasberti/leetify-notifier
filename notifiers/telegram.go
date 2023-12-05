package notifiers

import (
	"bytes"
	"encoding/json"
	"leetify_notifier/config"
	"net/http"

	"github.com/rs/zerolog/log"
)

func SendTelegramMessage(config *config.Config, message string) {
	log.Print("Sending Telegram message: " + message)

	url := "https://api.telegram.org/bot" + config.TelegramKey + "/sendMessage"

	data := map[string]any{
		"chat_id":    config.TelegramChatId,
		"text":       message,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Error().Err(err).Msg("Could not marshal data")
		return
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Error().Err(err).Msg("Could not send Telegram message")
	}
}
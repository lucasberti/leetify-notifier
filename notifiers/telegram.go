package notifiers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/lucasberti/leetify-notifier/config"

	"github.com/rs/zerolog/log"
)

func SendTelegramMessage(cfg *config.Config, message string) {
	log.Print("Sending Telegram message: " + message)

	url := "https://api.telegram.org/bot" + cfg.TelegramKey + "/sendMessage"

	data := map[string]interface{}{
		"chat_id":    cfg.TelegramChatId,
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
package notifiers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lucasberti/leetify-notifier/config"
	"github.com/technoweenie/multipartstreamer"

	"github.com/rs/zerolog/log"
)

func SendTelegramMessage(cfg *config.Config, message string) error {
	log.Print("Sending Telegram message: " + message)

	url := "https://api.telegram.org/bot" + cfg.TelegramKey + "/sendMessage"

	data := map[string]interface{}{
		"chat_id":    cfg.TelegramChatId,
		"text":       message,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Error().Err(err).Msg("Could not marshal data")
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Error().Err(err).Msg("Could not send Telegram message")
		return err
	}

	return nil
}

func SendTelegramVideo(cfg *config.Config, video []byte, caption string) error {
	log.Print("Sending Telegram video")

	url := "https://api.telegram.org/bot" + cfg.TelegramKey + "/sendVideo"

	ms := multipartstreamer.New()

	data := map[string]string{
		"chat_id":    strconv.FormatInt(cfg.TelegramChatId, 10),
		"caption":    caption,
	}

	ms.WriteFields(data)
	
	buf := bytes.NewBuffer(video)
	ms.WriteReader("video", "highlight.mp4", int64(len(video)), buf)

	log.Print(data)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	ms.SetupRequest(req)

	_, err = http.DefaultClient.Do(req)

	if err != nil {
		log.Error().Err(err).Msg("Could not send Telegram video")
		return err
	}

	return nil
}
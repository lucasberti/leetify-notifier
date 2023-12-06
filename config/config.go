package config

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
)

type Config struct {
	KnownMatchIds		[]string `json:"knownMatchIds"`
	KnownHighlightIds	[]string `json:"knownHighlightIds"`
	MainProfile			string   `json:"mainProfile"`
	TelegramKey 		string   `json:"telegramKey"`
	TelegramChatId 		int64    `json:"telegramChatId"`
}

func (config *Config) SaveConfig(path string) error {
	log.Print("Saving config...")

	jsonFile, err := json.MarshalIndent(config, "", "    ")

	if err != nil {
		return err
	}

	err = os.WriteFile(path, jsonFile, 0644)

	if err != nil {
		return err
	}

	log.Print("Saved config successfully: " + path)

	return nil
}

func LoadConfig(path string) (*Config, error) {
	log.Print("Loading config...")

	jsonFile, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config

	if err := json.Unmarshal(jsonFile, &config); err != nil {
		return nil, err
	}

	log.Print("Loaded config successfully: " + path)

	return &config, nil
}
package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	KnownMatchIds		[]string `json:"knownMatchIds"`
	KnownHighlightIds	[]string `json:"knownHighlightIds"`
	MainProfile			string   `json:"mainProfile"`
	TelegramKey 		string   `json:"telegramKey"`
	TelegramChatId 		int64    `json:"telegramChatId"`
}

func (config *Config) SaveConfig(path string) error {
	jsonFile, err := json.MarshalIndent(config, "", "    ")

	if err != nil {
		return err
	}

	err = os.WriteFile(path, jsonFile, 0644)

	if err != nil {
		return err
	}

	return nil
}

func LoadConfig(path string) (*Config, error) {
	jsonFile, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config

	json.Unmarshal(jsonFile, &config)

	return &config, nil
}
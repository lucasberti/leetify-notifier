package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lucasberti/leetify-notifier/config"
	"github.com/lucasberti/leetify-notifier/leetify"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CONFIG_PATH = "config.json"
	UPDATE_INTERVAL = 30 * time.Second
)


func run(ctx context.Context, cfg *config.Config) {
	ticker := time.NewTicker(UPDATE_INTERVAL)

	for {
		select {
		case <-ticker.C:
			log.Print("Checking for updates...")

			profile, err := leetify.GetProfile(cfg.MainProfile)

			if err != nil {
				log.Error().Err(err).Msg("Could not get main profile")
				continue
			}

			var wg sync.WaitGroup
			wg.Add(2)

			go checkGames(cfg, profile, &wg)
			go checkHighlights(cfg, profile, &wg)
		
			wg.Wait()

		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.LoadConfig(CONFIG_PATH)

	if err != nil {
		log.Error().Err(err).Msg("Could not load config")
		return
	}

	if cfg.MainProfile == "" {
		log.Error().Msg("Please set mainProfile in config.json")
		return
	}

	run(ctx, cfg)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-shutdown
	cancel()
}

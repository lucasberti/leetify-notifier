package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	config_path = "config.json"
)

func GetProfile(steam64Id string) (*Profile, error) {
	url := "https://api.leetify.com/api/profile/" + steam64Id

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var profileResponse Profile

	if err := json.NewDecoder(response.Body).Decode(&profileResponse); err != nil {
		return nil, err
	}

	log.Print("Got profile for " + steam64Id + ": " + profileResponse.Meta.Name)

	return &profileResponse, nil
}


func GetFriendsProfiles(friends map[string]string) []*Profile {
	c := make(chan *Profile)

	var wg sync.WaitGroup

	for _, friend := range friends {
		wg.Add(1)
		go func (friend string) {
			defer wg.Done()

			profile, err := GetProfile(friend)

			if err != nil {
				log.Error().Err(err).Msgf("Could not get friend profile: %s", friend)
			}

			c <- profile
		}(friend)
	}
	
	go func () {
		wg.Wait()
		close(c)
	}()

	var profiles []*Profile

	for profile := range c {
		profiles = append(profiles, profile)
	}

	return profiles
}


func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	config, err := LoadConfig(config_path)

	if err != nil {
		log.Error().Err(err).Msg("Could not load config")
		return
	}

	profile, err := GetProfile(config.MainProfile)

	if err != nil {
		log.Error().Err(err).Msg("Could not get main profile")
		return
	}
	
	allFriendsIds := profile.GetFriendsSteamIds()

	friendsProfiles := GetFriendsProfiles(allFriendsIds)

	log.Print(friendsProfiles)
}

package leetify

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
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
		go func(friend string) {
			defer wg.Done()

			profile, err := GetProfile(friend)

			if err != nil {
				log.Error().Err(err).Msgf("Could not get friend profile: %s", friend)
			}

			c <- profile
		}(friend)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	var profiles []*Profile

	for profile := range c {
		profiles = append(profiles, profile)
	}

	return profiles
}
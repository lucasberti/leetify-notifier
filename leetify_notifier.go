package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
				fmt.Println("Could not get friend profile: ", friend, err)
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
	mainProfile := "76561198040339223"

	profile, err := GetProfile(mainProfile)

	if err != nil {
		fmt.Println(err)
		return
	}
	
	allFriends := profile.GetFriendsSteamIds()

	friendsProfiles := GetFriendsProfiles(allFriends)

	fmt.Println(friendsProfiles)
}

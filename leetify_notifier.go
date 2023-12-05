package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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


func GetHighlights(steamid string, c chan map[string][]Highlight) {
	profile, err := GetProfile(steamid)

	if err != nil {
		c <- nil
	}

	idMap := make(map[string][]Highlight)

	if len(profile.Highlights) > 0 {
		idMap[profile.Meta.Name] = profile.Highlights
		
		c <- idMap
	} else {
		c <- nil
	}

	fmt.Println("Got highlights for " + profile.Meta.Name)
}

func main() {
	steam64 := "76561198040339223"

	profile, err := GetProfile(steam64)

	if err != nil {
		fmt.Println(err)
		return
	}
	
	allFriends := profile.GetFriendsSteamIds()

	c := make(chan map[string][]Highlight)
	defer close(c)

	for _, friend := range allFriends {
		go GetHighlights(friend, c)
	}

	for i := 0; i < len(allFriends); i++ {
		highlights := <-c

		if highlights != nil {
			fmt.Println(highlights)
		}
	}

	// fmt.Println(profile.GetHighlightsVideoURLs())
	// fmt.Println(allFriends)
}

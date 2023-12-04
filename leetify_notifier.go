package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetProfile(steam64Id string) Response {
	url := "https://api.leetify.com/api/profile/" + steam64Id

	response, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	var profileResponse Response

	if err := json.NewDecoder(response.Body).Decode(&profileResponse); err != nil {
		panic(err)
	}

	return profileResponse
}

func main() {
	steam64 := "76561198040339223"

	profile := GetProfile(steam64)
	steamIds := profile.ExtractSteamIdsFromFriends()

	fmt.Println(profile.getHighlightsVideoURLs())
	fmt.Println(steamIds)
}

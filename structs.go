package main

import "strings"

type Highlights struct {
	Description  string `json:"description"`
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnailUrl"`
}

type Rank struct {
	Type       string `json:"type"`
	SkillLevel uint16 `json:"skillLevel"`
}

type Teammates struct {
	ProfileUserLeetifyRating float32 `json:"profileUserLeetifyRating"`
	Rank                     Rank    `json:"rank"`
	SteamNickname            string  `json:"steamNickname"`
	Steam64Id                string  `json:"steam64Id"`
}

type Games struct {
	OwnTeamSteam64Ids []string `json:"ownTeamSteam64Ids"`
	GameId            string   `json:"gameId"`
	MapName           string   `json:"mapName"`
	MatchResult       string   `json:"matchResult"`
	Scores            []int    `json:"scores"`
}

type Response struct {
	Highlights []Highlights `json:"highlights"`
	Teammates  []Teammates  `json:"teammates"`
	Games      []Games      `json:"games"`
}

func (r Response) ExtractSteamIdsFromFriends() map[string]string {
	steamIds := make(map[string]string)

	for _, teammate := range r.Teammates {
		steamIds[teammate.SteamNickname] = teammate.Steam64Id
	}

	return steamIds
}

func (r Response) getHighlightsVideoURLs() []string {
	var videoURLs []string

	for _, highlight := range r.Highlights {
		original := highlight.ThumbnailUrl
		videoURL := strings.Replace(original, "/thumbs/", "/clips/", 1)
		final := strings.Replace(videoURL, "_thumb.jpg", ".mp4", 1)
		videoURLs = append(videoURLs, final)
	}
	
	return videoURLs
}
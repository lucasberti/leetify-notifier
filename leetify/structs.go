package leetify

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Highlight struct {
	Description  string `json:"description"`
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Steam64Id	 string `json:"steam64Id"`
}

type Rank struct {
	Type       string `json:"type"`
	SkillLevel uint16 `json:"skillLevel"`
}

type Teammate struct {
	ProfileUserLeetifyRating float32 `json:"profileUserLeetifyRating"`
	Rank                     Rank    `json:"rank"`
	SteamNickname            string  `json:"steamNickname"`
	Steam64Id                string  `json:"steam64Id"`
}

type Game struct {
	OwnTeamSteam64Ids []string `json:"ownTeamSteam64Ids"`
	GameId            string   `json:"gameId"`
	MapName           string   `json:"mapName"`
	MatchResult       string   `json:"matchResult"`
	Scores            []int    `json:"scores"`
	SkillLevel        uint16   `json:"skillLevel"`
	GameFinishedAt	  string   `json:"gameFinishedAt"`
}

type Meta struct {
	Name 		string `json:"name"`
	Steam64Id 	string `json:"steam64Id"`
}

type Profile struct {
	Highlights []Highlight `json:"highlights"`
	Teammates  []Teammate  `json:"teammates"`
	Games      []Game      `json:"games"`
	Meta       Meta        `json:"meta"`
}

func (g Game) GetGameLink() string {
	return "https://leetify.com/app/match-details/" + g.GameId
}

func (p Profile) GetLatestGame() *Game {
	return &p.Games[0]
}

func (p Profile) GetFriendsSteamIds() map[string]string {
	steamIds := make(map[string]string)

	for _, teammate := range p.Teammates {
		steamIds[teammate.SteamNickname] = teammate.Steam64Id
	}

	return steamIds
}

func (h Highlight) GetVideoURL() string {
	videoURL := strings.Replace(h.ThumbnailUrl, "/thumbs/", "/clips/", 1)
	videoURL = strings.Replace(videoURL, "_thumb.jpg", ".mp4", 1)

	return videoURL
}

func (h Highlight) GetVideoSize() uint32 {
	resp, err := http.Head(h.GetVideoURL())

	if err != nil {
		return 0
	}

	u64, err := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 64)

	if err != nil {
		return 0
	}

	u32 := uint32(u64)

	return u32
}

func (h Highlight) DownloadHighlight() ([]byte, error) {
	url := h.GetVideoURL()

	video, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer video.Body.Close()

	// Reading the whole file into memory isn't a good idea but 
	// it's not that big of a deal since there's not that many new highlights
	// at the same time
	videoData, _ := io.ReadAll(video.Body)

	return videoData, nil
}
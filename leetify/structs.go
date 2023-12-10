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
	Steam64Id    string `json:"steam64Id"`
	Username     string `json:"username"`
}

func (h *Highlight) GetVideoURL() string {
	videoURL := strings.Replace(h.ThumbnailUrl, "/thumbs/", "/clips/", 1)
	videoURL = strings.Replace(videoURL, "_thumb.jpg", ".mp4", 1)

	return videoURL
}

func (h *Highlight) GetVideoSize() int64 {
	resp, err := http.Head(h.GetVideoURL())

	if err != nil {
		return 0
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	if err != nil {
		return 0
	}

	return int64(size)
}

func (h *Highlight) DownloadHighlight() (io.ReadCloser, error) {
	url := h.GetVideoURL()

	video, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	return video.Body, nil
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
	GameFinishedAt    string   `json:"gameFinishedAt"`
}

func (g *Game) GetGameLink() string {
	return "https://leetify.com/app/match-details/" + g.GameId
}

type Meta struct {
	Name      string `json:"name"`
	Steam64Id string `json:"steam64Id"`
}

type Profile struct {
	Highlights []Highlight `json:"highlights"`
	Teammates  []Teammate  `json:"teammates"`
	Games      []Game      `json:"games"`
	Meta       Meta        `json:"meta"`
}

func (p *Profile) GetLatestGame() *Game {
	return &p.Games[0]
}

func (p *Profile) GetFriendsSteamIds() map[string]string {
	steamIds := make(map[string]string)

	for _, teammate := range p.Teammates {
		steamIds[teammate.SteamNickname] = teammate.Steam64Id
	}

	return steamIds
}

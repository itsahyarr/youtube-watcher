package internal

import (
	"net/url"
	"strings"
)

func IsValidYouTubeURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(u.Host)

	switch {
	case host == "www.youtube.com" || host == "youtube.com":
		return u.Path == "/watch" && u.Query().Get("v") != ""
	case host == "youtu.be":
		return u.Path != "" && u.Path != "/"
	default:
		return false
	}
}

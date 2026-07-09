package internal

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
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

func ValidateProxy(rawURL string) error {
	if rawURL == "" {
		return nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "http", "https", "socks5":
	default:
		return fmt.Errorf("unsupported proxy scheme %q: must be http, https, or socks5", scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("proxy URL has no host")
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		port = "80"
		if scheme == "socks5" {
			port = "1080"
		}
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("proxy host %q: DNS resolve failed: %w", host, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("proxy host %q: no addresses resolved", host)
	}

	target := net.JoinHostPort(addrs[0], port)
	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		return fmt.Errorf("proxy %q: TCP dial failed: %w", target, err)
	}
	conn.Close()

	return nil
}

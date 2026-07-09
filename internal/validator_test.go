package internal

import "testing"

func TestIsValidYouTubeURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://www.youtube.com/watch?v=abc123", true},
		{"https://youtube.com/watch?v=abc123", true},
		{"https://youtu.be/abc123", true},
		{"https://www.google.com", false},
		{"", false},
		{"not-a-url", false},
		{"https://www.youtube.com/", false},
	}
	for _, tt := range tests {
		got := IsValidYouTubeURL(tt.url)
		if got != tt.want {
			t.Errorf("IsValidYouTubeURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

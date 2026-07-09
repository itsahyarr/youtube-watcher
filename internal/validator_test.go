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

func TestValidateProxy(t *testing.T) {
	tests := []struct {
		name    string
		proxy   string
		wantErr bool
	}{
		{"empty", "", false},
		{"invalid scheme", "ftp://proxy:8080", true},
		{"no host", "http://", true},
		{"garbage", "not-a-url!!!", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProxy(tt.proxy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProxy(%q) error=%v, wantErr=%v", tt.proxy, err, tt.wantErr)
			}
		})
	}
}

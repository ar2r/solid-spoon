package downloader

import (
	"testing"
)

func TestParseQualityNum(t *testing.T) {
	tests := []struct {
		quality  string
		expected int
	}{
		{"360p", 360},
		{"480p", 480},
		{"720p", 720},
		{"1080p", 1080},
		{"1440p", 1440},
		{"2160p", 2160},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			result := parseQualityNum(tt.quality)
			if result != tt.expected {
				t.Errorf("parseQualityNum(%q) = %d, want %d", tt.quality, result, tt.expected)
			}
		})
	}
}

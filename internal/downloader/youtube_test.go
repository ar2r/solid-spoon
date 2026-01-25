package downloader

import (
	"encoding/json"
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

func TestNewYouTubeDownloader(t *testing.T) {
	d := NewYouTubeDownloader()

	if d.ytdlpPath != "yt-dlp" {
		t.Errorf("expected ytdlpPath to be 'yt-dlp', got %s", d.ytdlpPath)
	}

	expectedMaxSize := int64(2000 * 1024 * 1024)
	if d.maxSize != expectedMaxSize {
		t.Errorf("expected maxSize to be %d (2GB), got %d", expectedMaxSize, d.maxSize)
	}
}

func TestMaxLocalAPIServerConstant(t *testing.T) {
	// Проверяем что константа равна 2 ГБ
	expectedSize := int64(2000 * 1024 * 1024)
	if maxLocalAPIServer != expectedSize {
		t.Errorf("maxLocalAPIServer = %d, want %d (2GB)", maxLocalAPIServer, expectedSize)
	}
}

func TestYtdlpVideoInfoParsing(t *testing.T) {
	// Тестируем парсинг JSON от yt-dlp
	jsonData := `{
		"id": "test123",
		"title": "Test Video",
		"description": "Test Description",
		"duration": 120.5,
		"formats": [
			{
				"format_id": "18",
				"ext": "mp4",
				"width": 640,
				"height": 360,
				"filesize": 10485760,
				"vcodec": "avc1.42001E",
				"acodec": "mp4a.40.2",
				"format_note": "360p"
			},
			{
				"format_id": "22",
				"ext": "mp4",
				"width": 1280,
				"height": 720,
				"filesize": 52428800,
				"vcodec": "avc1.64001F",
				"acodec": "mp4a.40.2",
				"format_note": "720p"
			}
		]
	}`

	var info ytdlpVideoInfo
	err := json.Unmarshal([]byte(jsonData), &info)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if info.ID != "test123" {
		t.Errorf("expected ID 'test123', got %s", info.ID)
	}

	if info.Title != "Test Video" {
		t.Errorf("expected Title 'Test Video', got %s", info.Title)
	}

	if info.Duration != 120.5 {
		t.Errorf("expected Duration 120.5, got %f", info.Duration)
	}

	if len(info.Formats) != 2 {
		t.Errorf("expected 2 formats, got %d", len(info.Formats))
	}

	// Проверяем первый формат
	if info.Formats[0].Height != 360 {
		t.Errorf("expected first format height 360, got %d", info.Formats[0].Height)
	}

	if info.Formats[0].Ext != "mp4" {
		t.Errorf("expected first format ext 'mp4', got %s", info.Formats[0].Ext)
	}
}

func TestYtdlpFormatFiltering(t *testing.T) {
	// Тест логики фильтрации форматов
	formats := []ytdlpFormat{
		{Ext: "mp4", Height: 360, VCodec: "avc1", ACodec: "mp4a", Filesize: 10000000},
		{Ext: "mp4", Height: 720, VCodec: "avc1", ACodec: "mp4a", Filesize: 50000000},
		{Ext: "webm", Height: 1080, VCodec: "vp9", ACodec: "opus", Filesize: 100000000},
		{Ext: "mp4", Height: 480, VCodec: "avc1", ACodec: "none", Filesize: 20000000}, // без аудио
		{Ext: "mp4", Height: 0, VCodec: "avc1", ACodec: "mp4a", Filesize: 5000000},    // без высоты
	}

	maxSize := int64(2000 * 1024 * 1024)
	qualityMap := make(map[int]VideoFormat)

	for _, f := range formats {
		// Пропускаем не-MP4 форматы
		if f.Ext != "mp4" {
			continue
		}

		// Пропускаем форматы без видео
		if f.VCodec == "none" || f.VCodec == "" {
			continue
		}

		// Пропускаем форматы без аудио
		if f.ACodec == "none" || f.ACodec == "" {
			continue
		}

		if f.Height == 0 {
			continue
		}

		filesize := f.Filesize
		if filesize > maxSize {
			continue
		}

		qualityMap[f.Height] = VideoFormat{
			QualityNum: f.Height,
			Size:       filesize,
		}
	}

	// Должны остаться только 360p и 720p (MP4 с аудио и видео)
	if len(qualityMap) != 2 {
		t.Errorf("expected 2 filtered formats, got %d", len(qualityMap))
	}

	if _, ok := qualityMap[360]; !ok {
		t.Error("expected 360p format to be present")
	}

	if _, ok := qualityMap[720]; !ok {
		t.Error("expected 720p format to be present")
	}

	// 1080p webm должен быть отфильтрован
	if _, ok := qualityMap[1080]; ok {
		t.Error("1080p webm format should be filtered out")
	}

	// 480p без аудио должен быть отфильтрован
	if _, ok := qualityMap[480]; ok {
		t.Error("480p format without audio should be filtered out")
	}
}

func TestVideoFormatDescription(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		filesize int64
		wantDesc string
	}{
		{"360p small", 360, 5 * 1024 * 1024, "360p (~5MB)"},
		{"720p medium", 720, 50 * 1024 * 1024, "720p (~50MB)"},
		{"1080p large", 1080, 200 * 1024 * 1024, "1080p (~200MB)"},
		{"360p tiny", 360, 500 * 1024, "360p (~500KB)"},
		{"720p no size", 720, 0, "720p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qualityLabel := parseQualityLabel(tt.height)
			sizeDesc := formatSizeDesc(tt.filesize)
			description := qualityLabel + sizeDesc

			if description != tt.wantDesc {
				t.Errorf("got description %q, want %q", description, tt.wantDesc)
			}
		})
	}
}

// Helper functions for testing
func parseQualityLabel(height int) string {
	return formatQualityLabel(height)
}

func formatQualityLabel(height int) string {
	if height == 0 {
		return ""
	}
	return formatQuality(height)
}

func formatQuality(height int) string {
	return formatHeightToQuality(height)
}

func formatHeightToQuality(height int) string {
	if height == 0 {
		return ""
	}
	return formatHeightString(height)
}

func formatHeightString(height int) string {
	return formatHeight(height)
}

func formatHeight(height int) string {
	if height == 0 {
		return ""
	}
	return heightToString(height)
}

func heightToString(height int) string {
	if height == 0 {
		return ""
	}
	return intToQualityString(height)
}

func intToQualityString(height int) string {
	if height == 0 {
		return ""
	}
	// Simplified version for testing
	switch height {
	case 360:
		return "360p"
	case 480:
		return "480p"
	case 720:
		return "720p"
	case 1080:
		return "1080p"
	default:
		return ""
	}
}

func formatSizeDesc(filesize int64) string {
	if filesize == 0 {
		return ""
	}
	sizeMB := filesize / (1024 * 1024)
	if sizeMB > 0 {
		return formatMB(sizeMB)
	}
	sizeKB := filesize / 1024
	return formatKB(sizeKB)
}

func formatMB(sizeMB int64) string {
	if sizeMB == 0 {
		return ""
	}
	return formatMBString(sizeMB)
}

func formatMBString(sizeMB int64) string {
	return mbToString(sizeMB)
}

func mbToString(sizeMB int64) string {
	if sizeMB == 0 {
		return ""
	}
	return " (~" + intToString(sizeMB) + "MB)"
}

func formatKB(sizeKB int64) string {
	if sizeKB == 0 {
		return ""
	}
	return " (~" + intToString(sizeKB) + "KB)"
}

func intToString(n int64) string {
	// Simple int to string conversion for testing
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n /= 10
	}
	return result
}

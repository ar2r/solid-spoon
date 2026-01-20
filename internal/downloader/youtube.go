package downloader

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type Quality string

const (
	QualityLow    Quality = "360p"
	QualityMedium Quality = "480p"
	QualityHigh   Quality = "720p"
	QualityFull   Quality = "1080p"
)

type VideoFormat struct {
	Quality     Quality
	QualityNum  int
	Size        int64
	Description string
}

type YouTubeDownloader struct {
	client youtube.Client
}

func NewYouTubeDownloader() *YouTubeDownloader {
	return &YouTubeDownloader{
		client: youtube.Client{},
	}
}

func (d *YouTubeDownloader) GetAvailableFormats(videoID string) ([]VideoFormat, error) {
	video, err := d.client.GetVideo(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// Сначала пробуем форматы с аудио
	formats := video.Formats.WithAudioChannels()

	// Если нет форматов с аудио, берём все видео форматы
	if len(formats) == 0 {
		formats = video.Formats
	}

	if len(formats) == 0 {
		return nil, fmt.Errorf("no formats found")
	}

	qualityMap := make(map[string]VideoFormat)
	for _, f := range formats {
		if !strings.Contains(f.MimeType, "video/mp4") {
			continue
		}

		quality := f.QualityLabel
		if quality == "" {
			continue
		}

		qualityNum := parseQualityNum(quality)

		// Формируем описание размера
		var sizeDesc string
		if f.ContentLength > 0 {
			sizeMB := f.ContentLength / (1024 * 1024)
			if sizeMB > 0 {
				sizeDesc = fmt.Sprintf(" (~%dMB)", sizeMB)
			} else {
				sizeKB := f.ContentLength / 1024
				sizeDesc = fmt.Sprintf(" (~%dKB)", sizeKB)
			}
		}

		// Проверяем наличие аудио
		hasAudio := f.AudioChannels > 0
		audioDesc := ""
		if !hasAudio {
			audioDesc = " 🔇"
		}

		description := fmt.Sprintf("%s%s%s", quality, sizeDesc, audioDesc)

		// Предпочитаем форматы с аудио
		if existing, ok := qualityMap[quality]; ok {
			existingHasAudio := !strings.Contains(existing.Description, "🔇")
			if existingHasAudio && !hasAudio {
				continue // Пропускаем формат без аудио, если есть с аудио
			}
		}

		qualityMap[quality] = VideoFormat{
			Quality:     Quality(quality),
			QualityNum:  qualityNum,
			Size:        f.ContentLength,
			Description: description,
		}
	}

	result := make([]VideoFormat, 0, len(qualityMap))
	for _, vf := range qualityMap {
		result = append(result, vf)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].QualityNum < result[j].QualityNum
	})

	return result, nil
}

func (d *YouTubeDownloader) Download(videoID string) (string, error) {
	return d.DownloadWithQuality(videoID, "")
}

func (d *YouTubeDownloader) DownloadWithQuality(videoID string, quality Quality) (string, error) {
	video, err := d.client.GetVideo(videoID)
	if err != nil {
		return "", fmt.Errorf("failed to get video info: %w", err)
	}

	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return "", fmt.Errorf("no formats with audio found")
	}

	var selectedFormat *youtube.Format
	for i := range formats {
		if !strings.Contains(formats[i].MimeType, "video/mp4") {
			continue
		}

		if quality != "" && formats[i].QualityLabel == string(quality) {
			selectedFormat = &formats[i]
			break
		}

		if quality == "" {
			if selectedFormat == nil || formats[i].ContentLength < selectedFormat.ContentLength {
				selectedFormat = &formats[i]
			}
		}
	}

	if selectedFormat == nil {
		for i := range formats {
			if strings.Contains(formats[i].MimeType, "video/mp4") {
				selectedFormat = &formats[i]
				break
			}
		}
	}

	if selectedFormat == nil {
		selectedFormat = &formats[0]
	}

	stream, _, err := d.client.GetStream(video, selectedFormat)
	if err != nil {
		return "", fmt.Errorf("failed to get stream: %w", err)
	}
	defer stream.Close()

	tmpFile, err := os.CreateTemp("", "yt-*.mp4")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, stream)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	return tmpFile.Name(), nil
}

func parseQualityNum(quality string) int {
	var num int
	fmt.Sscanf(quality, "%dp", &num)
	return num
}

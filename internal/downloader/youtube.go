package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

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
	Width       int
	Height      int
}

type YouTubeDownloader struct {
	client youtube.Client
}

func NewYouTubeDownloader() *YouTubeDownloader {
	// Создаём HTTP-клиент с увеличенными таймаутами для больших файлов
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   2 * time.Minute, // 2 минуты на TLS handshake
			ResponseHeaderTimeout: 2 * time.Minute, // 2 минуты на получение заголовков
			IdleConnTimeout:       5 * time.Minute, // 5 минут на простой соединения
		},
		Timeout: 60 * time.Minute, // 60 минут на скачивание всего файла (до 2 ГБ)
	}

	return &YouTubeDownloader{
		client: youtube.Client{
			HTTPClient: httpClient,
		},
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

	const maxTelegramBotAPI = 50 * 1024 * 1024 // 50 МБ - реальный лимит Bot API

	qualityMap := make(map[string]VideoFormat)
	for _, f := range formats {
		if !strings.Contains(f.MimeType, "video/mp4") {
			continue
		}

		quality := f.QualityLabel
		if quality == "" {
			continue
		}

		// Пропускаем файлы больше 50 МБ (реальный лимит Bot API)
		if f.ContentLength > maxTelegramBotAPI {
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
			Width:       f.Width,
			Height:      f.Height,
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

type VideoInfo struct {
	FilePath    string
	Width       int
	Height      int
	Duration    int
	Title       string
	Description string
	Compressed  bool // Был ли файл сжат
}

func (d *YouTubeDownloader) DownloadWithQuality(videoID string, quality Quality) (string, error) {
	info, err := d.DownloadWithQualityInfo(videoID, quality)
	if err != nil {
		return "", err
	}
	return info.FilePath, nil
}

func (d *YouTubeDownloader) DownloadWithQualityInfo(videoID string, quality Quality) (*VideoInfo, error) {
	video, err := d.client.GetVideo(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return nil, fmt.Errorf("no formats with audio found")
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

	// Проверяем фактический размер выбранного формата
	const maxTelegramBotAPI = 50 * 1024 * 1024 // 50 МБ - реальный лимит Bot API
	if selectedFormat.ContentLength > maxTelegramBotAPI {
		sizeMB := float64(selectedFormat.ContentLength) / (1024 * 1024)
		return nil, fmt.Errorf("видео слишком большое (%.1f МБ), максимум 50 МБ", sizeMB)
	}

	stream, _, err := d.client.GetStream(video, selectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}
	defer stream.Close()

	tmpFile, err := os.CreateTemp("", "yt-*.mp4")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, stream)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	duration := int(video.Duration.Seconds())

	videoInfo := &VideoInfo{
		FilePath:    tmpFile.Name(),
		Width:       selectedFormat.Width,
		Height:      selectedFormat.Height,
		Duration:    duration,
		Title:       video.Title,
		Description: video.Description,
		Compressed:  false,
	}

	// Проверяем размер файла и сжимаем если нужно
	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	const maxSize = 50 * 1024 * 1024 // 50 МБ
	if fileInfo.Size() > maxSize {
		log.Printf("[YOUTUBE] File size %.2f MB exceeds limit, compressing...", float64(fileInfo.Size())/(1024*1024))
		compressedPath, err := compressVideo(tmpFile.Name(), maxSize)
		if err != nil {
			return nil, fmt.Errorf("failed to compress video: %w", err)
		}
		// Удаляем оригинальный файл
		os.Remove(tmpFile.Name())
		videoInfo.FilePath = compressedPath
		videoInfo.Compressed = true

		// Логируем результат сжатия
		compressedInfo, _ := os.Stat(compressedPath)
		log.Printf("[YOUTUBE] Compression complete: %.2f MB -> %.2f MB",
			float64(fileInfo.Size())/(1024*1024),
			float64(compressedInfo.Size())/(1024*1024))
	}

	return videoInfo, nil
}

// compressVideo сжимает видео до указанного размера с помощью ffmpeg
func compressVideo(inputPath string, targetSize int64) (string, error) {
	// Создаём временный файл для сжатого видео
	outputFile, err := os.CreateTemp("", "yt-compressed-*.mp4")
	if err != nil {
		return "", err
	}
	outputPath := outputFile.Name()
	outputFile.Close()

	// Получаем длительность видео
	durationCmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputPath)
	durationOut, err := durationCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get video duration: %w", err)
	}

	var durationSec float64
	fmt.Sscanf(string(durationOut), "%f", &durationSec)
	if durationSec == 0 {
		durationSec = 1
	}

	// Вычисляем целевой bitrate (оставляем запас 10%)
	targetSizeKb := float64(targetSize) * 0.9 / 1024
	targetBitrate := int((targetSizeKb * 8) / durationSec) // kbps

	// Минимальный bitrate для приемлемого качества
	if targetBitrate < 200 {
		targetBitrate = 200
	}

	// Сжимаем видео с помощью ffmpeg
	// -preset fast - быстрое кодирование
	// -b:v - битрейт видео
	// -maxrate и -bufsize для контроля размера
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-b:v", fmt.Sprintf("%dk", targetBitrate),
		"-maxrate", fmt.Sprintf("%dk", targetBitrate),
		"-bufsize", fmt.Sprintf("%dk", targetBitrate*2),
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("ffmpeg compression failed: %w", err)
	}

	// Проверяем размер результата
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		os.Remove(outputPath)
		return "", err
	}

	// Если всё ещё больше целевого размера, пробуем более агрессивное сжатие
	if compressedInfo.Size() > targetSize {
		newBitrate := int(float64(targetBitrate) * 0.7)
		if newBitrate < 150 {
			newBitrate = 150
		}

		cmd = exec.Command("ffmpeg",
			"-i", inputPath,
			"-c:v", "libx264",
			"-preset", "faster",
			"-b:v", fmt.Sprintf("%dk", newBitrate),
			"-maxrate", fmt.Sprintf("%dk", newBitrate),
			"-bufsize", fmt.Sprintf("%dk", newBitrate*2),
			"-c:a", "aac",
			"-b:a", "96k",
			"-movflags", "+faststart",
			"-y",
			outputPath,
		)

		if err := cmd.Run(); err != nil {
			os.Remove(outputPath)
			return "", fmt.Errorf("ffmpeg second pass failed: %w", err)
		}
	}

	return outputPath, nil
}

func parseQualityNum(quality string) int {
	var num int
	fmt.Sscanf(quality, "%dp", &num)
	return num
}

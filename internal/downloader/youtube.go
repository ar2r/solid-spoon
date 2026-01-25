package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

type Quality string

const (
	QualityLow    Quality = "360p"
	QualityMedium Quality = "480p"
	QualityHigh   Quality = "720p"
	QualityFull   Quality = "1080p"
)

// maxLocalAPIServer - максимальный размер файла для Local API Server (2 ГБ)
const maxLocalAPIServer = 2000 * 1024 * 1024

type VideoFormat struct {
	Quality     Quality
	QualityNum  int
	Size        int64
	Description string
	Width       int
	Height      int
}

type VideoInfo struct {
	FilePath    string
	Width       int
	Height      int
	Duration    int
	Title       string
	Description string
	Compressed  bool
}

// ytdlpVideoInfo represents the JSON output from yt-dlp -j
type ytdlpVideoInfo struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Duration    float64        `json:"duration"`
	Formats     []ytdlpFormat  `json:"formats"`
}

type ytdlpFormat struct {
	FormatID   string  `json:"format_id"`
	Ext        string  `json:"ext"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Filesize   int64   `json:"filesize"`
	FilesizeApprox int64 `json:"filesize_approx"`
	VCodec     string  `json:"vcodec"`
	ACodec     string  `json:"acodec"`
	FormatNote string  `json:"format_note"`
}

type YouTubeDownloader struct {
	ytdlpPath string
	maxSize   int64
}

func NewYouTubeDownloader() *YouTubeDownloader {
	return &YouTubeDownloader{
		ytdlpPath: "yt-dlp",
		maxSize:   maxLocalAPIServer,
	}
}

func (d *YouTubeDownloader) GetAvailableFormats(videoID string) ([]VideoFormat, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	cmd := exec.Command(d.ytdlpPath, "-j", url)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run yt-dlp: %w", err)
	}

	var info ytdlpVideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	qualityMap := make(map[int]VideoFormat)
	for _, f := range info.Formats {
		// Пропускаем не-MP4 форматы
		if f.Ext != "mp4" {
			continue
		}

		// Пропускаем форматы без видео
		if f.VCodec == "none" || f.VCodec == "" {
			continue
		}

		// Пропускаем форматы без аудио (предпочитаем с аудио)
		if f.ACodec == "none" || f.ACodec == "" {
			continue
		}

		if f.Height == 0 {
			continue
		}

		// Получаем размер файла
		filesize := f.Filesize
		if filesize == 0 {
			filesize = f.FilesizeApprox
		}

		// Пропускаем файлы больше лимита
		if filesize > d.maxSize {
			continue
		}

		qualityLabel := fmt.Sprintf("%dp", f.Height)

		// Формируем описание размера
		var sizeDesc string
		if filesize > 0 {
			sizeMB := filesize / (1024 * 1024)
			if sizeMB > 0 {
				sizeDesc = fmt.Sprintf(" (~%dMB)", sizeMB)
			} else {
				sizeKB := filesize / 1024
				sizeDesc = fmt.Sprintf(" (~%dKB)", sizeKB)
			}
		}

		description := fmt.Sprintf("%s%s", qualityLabel, sizeDesc)

		// Сохраняем только один формат для каждого качества (предпочитаем меньший размер)
		if existing, ok := qualityMap[f.Height]; ok {
			if filesize > 0 && filesize < existing.Size {
				qualityMap[f.Height] = VideoFormat{
					Quality:     Quality(qualityLabel),
					QualityNum:  f.Height,
					Size:        filesize,
					Description: description,
					Width:       f.Width,
					Height:      f.Height,
				}
			}
		} else {
			qualityMap[f.Height] = VideoFormat{
				Quality:     Quality(qualityLabel),
				QualityNum:  f.Height,
				Size:        filesize,
				Description: description,
				Width:       f.Width,
				Height:      f.Height,
			}
		}
	}

	// Если нет форматов с аудио, попробуем форматы которые yt-dlp может объединить
	if len(qualityMap) == 0 {
		for _, f := range info.Formats {
			if f.Ext != "mp4" && f.Ext != "webm" {
				continue
			}
			if f.VCodec == "none" || f.VCodec == "" {
				continue
			}
			if f.Height == 0 {
				continue
			}

			filesize := f.Filesize
			if filesize == 0 {
				filesize = f.FilesizeApprox
			}
			if filesize > d.maxSize {
				continue
			}

			qualityLabel := fmt.Sprintf("%dp", f.Height)
			var sizeDesc string
			if filesize > 0 {
				sizeMB := filesize / (1024 * 1024)
				if sizeMB > 0 {
					sizeDesc = fmt.Sprintf(" (~%dMB)", sizeMB)
				}
			}

			if _, ok := qualityMap[f.Height]; !ok {
				qualityMap[f.Height] = VideoFormat{
					Quality:     Quality(qualityLabel),
					QualityNum:  f.Height,
					Size:        filesize,
					Description: fmt.Sprintf("%s%s", qualityLabel, sizeDesc),
					Width:       f.Width,
					Height:      f.Height,
				}
			}
		}
	}

	if len(qualityMap) == 0 {
		return nil, fmt.Errorf("no suitable formats found")
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
	info, err := d.DownloadWithQualityInfo(videoID, quality)
	if err != nil {
		return "", err
	}
	return info.FilePath, nil
}

func (d *YouTubeDownloader) DownloadWithQualityInfo(videoID string, quality Quality) (*VideoInfo, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	// Создаём временный файл
	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("yt-%s.mp4", videoID))

	// Формируем аргументы yt-dlp
	// Используем только форматы с уже объединённым аудио (без ffmpeg)
	args := []string{
		"--no-playlist",
		"-o", outputPath,
	}

	// Добавляем фильтр по качеству - только форматы с видео И аудио (без merge)
	if quality != "" {
		height := parseQualityNum(string(quality))
		if height > 0 {
			// Выбираем лучший формат с указанным качеством, где есть и видео и аудио
			formatSpec := fmt.Sprintf("best[height<=%d][ext=mp4][acodec!=none][vcodec!=none]/best[height<=%d][acodec!=none][vcodec!=none]/best[ext=mp4][acodec!=none][vcodec!=none]", height, height)
			args = append(args, "-f", formatSpec)
		}
	} else {
		// Без указания качества - берём наименьший размер с аудио и видео
		args = append(args, "-f", "worst[ext=mp4][acodec!=none][vcodec!=none]/worst[acodec!=none][vcodec!=none]")
	}

	// Добавляем вывод JSON для получения метаданных
	args = append(args, "--print-json", url)

	cmd := exec.Command(d.ytdlpPath, args...)
	output, err := cmd.Output()
	if err != nil {
		// Удаляем частично скачанный файл
		os.Remove(outputPath)
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp download error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	// Парсим JSON вывод для получения метаданных
	var info ytdlpVideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		// Если не удалось распарсить, проверяем что файл скачался
		if _, statErr := os.Stat(outputPath); statErr != nil {
			return nil, fmt.Errorf("download failed: file not found")
		}
	}

	// Проверяем размер файла
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat downloaded file: %w", err)
	}

	if fileInfo.Size() > d.maxSize {
		os.Remove(outputPath)
		sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
		maxMB := float64(d.maxSize) / (1024 * 1024)
		return nil, fmt.Errorf("видео слишком большое (%.1f МБ), максимум %.0f МБ", sizeMB, maxMB)
	}

	// Получаем размеры видео из формата (если доступны)
	width, height := 0, 0
	for _, f := range info.Formats {
		if f.Width > width {
			width = f.Width
			height = f.Height
		}
	}

	return &VideoInfo{
		FilePath:    outputPath,
		Width:       width,
		Height:      height,
		Duration:    int(info.Duration),
		Title:       info.Title,
		Description: info.Description,
		Compressed:  false,
	}, nil
}

func parseQualityNum(quality string) int {
	var num int
	fmt.Sscanf(quality, "%dp", &num)
	return num
}

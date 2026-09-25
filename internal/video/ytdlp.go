package video

import (
	"context"
	"fmt"
	"time"

	"github.com/derethil/mise/internal/config/section"
	"github.com/lrstanley/go-ytdlp"
)

type YtdlpOption func(*ytdlp.Command)

type Progress struct {
	Status    string
	Total     int64
	Completed int64
}

type ProgressFunc func(Progress)

func DownloadVideo(ctx context.Context, url string, opts []YtdlpOption, extraArgs ...string) (*WorkDir, error) {
	workdir, err := NewWorkDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	dl := ytdlp.New().
		NoPlaylist().
		Color("no_color").
		Paths(workdir.Path).
		Output("video.%(ext)s").
		MergeOutputFormat("mp4")

	for _, opt := range opts {
		opt(dl)
	}

	result, err := dl.Run(ctx, append([]string{url}, extraArgs...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	_, err = result.GetExtractedInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get extracted info: %w", err)
	}

	return workdir, nil
}

func WithProgress(onProgress ProgressFunc) YtdlpOption {
	return func(c *ytdlp.Command) {
		c.ProgressFunc(200*time.Millisecond, func(u ytdlp.ProgressUpdate) {
			onProgress(Progress{
				Status:    string(u.Status),
				Total:     int64(u.TotalBytes),
				Completed: int64(u.DownloadedBytes),
			})
		})
	}
}

func WithConfig(cfg section.VideoConfig) YtdlpOption {
	return func(c *ytdlp.Command) {
		if cfg.Format != "" {
			c.Format(cfg.Format)
		}

		if cfg.CookiesFromBrowser != "" {
			c.CookiesFromBrowser(cfg.CookiesFromBrowser)
		}

		if cfg.CookiesFile != "" {
			c.Cookies(cfg.CookiesFile)
		}

		if cfg.Impersonate != "" {
			c.Impersonate(cfg.Impersonate)
		}

		if cfg.YtdlpPath != "" {
			c.SetExecutable(cfg.YtdlpPath)
		}

		if cfg.MaxDurationMinutes > 0 {
			c.MatchFilters(fmt.Sprintf("duration <= %d", cfg.MaxDurationMinutes*60))
		}
	}
}

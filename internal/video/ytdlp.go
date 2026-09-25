package video

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/derethil/mise/internal/config/section"
	"github.com/lrstanley/go-ytdlp"
)

const (
	progressInterval = 200 * time.Millisecond
	stderrTailLines  = 10
)

type Progress struct {
	Status    string
	Total     int64
	Completed int64
}

type ProgressFunc func(Progress)

func resolveBinary(ctx context.Context, cfg section.VideoConfig) (string, error) {
	if cfg.YtdlpPath != "" {
		path, err := exec.LookPath(cfg.YtdlpPath)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrBinaryMissing, err)
		}

		return path, nil
	}

	resolved, err := ytdlp.Install(ctx, &ytdlp.InstallOptions{AllowVersionMismatch: true})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrBinaryMissing, err)
	}

	return resolved.Executable, nil
}

func newCommand(workdir *WorkDir, cfg section.VideoConfig, executable string, onProgress ProgressFunc) *ytdlp.Command {
	cmd := ytdlp.New().
		NoPlaylist().
		FlatPlaylist().
		Color("no_color").
		Output("video.%(ext)s").
		MergeOutputFormat("mp4").
		Paths("home:"+workdir.Join(mediaDir)).
		Paths("temp:"+workdir.Join(tempDir)).
		PrintToFile(filepathTemplate, workdir.Join(filepathFilename)).
		SetExecutable(executable)

	applyConfig(cmd, cfg)
	applyProgress(cmd, onProgress)

	return cmd
}

func applyConfig(cmd *ytdlp.Command, cfg section.VideoConfig) {
	settings := []struct {
		value string
		apply func(string) *ytdlp.Command
	}{
		{cfg.Format, cmd.Format},
		{cfg.CookiesFile, cmd.Cookies},
		{cfg.CookiesFromBrowser, cmd.CookiesFromBrowser},
		{cfg.Impersonate, cmd.Impersonate},
	}

	for _, setting := range settings {
		if setting.value != "" {
			setting.apply(setting.value)
		}
	}
}

func applyProgress(cmd *ytdlp.Command, onProgress ProgressFunc) {
	if onProgress == nil {
		return
	}

	cmd.ProgressFunc(progressInterval, func(u ytdlp.ProgressUpdate) {
		onProgress(Progress{
			Status:    string(u.Status),
			Total:     int64(u.TotalBytes),
			Completed: int64(u.DownloadedBytes),
		})
	})
}

func classify(err error, res *ytdlp.Result) error {
	if _, ok := ytdlp.IsMisconfigError(err); ok {
		return fmt.Errorf("%w: %w", ErrBinaryMissing, err)
	}

	if _, ok := ytdlp.IsExitCodeError(err); ok {
		return &DownloadError{Stderr: stderrTail(res), err: err}
	}

	return &DownloadError{err: err}
}

func stderrTail(res *ytdlp.Result) string {
	if res == nil || res.Stderr == "" {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(res.Stderr), "\n")
	if len(lines) > stderrTailLines {
		lines = lines[len(lines)-stderrTailLines:]
	}

	return strings.Join(lines, "\n")
}

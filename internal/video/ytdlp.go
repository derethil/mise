package video

import (
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

func resolveDependencies(cfg section.ExtractConfig) (ytdlpPath, ffmpegPath string, err error) {
	resolve := func(cfgPath, name, configKey string) (string, error) {
		binary := name
		if cfgPath != "" {
			binary = cfgPath
		}

		path, err := exec.LookPath(binary)
		if err != nil {
			return "", &MissingBinaryError{
				Name:           name,
				ConfigKey:      configKey,
				ConfiguredPath: cfgPath,
				err:            err,
			}
		}

		return path, nil
	}

	ytdlpPath, err = resolve(cfg.YtdlpPath, "yt-dlp", "ytdlp_path")
	if err != nil {
		return "", "", err
	}

	ffmpegPath, err = resolve(cfg.FfmpegPath, "ffmpeg", "ffmpeg_path")
	if err != nil {
		return "", "", err
	}

	return ytdlpPath, ffmpegPath, nil
}

func newBaseCommand(workdir *WorkDir, cfg section.ExtractConfig, executable, ffmpeg string, onProgress ProgressFunc) *ytdlp.Command {
	cmd := ytdlp.New().
		NoPlaylist().
		FlatPlaylist().
		Color("no_color").
		Paths("home:"+workdir.Join(mediaDir)).
		Paths("temp:"+workdir.Join(tempDir)).
		PrintToFile(filepathTemplate, workdir.Join(filepathFilename)).
		SetExecutable(executable).
		FFmpegLocation(ffmpeg)

	applySourceConfig(cmd, cfg)
	applyProgress(cmd, onProgress)

	return cmd
}

func videoCommand(cmd *ytdlp.Command, format string) *ytdlp.Command {
	video := cmd.Clone().
		Output("video.%(ext)s").
		MergeOutputFormat("mp4")

	if format != "" {
		video.Format(format)
	}

	return video
}

func applySourceConfig(cmd *ytdlp.Command, cfg section.ExtractConfig) {
	settings := []struct {
		value string
		apply func(string) *ytdlp.Command
	}{
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

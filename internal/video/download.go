package video

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	metadataFilename = "video.info.json"
	filepathFilename = "video.filepath"
	filepathTemplate = "after_move:filepath"
	mediaDir         = "media"
	tempDir          = "tmp"
)

func (e *Extraction) Download(ctx context.Context) (*Media, error) {
	if err := e.workdir.WriteFile(metadataFilename, e.document); err != nil {
		return nil, fmt.Errorf("failed to write video metadata: %w", err)
	}

	if err := e.run(ctx); err != nil {
		return nil, err
	}

	path, err := locateDownload(e.workdir)
	if err != nil {
		return nil, err
	}

	return &Media{Source: e.Source, Path: path, WorkDir: e.workdir.Path}, nil
}

func (e *Extraction) run(ctx context.Context) error {
	cached := e.cmd.Clone().LoadInfoJSON(e.workdir.Join(metadataFilename))

	res, err := cached.Run(ctx, e.cfg.YtdlpArgs...)
	if err == nil {
		return nil
	}

	if ctx.Err() != nil {
		return classify(err, res)
	}

	slog.DebugContext(ctx, "retrying video download without cached metadata", slog.String("error", err.Error()))

	if retryErr := e.runFromURL(ctx); retryErr != nil {
		return errors.Join(classify(err, res), retryErr)
	}

	return nil
}

func (e *Extraction) runFromURL(ctx context.Context) error {
	args := append([]string{e.Source.URL}, e.cfg.YtdlpArgs...)

	res, err := e.cmd.Run(ctx, args...)
	if err != nil {
		return classify(err, res)
	}

	return nil
}

func locateDownload(workdir *WorkDir) (string, error) {
	reported, err := workdir.ReadFile(filepathFilename)
	if err != nil {
		return "", fmt.Errorf("%w: yt-dlp did not report an output path: %w", ErrDownloadFailed, err)
	}

	lines := strings.FieldsFunc(string(reported), func(r rune) bool { return r == '\n' || r == '\r' })
	if len(lines) == 0 {
		return "", fmt.Errorf("%w: yt-dlp reported an empty output path", ErrDownloadFailed)
	}

	path := strings.TrimSpace(lines[len(lines)-1])

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("%w: yt-dlp reported %s but it is not readable: %w", ErrDownloadFailed, path, err)
	}

	return path, nil
}

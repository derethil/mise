package video

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/derethil/mise/internal/config/section"
	"github.com/lrstanley/go-ytdlp"
)

type Extraction struct {
	Source Source

	cfg      section.VideoConfig
	cmd      *ytdlp.Command
	document []byte
	workdir  *WorkDir
}

func Extract(ctx context.Context, url string, cfg section.VideoConfig, onProgress ProgressFunc) (*Extraction, error) {
	executable, err := resolveBinary(ctx, cfg)
	if err != nil {
		return nil, err
	}

	workdir, err := newWorkDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	cmd := newCommand(workdir, cfg, executable, onProgress)

	infos, res, err := cmd.ExtractInfo(ctx, append([]string{url}, cfg.YtdlpArgs...)...)
	if err != nil {
		return nil, classify(err, res)
	}

	info, err := singleVideo(infos)
	if err != nil {
		return nil, err
	}

	if err := checkDuration(info, cfg.MaxDurationMinutes); err != nil {
		return nil, err
	}

	document, err := infoDocument(res, info)
	if err != nil {
		return nil, fmt.Errorf("failed to encode video metadata: %w", err)
	}

	return &Extraction{
		Source:   flattenSource(info, url),
		cfg:      cfg,
		cmd:      cmd,
		document: document,
		workdir:  workdir,
	}, nil
}

func (e *Extraction) WorkDir() string {
	return e.workdir.Path
}

func singleVideo(infos []*ytdlp.ExtractedInfo) (*ytdlp.ExtractedInfo, error) {
	switch {
	case len(infos) == 0:
		return nil, ErrNoVideo
	case len(infos) > 1, infos[0].IsPlaylist(), infos[0].Type == ytdlp.ExtractedTypeURL:
		return nil, ErrPlaylistURL
	default:
		return infos[0], nil
	}
}

func infoDocument(res *ytdlp.Result, info *ytdlp.ExtractedInfo) ([]byte, error) {
	for _, log := range res.OutputLogs {
		if log.JSON != nil {
			return *log.JSON, nil
		}
	}

	return json.MarshalIndent(info, "", "  ")
}

func checkDuration(info *ytdlp.ExtractedInfo, maxMinutes int) error {
	if deref(info.IsLive) {
		return ErrLiveVideo
	}

	limit := time.Duration(maxMinutes) * time.Minute
	if limit <= 0 {
		return nil
	}

	if info.Duration == nil {
		return ErrDurationUnknown
	}

	duration := time.Duration(*info.Duration * float64(time.Second))
	if duration > limit {
		return &TooLongError{Duration: duration, Limit: limit}
	}

	return nil
}

func flattenSource(info *ytdlp.ExtractedInfo, requestedURL string) Source {
	url := deref(info.WebpageURL)
	if url == "" {
		slog.Debug("webpage_url is empty, using requested URL instead", "requested_url", requestedURL)
		url = requestedURL
	}

	return Source{
		ID:          info.ID,
		URL:         url,
		Title:       deref(info.Title),
		Description: deref(info.Description),
		Uploader:    deref(info.Uploader),
		Duration:    time.Duration(deref(info.Duration) * float64(time.Second)),
		Thumbnail:   deref(info.Thumbnail),
		Extractor:   deref(info.Extractor),
	}
}

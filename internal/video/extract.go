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

	cfg      section.ExtractConfig
	baseCmd  *ytdlp.Command
	document []byte
	workdir  *WorkDir
}

func Probe(ctx context.Context, url string, cfg section.ExtractConfig, onProgress ProgressFunc) (*Extraction, error) {
	executable, ffmpeg, err := resolveDependencies(cfg)
	if err != nil {
		return nil, err
	}

	workdir, err := newWorkDir(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	baseCmd := newBaseCommand(workdir, cfg, executable, ffmpeg, onProgress)

	info, document, err := resolveInfo(ctx, url, cfg, baseCmd, workdir)
	if err != nil {
		return nil, err
	}

	if err := checkDuration(info, cfg.MaxDurationMinutes); err != nil {
		return nil, err
	}

	return &Extraction{
		Source:   flattenSource(info, url),
		cfg:      cfg,
		baseCmd:  baseCmd,
		document: document,
		workdir:  workdir,
	}, nil
}

func (e *Extraction) WorkDir() string {
	return e.workdir.Path
}

func (e *Extraction) Extract(ctx context.Context) (*Media, error) {
	videoPath, err := e.downloadVideo(ctx)
	if err != nil {
		return nil, err
	}

	audioPath, err := e.extractAudio(ctx)
	if err != nil {
		return nil, err
	}

	return &Media{Source: e.Source, VideoPath: videoPath, AudioPath: audioPath, WorkDir: e.workdir.Path}, nil
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

func resolveInfo(ctx context.Context, url string, cfg section.ExtractConfig, baseCmd *ytdlp.Command, workdir *WorkDir) (*ytdlp.ExtractedInfo, []byte, error) {
	info, document, err := loadCachedInfo(workdir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read cached video metadata: %w", err)
	}

	if info != nil {
		slog.InfoContext(ctx, "Using cached video, skipping fetch", "path", workdir.Join(metadataFilename))
		return info, document, nil
	}

	infos, res, err := baseCmd.ExtractInfo(ctx, append([]string{url}, cfg.YtdlpArgs...)...)
	if err != nil {
		return nil, nil, classify(err, res)
	}

	info, err = singleVideo(infos)
	if err != nil {
		return nil, nil, err
	}

	document, err = infoDocument(res, info)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode video metadata: %w", err)
	}

	return info, document, nil
}

func loadCachedInfo(workdir *WorkDir) (*ytdlp.ExtractedInfo, []byte, error) {
	if !workdir.Exists(metadataFilename) {
		return nil, nil, nil
	}

	document, err := workdir.ReadFile(metadataFilename)
	if err != nil {
		return nil, nil, err
	}

	info := new(ytdlp.ExtractedInfo)
	if err := json.Unmarshal(document, info); err != nil {
		return nil, nil, err
	}

	return info, document, nil
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

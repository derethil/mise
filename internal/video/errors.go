package video

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrBinaryMissing   = errors.New("yt-dlp executable not available")
	ErrDownloadFailed  = errors.New("video download failed")
	ErrNoVideo         = errors.New("no video found at url")
	ErrPlaylistURL     = errors.New("url resolves to a playlist")
	ErrVideoTooLong    = errors.New("video exceeds the configured maximum duration")
	ErrLiveVideo       = errors.New("url resolves to a live stream")
	ErrDurationUnknown = errors.New("video duration was not reported")
)

type TooLongError struct {
	Duration time.Duration
	Limit    time.Duration
}

func (e *TooLongError) Error() string {
	return fmt.Sprintf("%s: %s long, limit %s", ErrVideoTooLong, e.Duration, e.Limit)
}

func (e *TooLongError) Unwrap() error {
	return ErrVideoTooLong
}

type DownloadError struct {
	Stderr string
	err    error
}

func (e *DownloadError) Error() string {
	parts := []string{ErrDownloadFailed.Error()}

	if e.err != nil {
		parts = append(parts, e.err.Error())
	}

	if e.Stderr != "" {
		parts = append(parts, e.Stderr)
	}

	return strings.Join(parts, ": ")
}

func (e *DownloadError) Unwrap() []error {
	if e.err == nil {
		return []error{ErrDownloadFailed}
	}

	return []error{ErrDownloadFailed, e.err}
}

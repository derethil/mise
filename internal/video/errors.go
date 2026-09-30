package video

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrBinaryMissing   = errors.New("required executable not found")
	ErrDownloadFailed  = errors.New("video download failed")
	ErrNoVideo         = errors.New("no video found at url")
	ErrPlaylistURL     = errors.New("url resolves to a playlist")
	ErrVideoTooLong    = errors.New("video exceeds the configured maximum duration")
	ErrLiveVideo       = errors.New("url resolves to a live stream")
	ErrDurationUnknown = errors.New("video duration was not reported")
)

type MissingBinaryError struct {
	Name           string // e.g. "yt-dlp"
	ConfigKey      string // VideoConfig field's `key:` tag for overriding the path, e.g. "ytdlp_path"
	ConfiguredPath string // Empty when the default executable name was resolved through PATH.

	err error
}

func (e *MissingBinaryError) Error() string {
	if e.ConfiguredPath != "" {
		return fmt.Sprintf("%s: configured %s executable %q is unavailable: %s", ErrBinaryMissing, e.Name, e.ConfiguredPath, e.err)
	}

	return fmt.Sprintf("%s: %s not found, install it and ensure it's on PATH: %s", ErrBinaryMissing, e.Name, e.err)
}

func (e *MissingBinaryError) Unwrap() []error {
	return []error{ErrBinaryMissing, e.err}
}

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

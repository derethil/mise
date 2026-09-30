package cliutil

import (
	"testing"
	"time"

	"github.com/derethil/mise/internal/video"
	"github.com/stretchr/testify/assert"
)

func TestVideoUserErrorReferencesPublicDurationFlag(t *testing.T) {
	err := &video.TooLongError{Duration: 45 * time.Minute, Limit: 30 * time.Minute}

	message := UserMessage(VideoUserError(err))

	assert.Contains(t, message, "--max-duration")
	assert.NotContains(t, message, "video.max_duration_minutes")
}

func TestVideoUserErrorForConfiguredBinary(t *testing.T) {
	err := &video.MissingBinaryError{
		Name:           "ffmpeg",
		ConfigKey:      "ffmpeg_path",
		ConfiguredPath: "/missing/ffmpeg",
	}

	message := UserMessage(VideoUserError(err))

	assert.Contains(t, message, "/missing/ffmpeg")
	assert.Contains(t, message, "video.ffmpeg_path")
	assert.NotContains(t, message, "your PATH")
}

func TestVideoUserErrorForBinaryMissingFromPath(t *testing.T) {
	err := &video.MissingBinaryError{
		Name:      "ffmpeg",
		ConfigKey: "ffmpeg_path",
	}

	message := UserMessage(VideoUserError(err))

	assert.Contains(t, message, "ffmpeg")
	assert.Contains(t, message, "your PATH")
	assert.Contains(t, message, "video.ffmpeg_path")
}

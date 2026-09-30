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

func TestVideoUserErrorReferencesPublicYtdlpFlag(t *testing.T) {
	message := UserMessage(VideoUserError(video.ErrBinaryMissing))

	assert.Contains(t, message, "--yt-dlp-path")
	assert.NotContains(t, message, "video.ytdlp_path")
}

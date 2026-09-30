package video

import (
	"context"
	"slices"
	"testing"

	"github.com/derethil/mise/internal/config/section"
	"github.com/stretchr/testify/suite"
)

type AudioSuite struct {
	suite.Suite
}

func TestAudioSuite(t *testing.T) {
	suite.Run(t, new(AudioSuite))
}

func (s *AudioSuite) flagValue(args []string, flag string) (string, bool) {
	s.T().Helper()

	i := slices.Index(args, flag)
	if i < 0 || i+1 >= len(args) {
		return "", false
	}

	return args[i+1], true
}

func (s *AudioSuite) TestAudioCommandSetsExtractionFlags() {
	workdir := &WorkDir{Path: s.T().TempDir()}
	cmd := newCommand(workdir, section.VideoConfig{}, "/usr/bin/yt-dlp", nil)

	args := audioCommand(cmd, "mp3").BuildCommand(context.Background(), "https://tiktok.com/v").Args

	s.Contains(args, "--extract-audio")

	format, ok := s.flagValue(args, "--audio-format")
	s.Require().True(ok)
	s.Equal("mp3", format)

	output, ok := s.flagValue(args, "--output")
	s.Require().True(ok)
	s.Equal(audioOutputTemplate, output, "audio files must not collide with the video's own output name")
}

func (s *AudioSuite) TestAudioCommandDoesNotMutateTheOriginal() {
	workdir := &WorkDir{Path: s.T().TempDir()}
	cmd := newCommand(workdir, section.VideoConfig{}, "/usr/bin/yt-dlp", nil)

	_ = audioCommand(cmd, "mp3")

	args := cmd.BuildCommand(context.Background(), "https://tiktok.com/v").Args
	s.NotContains(args, "--extract-audio", "cloning must not leak audio flags back onto the video command")
}

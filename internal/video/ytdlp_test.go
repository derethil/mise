package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/derethil/mise/internal/config/section"
	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/suite"
)

type YtdlpSuite struct {
	suite.Suite
}

func TestYtdlpSuite(t *testing.T) {
	suite.Run(t, new(YtdlpSuite))
}

func (s *YtdlpSuite) buildArgs(cfg section.ExtractConfig) []string {
	s.T().Helper()

	workdir := &WorkDir{Path: s.T().TempDir()}
	base := newBaseCommand(workdir, cfg, "/usr/bin/yt-dlp", "/usr/bin/ffmpeg", nil)
	cmd := videoCommand(base, cfg.Format)

	return cmd.BuildCommand(context.Background(), "https://tiktok.com/v").Args
}

func (s *YtdlpSuite) flagValue(args []string, flag string) (string, bool) {
	s.T().Helper()

	i := slices.Index(args, flag)
	if i < 0 || i+1 >= len(args) {
		return "", false
	}

	return args[i+1], true
}

func (s *YtdlpSuite) TestDeterministicLayoutIsAlwaysSet() {
	args := s.buildArgs(section.ExtractConfig{})

	s.Contains(args, "--no-playlist")
	s.Contains(args, "--merge-output-format")

	output, ok := s.flagValue(args, "--output")
	s.Require().True(ok)
	s.Equal("video.%(ext)s", output, "a fixed output name keeps the workdir predictable")

	report, ok := s.flagValue(args, "--print-to-file")
	s.Require().True(ok)
	s.Equal(filepathTemplate, report)

	ffmpeg, ok := s.flagValue(args, "--ffmpeg-location")
	s.Require().True(ok, "yt-dlp must be pointed at the resolved ffmpeg, not left to find one itself")
	s.Equal("/usr/bin/ffmpeg", ffmpeg)
}

func (s *YtdlpSuite) TestMediaAndTempPathsAreSeparate() {
	args := s.buildArgs(section.ExtractConfig{})

	var paths []string
	for i, arg := range args {
		if arg == "--paths" && i+1 < len(args) {
			paths = append(paths, args[i+1])
		}
	}

	s.Require().Len(paths, 2)
	s.True(strings.HasPrefix(paths[0], "home:"))
	s.True(strings.HasSuffix(paths[0], "/"+mediaDir), "finished media lands in its own directory")
	s.True(strings.HasPrefix(paths[1], "temp:"))
	s.True(strings.HasSuffix(paths[1], "/"+tempDir), "fragments never touch the media directory")
}

func (s *YtdlpSuite) TestUnsetConfigAddsNoFlags() {
	args := s.buildArgs(section.ExtractConfig{})

	for _, flag := range []string{"--format", "--cookies", "--cookies-from-browser", "--impersonate"} {
		s.NotContains(args, flag, "%s must not appear when unconfigured", flag)
	}
}

func (s *YtdlpSuite) TestConfigIsApplied() {
	cfg := section.ExtractConfig{
		Format:             "best[ext=mp4]",
		CookiesFile:        "/tmp/cookies.txt",
		CookiesFromBrowser: "firefox",
		Impersonate:        "chrome",
	}

	args := s.buildArgs(cfg)

	for flag, want := range map[string]string{
		"--format":               cfg.Format,
		"--cookies":              cfg.CookiesFile,
		"--cookies-from-browser": cfg.CookiesFromBrowser,
		"--impersonate":          cfg.Impersonate,
	} {
		got, ok := s.flagValue(args, flag)
		s.Require().True(ok, "%s missing from argv", flag)
		s.Equal(want, got)
	}
}

func (s *YtdlpSuite) TestResolvedExecutableIsTheOneRun() {
	workdir := &WorkDir{Path: s.T().TempDir()}
	base := newBaseCommand(workdir, section.ExtractConfig{YtdlpPath: "yt-dlp"}, "/resolved/yt-dlp", "/usr/bin/ffmpeg", nil)
	cmd := videoCommand(base, "")

	args := cmd.BuildCommand(context.Background(), "https://tiktok.com/v").Args

	s.Equal("/resolved/yt-dlp", args[0], "the validated binary is the one executed")
}

func (s *YtdlpSuite) TestResolveDependenciesRejectsUnusableYtdlpPath() {
	for _, path := range []string{"definitely-not-on-path", "/nonexistent/yt-dlp"} {
		s.Run(path, func() {
			_, _, err := resolveDependencies(section.ExtractConfig{YtdlpPath: path})

			s.ErrorIs(err, ErrBinaryMissing)

			var missing *MissingBinaryError
			s.Require().ErrorAs(err, &missing)
			s.Equal("yt-dlp", missing.Name)
			s.Equal(path, missing.ConfiguredPath)
		})
	}
}

func (s *YtdlpSuite) TestResolveDependenciesRejectsUnusableFfmpegPath() {
	dir := s.T().TempDir()
	s.writeFakeBinary(dir, "yt-dlp")
	s.T().Setenv("PATH", dir)

	for _, path := range []string{"definitely-not-on-path", "/nonexistent/ffmpeg"} {
		s.Run(path, func() {
			_, _, err := resolveDependencies(section.ExtractConfig{
				FfmpegPath: path,
			})

			s.ErrorIs(err, ErrBinaryMissing)

			var missing *MissingBinaryError
			s.Require().ErrorAs(err, &missing)
			s.Equal("ffmpeg", missing.Name)
			s.Equal(path, missing.ConfiguredPath)
		})
	}
}

func (s *YtdlpSuite) TestResolveDependenciesUsesConfiguredPaths() {
	dir := s.T().TempDir()
	fakeYtdlp := s.writeFakeBinary(dir, "yt-dlp")
	fakeFfmpeg := s.writeFakeBinary(dir, "ffmpeg")
	s.T().Setenv("PATH", dir)

	ytdlpPath, ffmpegPath, err := resolveDependencies(section.ExtractConfig{
		YtdlpPath:  "yt-dlp",
		FfmpegPath: "ffmpeg",
	})

	s.Require().NoError(err)
	s.Equal(fakeYtdlp, ytdlpPath, "a bare name resolves through PATH to an absolute path")
	s.Equal(fakeFfmpeg, ffmpegPath, "a bare name resolves through PATH to an absolute path")
}

func (s *YtdlpSuite) TestResolveDependenciesFallsBackToBareNameWhenUnconfigured() {
	dir := s.T().TempDir()
	fakeYtdlp := s.writeFakeBinary(dir, "yt-dlp")
	fakeFfmpeg := s.writeFakeBinary(dir, "ffmpeg")
	s.T().Setenv("PATH", dir)

	ytdlpPath, ffmpegPath, err := resolveDependencies(section.ExtractConfig{})

	s.Require().NoError(err)
	s.Equal(fakeYtdlp, ytdlpPath, "an already-installed yt-dlp must be used, never downloaded")
	s.Equal(fakeFfmpeg, ffmpegPath, "an already-installed ffmpeg must be used, never downloaded")
}

// writeFakeBinary writes an empty executable stub named name into dir and
// returns its absolute path. resolveDependencies only ever stats/looks up
// binaries via PATH, it never executes them, so the stub's contents don't matter.
func (s *YtdlpSuite) writeFakeBinary(dir, name string) string {
	s.T().Helper()

	path := filepath.Join(dir, name)
	s.Require().NoError(os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755))

	return path
}

func (s *YtdlpSuite) TestStderrTailIsNilSafe() {
	s.Empty(stderrTail(nil))
	s.Empty(stderrTail(&ytdlp.Result{}))
}

func (s *YtdlpSuite) TestStderrTailKeepsTheEndOfStderr() {
	var lines []string
	for i := range 25 {
		lines = append(lines, string(rune('a'+i)))
	}

	tail := stderrTail(&ytdlp.Result{Stderr: strings.Join(lines, "\n")})

	s.Len(strings.Split(tail, "\n"), stderrTailLines)
	s.Contains(tail, lines[24], "yt-dlp puts the real error last")
	s.NotContains(tail, lines[0]+"\n")
}

func (s *YtdlpSuite) TestClassifyWrapsUnknownErrors() {
	err := classify(errors.New("boom"), nil)

	s.ErrorIs(err, ErrDownloadFailed, "callers match on the sentinel")
	s.ErrorContains(err, "boom", "the cause survives for debugging")
}

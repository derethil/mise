package video

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DownloadSuite struct {
	suite.Suite

	workdir *WorkDir
}

func TestDownloadSuite(t *testing.T) {
	suite.Run(t, new(DownloadSuite))
}

func (s *DownloadSuite) SetupTest() {
	s.workdir = &WorkDir{Path: s.T().TempDir()}
}

func (s *DownloadSuite) reportMedia(names ...string) string {
	s.T().Helper()

	media := s.workdir.Join(mediaDir)
	s.Require().NoError(os.MkdirAll(media, 0o755))

	var contents string
	for _, name := range names {
		path := filepath.Join(media, name)
		s.Require().NoError(os.WriteFile(path, []byte("x"), 0o644))
		contents += path + "\n"
	}

	s.Require().NoError(s.workdir.WriteFile(filepathFilename, []byte(contents)))

	return filepath.Join(media, names[len(names)-1])
}

func (s *DownloadSuite) TestReturnsReportedPath() {
	want := s.reportMedia("video.mp4")

	got, err := locateDownload(s.workdir)

	s.Require().NoError(err)
	s.Equal(want, got)
}

func (s *DownloadSuite) TestUsesLastLineWhenYtdlpAppends() {
	want := s.reportMedia("video.webm", "video.mp4")

	got, err := locateDownload(s.workdir)

	s.Require().NoError(err)
	s.Equal(want, got, "--print-to-file appends across the cached run and its retry")
}

func (s *DownloadSuite) TestIgnoresTrailingBlankLines() {
	want := s.reportMedia("video.mp4")
	s.Require().NoError(s.workdir.WriteFile(filepathFilename, []byte(want+"\n\n\n")))

	got, err := locateDownload(s.workdir)

	s.Require().NoError(err)
	s.Equal(want, got)
}

func (s *DownloadSuite) TestFailsWhenNothingWasReported() {
	_, err := locateDownload(s.workdir)

	s.ErrorIs(err, ErrDownloadFailed)
}

func (s *DownloadSuite) TestFailsWhenReportFileIsBlank() {
	s.Require().NoError(s.workdir.WriteFile(filepathFilename, []byte("\n \n")))

	_, err := locateDownload(s.workdir)

	s.ErrorIs(err, ErrDownloadFailed)
}

func (s *DownloadSuite) TestFailsWhenReportedFileIsMissing() {
	missing := s.workdir.Join(mediaDir, "video.mp4")
	s.Require().NoError(s.workdir.WriteFile(filepathFilename, []byte(missing+"\n")))

	_, err := locateDownload(s.workdir)

	s.ErrorIs(err, ErrDownloadFailed)
	s.ErrorContains(err, "not readable")
}

func (s *DownloadSuite) TestIgnoresSiblingArtifacts() {
	want := s.reportMedia("video.mp4")

	for _, sibling := range []string{"video.description", "video.webp", "video.info.json"} {
		s.Require().NoError(os.WriteFile(s.workdir.Join(mediaDir, sibling), []byte("x"), 0o644))
	}

	got, err := locateDownload(s.workdir)

	s.Require().NoError(err)
	s.Equal(want, got, "yt-dlp reports the media file, so siblings in the same dir are irrelevant")
}

func (s *DownloadSuite) TestDoesNotIdentifyMediaByExtension() {
	want := s.reportMedia("video.mkv")

	got, err := locateDownload(s.workdir)

	s.Require().NoError(err)
	s.Equal(want, got, "any container yt-dlp reports is accepted")
}

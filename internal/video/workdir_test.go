package video

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type WorkDirSuite struct {
	suite.Suite

	workdir *WorkDir
}

func TestWorkDirSuite(t *testing.T) {
	suite.Run(t, new(WorkDirSuite))
}

func (s *WorkDirSuite) SetupTest() {
	s.workdir = &WorkDir{Path: s.T().TempDir()}
}

func (s *WorkDirSuite) TestJoinStaysInsideTheWorkdir() {
	workdir := &WorkDir{Path: "/tmp/mise-import-abc"}

	s.Equal("/tmp/mise-import-abc/video.mp4", workdir.Join("video.mp4"))
	s.Equal("/tmp/mise-import-abc/media/video.mp4", workdir.Join("media", "video.mp4"))
}

func (s *WorkDirSuite) TestWriteAndReadRoundTrip() {
	s.Require().NoError(s.workdir.WriteFile("notes.txt", []byte("hello")))

	got, err := s.workdir.ReadFile("notes.txt")

	s.Require().NoError(err)
	s.Equal("hello", string(got))
}

func (s *WorkDirSuite) TestWriteJSONIsReadableOnDisk() {
	s.Require().NoError(s.workdir.WriteJSON("meta.json", map[string]string{"id": "abc"}))

	got, err := s.workdir.ReadFile("meta.json")

	s.Require().NoError(err)
	s.JSONEq(`{"id":"abc"}`, string(got))
}

func (s *WorkDirSuite) TestExists() {
	s.False(s.workdir.Exists("video.mp4"))

	s.Require().NoError(s.workdir.WriteFile("video.mp4", []byte("x")))

	s.True(s.workdir.Exists("video.mp4"))
}

func (s *WorkDirSuite) TestGlobIsRelativeToTheWorkdir() {
	s.Require().NoError(os.MkdirAll(s.workdir.Join("media"), 0o755))
	s.Require().NoError(s.workdir.WriteFile(filepath.Join("media", "video.mp4"), []byte("x")))
	s.Require().NoError(s.workdir.WriteFile("video.info.json", []byte("{}")))

	matches, err := s.workdir.Glob(filepath.Join("media", "*"))

	s.Require().NoError(err)
	s.Require().Len(matches, 1)
	s.Equal(s.workdir.Join("media", "video.mp4"), matches[0])
}

func (s *WorkDirSuite) TestNewWorkDirCreatesIsolatedDirectories() {
	first, err := newWorkDir()
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = first.Cleanup() })

	second, err := newWorkDir()
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = second.Cleanup() })

	s.NotEqual(first.Path, second.Path, "concurrent imports must not share a workdir")
	s.DirExists(first.Path)
}

func (s *WorkDirSuite) TestCleanupRemovesEverything() {
	workdir, err := newWorkDir()
	s.Require().NoError(err)
	s.Require().NoError(workdir.WriteFile("video.mp4", []byte("x")))

	s.Require().NoError(workdir.Cleanup())

	s.NoDirExists(workdir.Path)
}

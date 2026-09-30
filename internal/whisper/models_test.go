package whisper

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PullSuite struct {
	suite.Suite

	server *httptest.Server
	body   string
}

func (s *PullSuite) SetupTest() {
	s.body = "fake model bytes"

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(s.body))
	}))

	dataDir, err := os.MkdirTemp("", "mise-whisper-test")
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = os.RemoveAll(dataDir) })

	config.DataDir = dataDir
	baseURL = s.server.URL + "/"
}

func (s *PullSuite) TearDownTest() {
	s.server.Close()
}

func (s *PullSuite) TestPullDownloadsModel() {
	downloaded, err := Pull(s.T().Context(), "tiny", nil)
	s.Require().NoError(err)
	s.True(downloaded)

	data, err := os.ReadFile(ModelPath("tiny"))
	s.Require().NoError(err)
	s.Equal(s.body, string(data))
}

func (s *PullSuite) TestPullSkipsExistingModel() {
	require.NoError(s.T(), os.MkdirAll(ModelsDir(), 0o755))
	require.NoError(s.T(), os.WriteFile(ModelPath("tiny"), []byte("already here"), 0o644))

	downloaded, err := Pull(s.T().Context(), "tiny", nil)
	s.Require().NoError(err)
	s.False(downloaded)

	data, err := os.ReadFile(ModelPath("tiny"))
	s.Require().NoError(err)
	s.Equal("already here", string(data))
}

func (s *PullSuite) TestEnsureSkipsExistingModel() {
	require.NoError(s.T(), os.MkdirAll(ModelsDir(), 0o755))
	require.NoError(s.T(), os.WriteFile(ModelPath("tiny"), []byte("already here"), 0o644))

	err := Ensure(s.T().Context(), "tiny", nil)
	s.Require().NoError(err)

	data, err := os.ReadFile(ModelPath("tiny"))
	s.Require().NoError(err)
	s.Equal("already here", string(data))
}

func (s *PullSuite) TestPullRejectsInvalidModel() {
	_, err := Pull(s.T().Context(), "not-a-real-model", nil)
	s.ErrorIs(err, ErrInvalidModel)
}

func (s *PullSuite) TestPullReportsProgress() {
	var lastCompleted int64

	downloaded, err := Pull(s.T().Context(), "base.en", func(p PullProgress) error {
		lastCompleted = p.Completed
		return nil
	})
	s.Require().NoError(err)
	s.True(downloaded)
	s.Equal(int64(len(s.body)), lastCompleted)
}

func TestPullSuite(t *testing.T) {
	suite.Run(t, new(PullSuite))
}

func TestModelPath(t *testing.T) {
	config.DataDir = t.TempDir()

	assert.Equal(t, filepath.Join(config.DataDir, "whisper", "ggml-tiny.en.bin"), ModelPath("tiny.en"))
	assert.Equal(t, filepath.Join(config.DataDir, "whisper", "ggml-large-v3.bin"), ModelPath("large"))
	assert.Equal(t, filepath.Join(config.DataDir, "whisper", "ggml-large-v3-turbo.bin"), ModelPath("turbo"))
}

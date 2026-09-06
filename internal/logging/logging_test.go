package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/suite"
)

type LoggingSuite struct {
	suite.Suite
	origStateDir string
}

func (s *LoggingSuite) SetupTest() {
	s.origStateDir = config.StateDir
	config.StateDir = s.T().TempDir()
}

func (s *LoggingSuite) TearDownTest() {
	config.StateDir = s.origStateDir
}

func (s *LoggingSuite) logFilePath() string {
	return filepath.Join(config.StateDir, "mise.log")
}

func (s *LoggingSuite) readLogLines() []map[string]any {
	data, err := os.ReadFile(s.logFilePath())
	s.Require().NoError(err)

	var lines []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}

		var m map[string]any
		s.Require().NoError(json.Unmarshal([]byte(line), &m))
		lines = append(lines, m)
	}

	return lines
}

func (s *LoggingSuite) captureStdout(fn func()) string {
	orig := os.Stdout
	r, w, err := os.Pipe()
	s.Require().NoError(err)
	os.Stdout = w

	fn()

	s.Require().NoError(w.Close())
	os.Stdout = orig

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	s.Require().NoError(err)

	return buf.String()
}

func (s *LoggingSuite) TestInitCreatesStateDirAndLogFile() {
	s.Require().NoError(Init(false))

	info, err := os.Stat(s.logFilePath())
	s.Require().NoError(err)
	s.False(info.IsDir())
}

func (s *LoggingSuite) TestStdoutGetsMessageOnly() {
	stdout := s.captureStdout(func() {
		s.Require().NoError(Init(false))
		ctx := NewInvocation(context.Background(), "recipe clean", []string{"recipe", "clean", "5"})
		slog.InfoContext(ctx, "hello")
	})

	s.Contains(stdout, "hello\n")
	s.NotContains(stdout, "invocation_id")
	s.NotContains(stdout, "{")
}

func (s *LoggingSuite) TestFileGetsStructuredAttrs() {
	s.captureStdout(func() {
		s.Require().NoError(Init(false))
		ctx := NewInvocation(context.Background(), "recipe clean", []string{"recipe", "clean", "5"})
		slog.InfoContext(ctx, "hello")
	})

	lines := s.readLogLines()
	s.Require().Len(lines, 1)

	s.Equal("hello", lines[0]["msg"])
	s.Equal("recipe clean", lines[0]["command"])
	s.Equal([]any{"recipe", "clean", "5"}, lines[0]["args"])
	s.NotEmpty(lines[0]["invocation_id"])
	s.Contains(lines[0], "duration_ms")
}

func (s *LoggingSuite) TestLogWithoutInvocationOmitsInvocationAttrs() {
	s.captureStdout(func() {
		s.Require().NoError(Init(false))
		slog.Info("hello")
	})

	lines := s.readLogLines()
	s.Require().Len(lines, 1)

	s.NotContains(lines[0], "invocation_id")
	s.NotContains(lines[0], "command")
}

func (s *LoggingSuite) TestDebugAlwaysWrittenToFile() {
	s.captureStdout(func() {
		s.Require().NoError(Init(false))
		slog.Debug("should still be logged")
	})

	lines := s.readLogLines()
	s.Require().Len(lines, 1)
	s.Equal("should still be logged", lines[0]["msg"])
}

func (s *LoggingSuite) TestDebugSuppressedOnStdoutWithoutVerbose() {
	stdout := s.captureStdout(func() {
		s.Require().NoError(Init(false))
		slog.Debug("should not appear on stdout")
	})

	s.NotContains(stdout, "should not appear on stdout")
}

func (s *LoggingSuite) TestVerboseEnablesDebug() {
	s.captureStdout(func() {
		s.Require().NoError(Init(true))
		slog.Debug("debug line")
	})

	lines := s.readLogLines()
	s.Require().Len(lines, 1)
	s.Equal("debug line", lines[0]["msg"])
}

func TestLoggingSuite(t *testing.T) {
	suite.Run(t, new(LoggingSuite))
}

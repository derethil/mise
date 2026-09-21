//go:build integration

package tandoor_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const diagnosticLines = 50

func (s *TandoorSuite) AfterTest(suiteName, testName string) {
	if s.T().Failed() {
		s.logContainers()
	}
}

func (s *TandoorSuite) logContainers() {
	ctx, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	dir, err := os.MkdirTemp("", fmt.Sprintf("mise-tandoor-%s-*", s.version))
	if err != nil {
		s.T().Logf("full logs unavailable: %v", err)
	}

	for _, service := range []string{databaseContainer, tandoorContainer} {
		s.T().Logf("[%s tandoor %s]\n%s", service, s.version, s.containerLogs(ctx, dir, service))
	}
}

func (s *TandoorSuite) containerLogs(ctx context.Context, dir, service string) string {
	container, err := s.stack.ServiceContainer(ctx, service)
	if err != nil {
		return fmt.Sprintf("no container: %v", err)
	}

	reader, err := container.Logs(ctx)
	if err != nil {
		return fmt.Sprintf("no logs: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Sprintf("unreadable logs: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) <= diagnosticLines {
		return strings.Join(lines, "\n")
	}

	return fmt.Sprintf(
		"%s\n\n(last %d of %d lines. %s)\n\n",
		strings.Join(lines[len(lines)-diagnosticLines:], "\n"),
		diagnosticLines,
		len(lines),
		s.writeFullLog(dir, service, content),
	)
}

func (s *TandoorSuite) writeFullLog(dir, service string, content []byte) string {
	if dir == "" {
		return "full log unavailable"
	}

	path := filepath.Join(dir, service+".log")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Sprintf("full log unavailable: %v", err)
	}

	return fmt.Sprintf("for the full log output, see %s", path)
}

package video

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ErrorsSuite struct {
	suite.Suite
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsSuite))
}

func (s *ErrorsSuite) TestTooLongErrorMatchesItsSentinel() {
	err := error(&TooLongError{Duration: 30 * time.Minute, Limit: 20 * time.Minute})

	s.ErrorIs(err, ErrVideoTooLong)
	s.ErrorContains(err, "30m0s")
	s.ErrorContains(err, "20m0s")
}

func (s *ErrorsSuite) TestDownloadErrorMatchesSentinelAndCause() {
	cause := errors.New("exit code 1")
	err := error(&DownloadError{Stderr: "ERROR: Sign in to confirm", err: cause})

	s.ErrorIs(err, ErrDownloadFailed, "the sentinel drives the user-facing message")
	s.ErrorIs(err, cause, "the yt-dlp error stays reachable for debugging")
}

func (s *ErrorsSuite) TestDownloadErrorPrefersStderr() {
	err := &DownloadError{Stderr: "ERROR: Sign in to confirm", err: errors.New("exit code 1")}

	s.Contains(err.Error(), "Sign in to confirm", "yt-dlp's own words reach the user verbatim")
}

func (s *ErrorsSuite) TestDownloadErrorFallsBackToCause() {
	err := &DownloadError{err: errors.New("context canceled")}

	s.Contains(err.Error(), "context canceled")
}

func (s *ErrorsSuite) TestJoinedRetryErrorsKeepBothCauses() {
	first := &DownloadError{Stderr: "first failure", err: errors.New("exit code 1")}
	second := &DownloadError{Stderr: "second failure", err: errors.New("exit code 2")}

	joined := errors.Join(error(first), error(second))

	s.Require().ErrorIs(joined, ErrDownloadFailed)
	s.ErrorContains(joined, "first failure")
	s.ErrorContains(joined, "second failure")

	var found *DownloadError
	s.Require().ErrorAs(joined, &found)
	s.Equal("first failure", found.Stderr, "the earlier, usually more specific failure is surfaced")
}

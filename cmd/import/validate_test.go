package importcmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/stretchr/testify/suite"
)

type ValidateSuite struct {
	suite.Suite

	cookies string
}

func TestValidateSuite(t *testing.T) {
	suite.Run(t, new(ValidateSuite))
}

func (s *ValidateSuite) SetupTest() {
	s.cookies = filepath.Join(s.T().TempDir(), "cookies.txt")
	s.Require().NoError(os.WriteFile(s.cookies, []byte("# Netscape HTTP Cookie File\n"), 0o644))
}

// run drives the real command so flag parsing and ArgValidator ordering are
// exercised. Every case here must fail validation, which happens before the
// Before hook and Action, so nothing reaches the network.
func (s *ValidateSuite) run(args ...string) error {
	s.T().Helper()

	return Command.Run(context.Background(), append([]string{"import"}, args...))
}

func (s *ValidateSuite) TestRejectsMissingURL() {
	err := s.run()

	s.Require().Error(err)
	s.ErrorIs(err, cliutil.ErrIncorrectUsage)
}

func (s *ValidateSuite) TestRejectsMalformedURL() {
	err := s.run("not-a-url")

	s.Require().Error(err)
	s.Equal("not-a-url is not a valid url", cliutil.UserMessage(err))
}

func (s *ValidateSuite) TestRejectsUnsupportedSource() {
	err := s.run("https://vimeo.com/123")

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "is not a supported import source")
}

func (s *ValidateSuite) TestRejectsUnsupportedBrowser() {
	err := s.run("https://tiktok.com/@a/video/1", "--"+flagCookiesFromBrowser, "netscape")

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "is not a supported browser")
}

func (s *ValidateSuite) TestRejectsMissingCookiesFile() {
	err := s.run("https://tiktok.com/@a/video/1", "--"+flagCookiesFile, "/nonexistent/cookies.txt")

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "does not exist")
}

func (s *ValidateSuite) TestRejectsDirectoryAsCookiesFile() {
	err := s.run("https://tiktok.com/@a/video/1", "--"+flagCookiesFile, s.T().TempDir())

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "is a directory, not a file")
}

func (s *ValidateSuite) TestRejectsBothCookieSources() {
	err := s.run(
		"https://tiktok.com/@a/video/1",
		"--"+flagCookiesFile, s.cookies,
		"--"+flagCookiesFromBrowser, "firefox",
	)

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "cannot be used together")
}

func (s *ValidateSuite) TestURLIsValidatedBeforeCookies() {
	err := s.run("https://vimeo.com/123", "--"+flagCookiesFromBrowser, "netscape")

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "is not a supported import source",
		"the url is reported first, so the user fixes the real problem")
}

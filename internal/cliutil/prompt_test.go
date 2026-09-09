package cliutil

import (
	"os"
	"os/exec"
	"testing"

	"github.com/manifoldco/promptui"
	"github.com/stretchr/testify/suite"
)

type PromptSuite struct {
	suite.Suite
}

func TestPromptSuite(t *testing.T) {
	suite.Run(t, new(PromptSuite))
}

func (s *PromptSuite) TestConfirmErrorsWhenNotInteractive() {
	ok, err := Confirm("proceed?")

	s.False(ok)
	s.ErrorIs(err, promptui.ErrEOF)
}

func (s *PromptSuite) TestAutoConfirmAlwaysAccepts() {
	ok, err := AutoConfirm("proceed?")

	s.Require().NoError(err)
	s.True(ok)
}

func (s *PromptSuite) TestConfirmOrDieExitsWhenNotInteractive() {
	cmd := exec.Command(os.Args[0], "-test.run=TestConfirmOrDieHelperProcess")
	cmd.Env = append(os.Environ(), "GO_WANT_CONFIRM_OR_DIE_HELPER=1")

	err := cmd.Run()

	var exitErr *exec.ExitError
	s.Require().ErrorAs(err, &exitErr)
	s.Equal(1, exitErr.ExitCode())
}

func TestConfirmOrDieHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_CONFIRM_OR_DIE_HELPER") != "1" {
		t.Skip("not invoked as a helper process")
	}

	ConfirmOrDie("proceed?")
}

func (s *PromptSuite) TestSelectOptionErrorsWhenNotInteractive() {
	result, err := SelectOption("pick one", []string{"a", "b"})

	s.Empty(result)
	s.Error(err)
}

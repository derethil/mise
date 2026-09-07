package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/stretchr/testify/suite"
)

type ErrorsSuite struct {
	suite.Suite
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsSuite))
}

func (s *ErrorsSuite) TestTandoorUserError_NotFound() {
	err := tandoorUserError(fmt.Errorf("wrapped: %w", tandoor.ErrTandoorNotFound))

	s.Equal("Recipe not found in Tandoor. Double-check the recipe ID.", userMessage(err))
}

func (s *ErrorsSuite) TestTandoorUserError_Unauthorized() {
	err := tandoorUserError(fmt.Errorf("wrapped: %w", tandoor.ErrTandoorUnauthorized))

	s.Equal("Tandoor rejected the request as unauthorized. Check your Tandoor API token.", userMessage(err))
}

func (s *ErrorsSuite) TestTandoorUserError_RequestFailed() {
	err := tandoorUserError(fmt.Errorf("wrapped: %w", tandoor.ErrTandoorRequestFailed))

	s.Equal("Could not reach Tandoor. Check tandoor.base_url and that the server is running.", userMessage(err))
}

func (s *ErrorsSuite) TestTandoorUserError_UnknownErrorPassesThrough() {
	original := errors.New("boom")

	err := tandoorUserError(original)

	s.Same(original, err)
}

func (s *ErrorsSuite) TestTandoorUserError_Nil() {
	s.NoError(tandoorUserError(nil))
}

func (s *ErrorsSuite) TestUserMessage_PlainErrorUsesErrorString() {
	s.Equal("boom", userMessage(errors.New("boom")))
}

func (s *ErrorsSuite) TestUserMessage_UnwrapsToOriginalError() {
	err := errWithUserMessage(tandoor.ErrTandoorNotFound, "friendly message")

	s.True(errors.Is(err, tandoor.ErrTandoorNotFound))
	s.Equal("friendly message", userMessage(err))
}

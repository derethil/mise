package cliutil

import (
	"context"
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
	err := TandoorUserError(fmt.Errorf("wrapped: %w", tandoor.ErrTandoorNotFound))

	s.Equal("Recipe not found in Tandoor. Double-check the recipe ID.", UserMessage(err))
}

func (s *ErrorsSuite) TestTandoorUserError_Unauthorized() {
	err := TandoorUserError(fmt.Errorf("wrapped: %w", tandoor.ErrTandoorUnauthorized))

	s.Equal("Tandoor rejected the request as unauthorized. Check your Tandoor API token.", UserMessage(err))
}

func (s *ErrorsSuite) TestTandoorUserError_RequestFailed() {
	err := TandoorUserError(fmt.Errorf("wrapped: %w", tandoor.ErrTandoorRequestFailed))

	s.Equal("Could not reach Tandoor. Check tandoor.base_url and that the server is running.", UserMessage(err))
}

func (s *ErrorsSuite) TestTandoorUserError_UnknownErrorPassesThrough() {
	original := errors.New("boom")

	err := TandoorUserError(original)

	s.Same(original, err)
}

func (s *ErrorsSuite) TestTandoorUserError_Nil() {
	s.NoError(TandoorUserError(nil))
}

func (s *ErrorsSuite) TestUserMessage_PlainErrorUsesErrorString() {
	s.Equal("boom", UserMessage(errors.New("boom")))
}

func (s *ErrorsSuite) TestUserMessage_UnwrapsToOriginalError() {
	err := ErrWithUserMessage(tandoor.ErrTandoorNotFound, "friendly message")

	s.True(errors.Is(err, tandoor.ErrTandoorNotFound))
	s.Equal("friendly message", UserMessage(err))
}

func (s *ErrorsSuite) TestUserMessage_ContextCanceled() {
	s.Equal("Cancelled.", UserMessage(context.Canceled))
}

func (s *ErrorsSuite) TestUserMessage_WrappedContextCanceled() {
	err := fmt.Errorf("recipe 42: %w", context.Canceled)

	s.Equal("Cancelled.", UserMessage(err))
}

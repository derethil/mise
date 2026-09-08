// Package cliutil holds helpers shared across mise's CLI commands: error
// wrapping for user-facing messages, AI feature loading, and progress/confirm
// prompts.
package cliutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/core/status"
)

var ErrIncorrectUsage = errors.New("incorrect usage")

type userError interface {
	error
	UserMessage() string
}

func UserMessage(err error) string {
	if errors.Is(err, context.Canceled) {
		return "Cancelled."
	}

	if ue, ok := errors.AsType[userError](err); ok {
		return ue.UserMessage()
	}
	return err.Error()
}

type userFacingError struct {
	err     error
	message string
}

func (e *userFacingError) Error() string       { return e.err.Error() }
func (e *userFacingError) Unwrap() error       { return e.err }
func (e *userFacingError) UserMessage() string { return e.message }

func ErrWithUserMessage(err error, format string, args ...any) error {
	return &userFacingError{err: err, message: fmt.Sprintf(format, args...)}
}

func TandoorUserError(err error) error {
	switch {
	case errors.Is(err, tandoor.ErrTandoorNotFound):
		return ErrWithUserMessage(err, "Recipe not found in Tandoor. Double-check the recipe ID.")
	case errors.Is(err, tandoor.ErrTandoorUnauthorized):
		return ErrWithUserMessage(err, "Tandoor rejected the request as unauthorized. Check your Tandoor API token.")
	case errors.Is(err, tandoor.ErrTandoorRequestFailed):
		return ErrWithUserMessage(err, "Could not reach Tandoor. Check tandoor.base_url and that the server is running.")
	default:
		return err
	}
}

func AIUserError(err error, model string) error {
	switch {
	case errors.Is(err, ai.ErrModelMissingTools):
		return ErrWithUserMessage(err, "Model %s doesn't support tool calling, which recipe clean needs to look up existing foods and units. Choose a different model with --model or modify your config.", model)
	case errors.Is(err, config.ErrInvalidConfig):
		return ErrWithUserMessage(err, "The AI provider isn't configured correctly. Check your provider settings (e.g. providers.ollama.base_url) and try again.")
	case errors.Is(err, status.ErrNotFound):
		return ErrWithUserMessage(err, "Unable to load model %s. Please ensure it is available for use by your provider.", model)
	case errors.Is(err, ai.ErrMalformedResponse):
		return ErrWithUserMessage(err, "Model %s didn't return a properly formatted response. Try again, or use a different/more capable model with --model.", model)
	default:
		return err
	}
}

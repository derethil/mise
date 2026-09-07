package cmd

import (
	"errors"
	"fmt"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/core/status"
)

type userError interface {
	error
	UserMessage() string
}

func userMessage(err error) string {
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

func errWithUserMessage(err error, format string, args ...any) error {
	return &userFacingError{err: err, message: fmt.Sprintf(format, args...)}
}

func tandoorUserError(err error) error {
	switch {
	case errors.Is(err, tandoor.ErrTandoorNotFound):
		return errWithUserMessage(err, "Recipe not found in Tandoor. Double-check the recipe ID.")
	case errors.Is(err, tandoor.ErrTandoorUnauthorized):
		return errWithUserMessage(err, "Tandoor rejected the request as unauthorized. Check your Tandoor API token.")
	case errors.Is(err, tandoor.ErrTandoorRequestFailed):
		return errWithUserMessage(err, "Could not reach Tandoor. Check tandoor.base_url and that the server is running.")
	default:
		return err
	}
}

func aiUserError(err error, model string) error {
	switch {
	case errors.Is(err, ai.ErrModelMissingTools):
		return errWithUserMessage(err, "Model %s doesn't support tool calling, which recipe clean needs to look up existing foods and units. Choose a different model with --model or modify your config.", model)
	case errors.Is(err, config.ErrInvalidConfig):
		return errWithUserMessage(err, "The AI provider isn't configured correctly. Check your provider settings (e.g. providers.ollama.base_url) and try again.")
	case errors.Is(err, status.ErrNotFound):
		return errWithUserMessage(err, "Unable to load model %s. Please ensure it is available for use by your provider.", model)
	case errors.Is(err, ai.ErrMalformedResponse):
		return errWithUserMessage(err, "Model %s didn't return a properly formatted response. Try again, or use a different/more capable model with --model.", model)
	default:
		return err
	}
}

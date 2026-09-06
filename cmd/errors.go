package cmd

import (
	"errors"
	"fmt"
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

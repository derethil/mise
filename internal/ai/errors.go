package ai

import "errors"

var (
	ErrPromptNotFound       = errors.New("prompt not found")
	ErrModelMissingTools    = errors.New("model does not support tool calling")
	ErrNoModels             = errors.New("no models provided")
	ErrFeatureNotRegistered = errors.New("feature is not registered")
)

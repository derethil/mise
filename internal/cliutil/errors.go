// Package cliutil holds helpers shared across mise's CLI commands: error
// wrapping for user-facing messages, AI feature loading, and progress/confirm
// prompts.
package cliutil

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/derethil/mise/internal/video"
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

func WarnIfTandoorVersionUnsupported(ctx context.Context, client *tandoor.Client) {
	version, supported, err := client.VersionSupported(ctx)
	if err != nil {
		slog.DebugContext(ctx, "could not determine Tandoor version", "error", err)
		return
	}

	if !supported {
		slog.WarnContext(ctx, fmt.Sprintf(
			"Warning: Your Tandoor version (%s) is not officially supported. Supported versions are %s to %s. Some features may not work as expected.",
			version, tandoor.MinSupportedVersion, tandoor.MaxSupportedVersion,
		))
	}
}

func AIUserError(err error, model string) error {
	switch {
	case errors.Is(err, providers.ErrOllamaNotInstalled):
		return ErrWithUserMessage(err, "Could not find `ollama` on your PATH. Install Ollama or add it to your PATH, then try again.")
	case errors.Is(err, providers.ErrOllamaUnavailable):
		return ErrWithUserMessage(err, "Could not connect to Ollama. Make sure it is running (try `ollama serve`) or provide --providers.ollama.autostart and try again.")
	case errors.Is(err, ai.ErrModelMissingTools):
		return ErrWithUserMessage(err, "Model %s doesn't support tool calling, which mise's AI commands need to look up existing entries in Tandoor. Choose a different model with --model or modify your config.", model)
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

func VideoUserError(err error) error {
	if tooLong, ok := errors.AsType[*video.TooLongError](err); ok {
		return ErrWithUserMessage(err, "That video is %s long (limit %s). Raise `video.max_duration_minutes` to import it anyway.",
			tooLong.Duration.Round(time.Second), tooLong.Limit)
	}

	if download, ok := errors.AsType[*video.DownloadError](err); ok && download.Stderr != "" {
		return ErrWithUserMessage(err, "yt-dlp could not download this video:\n%s", download.Stderr)
	}

	switch {
	case errors.Is(err, video.ErrBinaryMissing):
		return ErrWithUserMessage(err, "Could not find `yt-dlp` on your PATH, and downloading one failed. Install it (`nix profile install nixpkgs#yt-dlp`) or set `video.ytdlp_path`.")
	case errors.Is(err, video.ErrPlaylistURL):
		return ErrWithUserMessage(err, "That URL is a playlist or profile. Pass a link to a single video.")
	case errors.Is(err, video.ErrLiveVideo):
		return ErrWithUserMessage(err, "That URL is a live stream, which mise can't import. Wait for the recording to be published and try again.")
	case errors.Is(err, video.ErrDurationUnknown):
		return ErrWithUserMessage(err, "yt-dlp didn't report a duration for that video, so mise can't check it against `video.max_duration_minutes`. Set that to 0 to import it anyway.")
	case errors.Is(err, video.ErrNoVideo):
		return ErrWithUserMessage(err, "No video was found at that URL.")
	case errors.Is(err, video.ErrDownloadFailed):
		return ErrWithUserMessage(err, "yt-dlp could not download this video: %s", err)
	default:
		return err
	}
}

// Package whisper resolves and downloads the ggml model files used by
// whisper.cpp for transcription.
package whisper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/config/section"
)

var baseURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/"

const ext = ".bin"

var ErrInvalidModel = errors.New("invalid whisper model")
var ErrPullDeclined = errors.New("model download declined")

type ConfirmFunc func(question string) (bool, error)

var modelAliases = map[string]string{
	"large": "large-v3",
	"turbo": "large-v3-turbo",
}

func ModelsDir() string {
	return filepath.Join(config.DataDir, "whisper")
}

func modelFilename(model string) string {
	if alias, ok := modelAliases[model]; ok {
		model = alias
	}

	return "ggml-" + model + ext
}

func ModelPath(model string) string {
	return filepath.Join(ModelsDir(), modelFilename(model))
}

func modelURL(model string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, modelFilename(model))

	return u.String(), nil
}

type PullProgress struct {
	Total, Completed int64
}

type PullProgressFunc func(PullProgress) error

func Pull(ctx context.Context, model string, onProgress PullProgressFunc) (bool, error) {
	if !section.IsValidWhisperModel(model) {
		return false, fmt.Errorf("%w: %s", ErrInvalidModel, model)
	}

	destination := ModelPath(model)
	if _, err := os.Stat(destination); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	if err := os.MkdirAll(ModelsDir(), 0o755); err != nil {
		return false, err
	}

	src, err := modelURL(model)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("download %s: %s", model, resp.Status)
	}

	tmp := destination + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return false, err
	}
	defer func() { _ = os.Remove(tmp) }()

	if err := copyWithProgress(out, resp.Body, resp.ContentLength, onProgress); err != nil {
		_ = out.Close()
		return false, err
	}

	if err := out.Close(); err != nil {
		return false, err
	}

	if err := os.Rename(tmp, destination); err != nil {
		return false, err
	}

	return true, nil
}

func Ensure(ctx context.Context, model string, confirm ConfirmFunc, onProgress PullProgressFunc) error {
	destination := ModelPath(model)
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	ok, err := confirm(fmt.Sprintf("Whisper model %q is not available locally. Download it now?", model))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrPullDeclined, model)
	}

	_, err = Pull(ctx, model, onProgress)
	return err
}

func copyWithProgress(dst io.Writer, src io.Reader, total int64, onProgress PullProgressFunc) error {
	buf := make([]byte, 64*1024)
	var completed int64

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return err
			}

			completed += int64(n)
			if onProgress != nil {
				if err := onProgress(PullProgress{Total: total, Completed: completed}); err != nil {
					return err
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
	}
}

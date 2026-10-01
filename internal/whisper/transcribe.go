package whisper

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	cwhisper "github.com/ggerganov/whisper.cpp/bindings/go"
	"github.com/go-audio/wav"
)

func Transcribe(ctx context.Context, modelCfg string, audioPath string, confirm ConfirmFunc, onProgress PullProgressFunc, opts ...Option) (string, error) {
	if err := Ensure(ctx, modelCfg, confirm, onProgress); err != nil {
		return "", err
	}

	model := cwhisper.Whisper_init(ModelPath(modelCfg))
	if model == nil {
		return "", fmt.Errorf("failed to load whisper model %q", modelCfg)
	}
	defer model.Whisper_free()

	params := model.Whisper_full_default_params(cwhisper.SAMPLING_BEAM_SEARCH)

	params.SetThreads(runtime.NumCPU())
	params.SetNoContext(true)
	params.SetPrintSpecial(false)
	params.SetPrintProgress(false)
	params.SetPrintRealtime(false)
	params.SetPrintTimestamps(false)

	for _, opt := range opts {
		if err := opt(model, &params); err != nil {
			return "", err
		}
	}

	return process(model, params, audioPath)
}

func loadAudioBuffer(audioPath string) ([]float32, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	decoder := wav.NewDecoder(file)

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	if decoder.SampleRate != cwhisper.SampleRate {
		return nil, fmt.Errorf("audio sample rate %d does not match expected sample rate %d", decoder.SampleRate, cwhisper.SampleRate)
	}

	if decoder.NumChans != 1 {
		return nil, fmt.Errorf("unsupported number of channels %d, expected 1", decoder.NumChans)
	}

	return buf.AsFloat32Buffer().Data, nil
}

func process(model *cwhisper.Context, params cwhisper.Params, audioPath string) (string, error) {
	data, err := loadAudioBuffer(audioPath)
	if err != nil {
		return "", err
	}

	// TODO: route through ProgressFunc once output/progress UX is refactored
	cb := func(progress int) {
		fmt.Printf("\rProgress: %d%%", min(progress, 100))
	}

	if err := model.Whisper_full(params, data, nil, nil, cb); err != nil {
		return "", fmt.Errorf("failed to process audio: %w", err)
	}

	return outputTranscript(model), nil
}

func outputTranscript(model *cwhisper.Context) string {
	builder := &strings.Builder{}

	for i := range model.Whisper_full_n_segments() {
		builder.WriteString(model.Whisper_full_get_segment_text(i))
	}

	return builder.String()
}

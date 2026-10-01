package importcmd

import (
	"context"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/config/section"
	"github.com/derethil/mise/internal/video"
	"github.com/derethil/mise/internal/whisper"
)

func transcribe(ctx context.Context, cfg config.Config, media video.Media) (string, error) {
	if cfg.Models.Whisper == "" {
		return "", cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage,
			"models.whisper is not set. Configure a model (%s) and try again.", section.WhisperModelTiers)
	}

	language, err := chooseLanguage(cfg.Transcription.DefaultLanguage)
	if err != nil {
		return "", err
	}

	translate := language != "en" && chooseTranslate(cfg.Transcription.DefaultTranslate)

	return whisper.Transcribe(ctx,
		cfg.Models.Whisper,
		media.AudioPath,
		cliutil.Confirm,
		whisperProgressPrinter(),
		whisper.WithLanguage(language),
		whisper.WithTranslate(translate),
	)
}

func chooseLanguage(defaultLanguage string) (string, error) {
	return cliutil.SelectOptionWithDefault("Select the spoken language", whisper.Languages(), defaultLanguage)
}

func chooseTranslate(defaultTranslate bool) bool {
	translate, _ := cliutil.ConfirmWithDefault("Translate to English?", defaultTranslate)
	return translate
}

func whisperProgressPrinter() whisper.PullProgressFunc {
	printProgress := cliutil.PrintProgress()

	return func(p whisper.PullProgress) error {
		return printProgress(cliutil.Progress{Label: "whisper", Status: "downloading", Total: p.Total, Completed: p.Completed})
	}
}

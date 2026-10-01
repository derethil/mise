package whisper

import (
	"fmt"

	cwhisper "github.com/ggerganov/whisper.cpp/bindings/go"
)

type Option func(ctx *cwhisper.Context, params *cwhisper.Params) error

func WithLanguage(language string) Option {
	return func(ctx *cwhisper.Context, params *cwhisper.Params) error {
		if language == "auto" {
			return params.SetLanguage(-1)
		}

		id := ctx.Whisper_lang_id(language)
		if id < 0 {
			return fmt.Errorf("unsupported language %q for this model", language)
		}

		if err := params.SetLanguage(id); err != nil {
			return fmt.Errorf("failed to set language %q: %w", language, err)
		}

		return nil
	}
}

func WithTranslate(translate bool) Option {
	return func(ctx *cwhisper.Context, params *cwhisper.Params) error {
		if translate && ctx.Whisper_is_multilingual() == 0 {
			return fmt.Errorf("translation requested, but the model is not multilingual")
		}

		params.SetTranslate(translate)
		return nil
	}
}

func Languages() []string {
	max := cwhisper.Whisper_lang_max_id()
	languages := make([]string, 0, max+2)
	languages = append(languages, "auto")

	for i := range max + 1 {
		languages = append(languages, cwhisper.Whisper_lang_str(i))
	}

	return languages
}

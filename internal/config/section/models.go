package section

import (
	"slices"
	"strings"
)

type ModelSize string

const (
	ModelSmall ModelSize = "small"
	ModelLarge ModelSize = "large"
)

type ModelsConfig struct {
	Small   string `key:"small" usage:"Model for simpler tasks, as provider/model"`
	Large   string `key:"large" usage:"Model for harder tasks, as provider/model"`
	Whisper string `key:"whisper" command:"import" usage:"Whisper model for transcription: tiny, base, small, medium, large, turbo (append .en for English-only)"`
}

func (m ModelsConfig) Get(size ModelSize) string {
	if size == ModelLarge {
		return m.Large
	}

	return m.Small
}

var WhisperModelTiers = []string{"tiny", "base", "small", "medium", "large", "turbo"}

var whisperEnglishOnlyTiers = []string{"tiny", "base", "small", "medium"}

func IsValidWhisperModel(model string) bool {
	tier, englishOnly := strings.CutSuffix(model, ".en")

	for _, t := range WhisperModelTiers {
		if tier == t {
			return !englishOnly || slices.Contains(whisperEnglishOnlyTiers, t)
		}
	}

	return false
}

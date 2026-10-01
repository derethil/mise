package section

type TranscriptionConfig struct {
	DefaultLanguage  string `key:"default_language" flag:"default-language" usage:"Default answer for the import language prompt, as a language code (e.g. \"en\")"`
	DefaultTranslate bool   `key:"default_translate" flag:"default-translate" usage:"Default answer for the import translate-to-English prompt"`
}

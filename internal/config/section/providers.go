// Package section provides configuration struct sections for the application.
package section

import "reflect"

type ProviderConfig struct {
	BaseURL string
	APIKey  string
	Timeout int
}

type ProviderSettings interface {
	Common() ProviderConfig
	ApplyCommon(ProviderConfig)
}

type OllamaConfig struct {
	BaseURL   string `key:"base_url" usage:"Base URL for the Ollama API, e.g. http://localhost:11434"`
	Timeout   int    `key:"timeout" flag:"-" usage:"Seconds to wait for a response from Ollama"`
	Autostart bool   `key:"autostart" usage:"Start Ollama automatically if not running, requires Ollama to be installed and in PATH"`
}

func (o *OllamaConfig) Common() ProviderConfig {
	return ProviderConfig{BaseURL: o.BaseURL, Timeout: o.Timeout}
}

func (o *OllamaConfig) ApplyCommon(cfg ProviderConfig) {
	o.BaseURL = cfg.BaseURL
	o.Timeout = cfg.Timeout
}

type ProvidersConfig struct {
	Ollama OllamaConfig `key:"ollama"`
}

func (p *ProvidersConfig) Get(name string) (ProviderConfig, bool) {
	settings, ok := p.settings(name)
	if !ok {
		return ProviderConfig{}, false
	}

	return settings.Common(), true
}

func (p *ProvidersConfig) Set(name string, cfg ProviderConfig) bool {
	settings, ok := p.settings(name)
	if !ok {
		return false
	}

	settings.ApplyCommon(cfg)
	return true
}

func (p *ProvidersConfig) settings(name string) (ProviderSettings, bool) {
	v := reflect.ValueOf(p).Elem()
	t := v.Type()

	for i := range t.NumField() {
		if t.Field(i).Tag.Get("key") != name {
			continue
		}

		settings, ok := v.Field(i).Addr().Interface().(ProviderSettings)
		return settings, ok
	}

	return nil, false
}

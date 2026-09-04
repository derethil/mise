package ai

import (
	"fmt"
	"strings"
)

type ModelRef struct {
	Provider string
	Name     string
	Tag      string
}

func (m ModelRef) String() string {
	if m.Tag == "" {
		return fmt.Sprintf("%s/%s", m.Provider, m.Name)
	}

	return fmt.Sprintf("%s/%s:%s", m.Provider, m.Name, m.Tag)
}

func ParseModel(raw string) (ModelRef, error) {
	provider, rest, ok := strings.Cut(raw, "/")
	if !ok {
		return ModelRef{}, fmt.Errorf("invalid model format, expected provider/model: %s", raw)
	}

	if _, ok := providerFactories[provider]; !ok {
		return ModelRef{}, fmt.Errorf("unsupported provider: %s", provider)
	}

	name, tag, _ := strings.Cut(rest, ":")

	return ModelRef{Provider: provider, Name: name, Tag: tag}, nil
}

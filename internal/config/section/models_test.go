package section

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelsConfigGet(t *testing.T) {
	models := ModelsConfig{Small: "ollama/small", Large: "ollama/large"}

	assert.Equal(t, "ollama/small", models.Get(ModelSmall))
	assert.Equal(t, "ollama/large", models.Get(ModelLarge))
	assert.Equal(t, "ollama/small", models.Get(ModelSize("")), "unrecognized sizes fall back to small")
}

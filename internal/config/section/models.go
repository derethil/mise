package section

type ModelSize string

const (
	ModelSmall ModelSize = "small"
	ModelLarge ModelSize = "large"
)

type ModelsConfig struct {
	Small string `key:"small" flag:"-" usage:"Model for simpler tasks, as provider/model"`
	Large string `key:"large" flag:"-" usage:"Model for harder tasks, as provider/model"`
}

func (m ModelsConfig) Get(size ModelSize) string {
	if size == ModelLarge {
		return m.Large
	}

	return m.Small
}

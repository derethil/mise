package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkSchemaFlagOptOut(t *testing.T) {
	type leaf struct {
		Plain  string `key:"plain"`
		Opaque string `key:"opaque" flag:"-"`
	}

	type root struct {
		Visible leaf `key:"visible"`
		Hidden  leaf `key:"hidden" flag:"-"`
	}

	flags := map[string]bool{}
	walkSchema(reflect.TypeFor[root](), "", func(f schemaField) {
		flags[f.Key] = f.Flag
	})

	assert.Equal(t, map[string]bool{
		"visible.plain":  true,
		"visible.opaque": false,
		"hidden.plain":   false,
		"hidden.opaque":  false,
	}, flags)
}

func TestWalkSchemaSkipsUntaggedFields(t *testing.T) {
	type root struct {
		Tagged   string `key:"tagged"`
		Untagged string
	}

	var keys []string
	walkSchema(reflect.TypeFor[root](), "", func(f schemaField) {
		keys = append(keys, f.Key)
	})

	assert.Equal(t, []string{"tagged"}, keys)
}

func TestProvidersConfigGet(t *testing.T) {
	providers := ProvidersConfig{
		Ollama: ProviderConfig{BaseURL: "http://localhost:11434", APIKey: "secret"},
	}

	cfg, ok := providers.Get("ollama")
	assert.True(t, ok)
	assert.Equal(t, providers.Ollama, cfg)

	_, ok = providers.Get("openai")
	assert.False(t, ok)
}

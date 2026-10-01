package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkSchemaFlagsAreOptIn(t *testing.T) {
	type leaf struct {
		Plain   string `key:"plain"`
		Exposed string `key:"exposed" flag:"public-name"`
	}

	type root struct {
		Section leaf `key:"section"`
	}

	flags := map[string]string{}
	walkSchema(reflect.TypeFor[root](), "", func(f schemaField) {
		if f.FlagName != "" {
			flags[f.Key] = f.FlagName
		}
	})

	assert.Equal(t, map[string]string{"section.exposed": "public-name"}, flags)
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

func TestWalkSchemaIndexResolvesThroughAnonymousEmbedding(t *testing.T) {
	type embedded struct {
		Value string `key:"value"`
	}

	type root struct {
		embedded
		Section struct {
			embedded
		} `key:"section"`
	}

	data := root{}
	data.Value = "top"
	data.Section.Value = "nested"

	indexes := map[string][]int{}
	walkSchema(reflect.TypeFor[root](), "", func(f schemaField) {
		indexes[f.Key] = f.Index
	})

	value := reflect.ValueOf(data)
	assert.Equal(t, "top", value.FieldByIndex(indexes["value"]).String())
	assert.Equal(t, "nested", value.FieldByIndex(indexes["section.value"]).String())
}

func TestFieldTypeOf(t *testing.T) {
	assert.Equal(t, fieldTypeString, fieldTypeOf(reflect.TypeOf("")))
	assert.Equal(t, fieldTypeStrings, fieldTypeOf(reflect.TypeOf([]string{})))
	assert.Equal(t, fieldTypeInt, fieldTypeOf(reflect.TypeOf(0)))
	assert.Equal(t, fieldTypeInt, fieldTypeOf(reflect.TypeOf(int64(0))))
	assert.Equal(t, fieldTypeBool, fieldTypeOf(reflect.TypeOf(true)))
	assert.Equal(t, fieldTypeString, fieldTypeOf(reflect.TypeOf(3.14)), "unhandled kinds default to string")
	assert.Equal(t, fieldTypeString, fieldTypeOf(reflect.TypeOf([]int{})), "only string slices get the multi-value flag type")
}

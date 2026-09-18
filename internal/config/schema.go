package config

import (
	"maps"
	"reflect"
	"slices"
)

type fieldType int

const (
	fieldTypeString fieldType = iota
	fieldTypeStrings
	fieldTypeInt
	fieldTypeBool
)

type schemaField struct {
	Key      string
	Usage    string
	Category string
	Command  string
	Type     fieldType

	// Flag determines whether this field is settable via command line flag. Fields with
	// `flag:"-"` are ignored and only settable via config file or environment variable.
	Flag bool
}

func walkSchema(t reflect.Type, prefix string, visit func(schemaField)) {
	walkFields(t, prefix, "", "", true, visit)
}

func walkFields(t reflect.Type, prefix, category, command string, flag bool, visit func(schemaField)) {
	for field := range t.Fields() {
		name := field.Tag.Get("key")
		if name == "" {
			continue
		}

		key := name
		if prefix != "" {
			key = prefix + "." + name
		}

		enabled := flag && field.Tag.Get("flag") != "-"

		fieldCategory := category
		if taggedCategory := field.Tag.Get("category"); taggedCategory != "" {
			fieldCategory = taggedCategory
		}

		fieldCommand := command
		if taggedCommand := field.Tag.Get("command"); taggedCommand != "" {
			fieldCommand = taggedCommand
		}

		if field.Type.Kind() == reflect.Struct {
			walkFields(field.Type, key, fieldCategory, fieldCommand, enabled, visit)
			continue
		}

		if field.Type == reflect.TypeFor[ProvidersConfig]() {
			for _, name := range slices.Sorted(maps.Keys(defaultConfig.Providers)) {
				walkFields(reflect.TypeFor[ProviderConfig](), key+"."+name, fieldCategory, fieldCommand, enabled, visit)
			}
			continue
		}

		visit(schemaField{
			Key:      key,
			Usage:    field.Tag.Get("usage"),
			Category: fieldCategory,
			Command:  fieldCommand,
			Type:     fieldTypeOf(field.Type),
			Flag:     enabled,
		})
	}
}

func fieldTypeOf(t reflect.Type) fieldType {
	switch t.Kind() {
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return fieldTypeStrings
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldTypeInt
	case reflect.Bool:
		return fieldTypeBool
	}

	return fieldTypeString
}

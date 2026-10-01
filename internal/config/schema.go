package config

import "reflect"

type fieldType int

const (
	fieldTypeString fieldType = iota
	fieldTypeStrings
	fieldTypeInt
	fieldTypeBool
)

type schemaField struct {
	Key      string
	Index    []int
	FlagName string
	Alias    string
	Usage    string
	Category string
	Command  string
	Type     fieldType
}

func walkSchema(t reflect.Type, prefix string, visit func(schemaField)) {
	walkFields(t, prefix, nil, "", "", visit)
}

func walkFields(t reflect.Type, prefix string, path []int, category, command string, visit func(schemaField)) {
	for field := range t.Fields() {
		fieldPath := append(append([]int{}, path...), field.Index...)

		name := field.Tag.Get("key")
		if name == "" {
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				walkFields(field.Type, prefix, fieldPath, category, command, visit)
			}
			continue
		}

		key := name
		if prefix != "" {
			key = prefix + "." + name
		}

		fieldCategory := category
		if taggedCategory := field.Tag.Get("category"); taggedCategory != "" {
			fieldCategory = taggedCategory
		}

		fieldCommand := command
		if taggedCommand := field.Tag.Get("command"); taggedCommand != "" {
			fieldCommand = taggedCommand
		}

		if field.Type.Kind() == reflect.Struct {
			walkFields(field.Type, key, fieldPath, fieldCategory, fieldCommand, visit)
			continue
		}

		flagName := field.Tag.Get("flag")

		visit(schemaField{
			Key:      key,
			Index:    fieldPath,
			FlagName: flagName,
			Alias:    field.Tag.Get("alias"),
			Usage:    field.Tag.Get("usage"),
			Category: fieldCategory,
			Command:  fieldCommand,
			Type:     fieldTypeOf(field.Type),
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

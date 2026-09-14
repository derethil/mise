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
	Key   string
	Usage string
	Type  fieldType

	// Flag determines whether this field is settable via command line flag. Fields with
	// `flag:"-"` are ignored and only settable via config file or environment variable.
	Flag bool
}

func walkSchema(t reflect.Type, prefix string, visit func(schemaField)) {
	walkFields(t, prefix, true, visit)
}

func walkFields(t reflect.Type, prefix string, flag bool, visit func(schemaField)) {
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

		if field.Type.Kind() == reflect.Struct {
			walkFields(field.Type, key, enabled, visit)
			continue
		}

		visit(schemaField{
			Key:   key,
			Usage: field.Tag.Get("usage"),
			Type:  fieldTypeOf(field.Type),
			Flag:  enabled,
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

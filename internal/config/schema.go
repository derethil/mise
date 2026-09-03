package config

import "reflect"

type schemaField struct {
	Key   string
	Usage string

	// Flag determines whether this field is settable via command line flag. Fields with
	// `flag:"-"` are ignored and only settable via config file or environment variable.
	Flag bool
}

func walkSchema(t reflect.Type, prefix string, visit func(schemaField)) {
	for i := range t.NumField() {
		field := t.Field(i)

		name := field.Tag.Get("key")
		if name == "" {
			continue
		}

		key := name
		if prefix != "" {
			key = prefix + "." + name
		}

		if field.Type.Kind() == reflect.Struct {
			walkSchema(field.Type, key, visit)
			continue
		}

		visit(schemaField{
			Key:   key,
			Usage: field.Tag.Get("usage"),
			Flag:  field.Tag.Get("flag") != "-",
		})
	}
}

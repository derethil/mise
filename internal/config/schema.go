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
			Flag:  enabled,
		})
	}
}

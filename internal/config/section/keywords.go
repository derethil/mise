package section

type KeywordsConfig struct {
	SchemaFile string   `key:"schema_file" flag:"schema-file" usage:"Path to the file describing your keyword schema"`
	Keep       []string `key:"keep" flag:"keep" alias:"k" usage:"Comma-separated keywords to preserve when using recipe keyword --replace"`
}

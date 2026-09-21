package section

type KeywordsConfig struct {
	SchemaFile string   `key:"schema_file" usage:"Path to the file describing your keyword schema"`
	Ignore     []string `key:"ignore" usage:"Keywords to protect from removal when using recipe keyword --replace"`
}

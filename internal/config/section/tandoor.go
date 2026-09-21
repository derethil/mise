package section

type TandoorConfig struct {
	Token     string `key:"token" usage:"Tandoor API token"`
	BaseURL   string `key:"base_url" usage:"Tandoor base URL"`
	BackupDir string `key:"backup_dir" flag:"-" usage:"Directory to store recipe backups"`
}

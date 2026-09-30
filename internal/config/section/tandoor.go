package section

type TandoorConfig struct {
	Token     string `key:"token" usage:"Tandoor API token"`
	BaseURL   string `key:"base_url" flag:"tandoor-url" usage:"Tandoor base URL"`
	BackupDir string `key:"backup_dir" usage:"Directory to store recipe backups"`
}

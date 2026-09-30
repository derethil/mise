package section

type BackupConfig struct {
	Keep int `key:"keep" usage:"Number of backups to keep per recipe (0 = keep all)"`
}

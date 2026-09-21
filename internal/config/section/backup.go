package section

type BackupConfig struct {
	Keep int `key:"keep" flag:"-" usage:"Number of backups to keep per recipe (0 = keep all)"`
}

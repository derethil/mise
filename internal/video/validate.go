package video

import "strings"

var SupportedImportSources = []string{
	"tiktok.com",
	"youtube.com",
	"instagram.com",
}

func IsSupportedImportSource(host string) bool {
	host = strings.ToLower(host)

	for _, domain := range SupportedImportSources {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

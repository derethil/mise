package video

import (
	"slices"
	"strings"
)

var SupportedImportSources = []string{
	"tiktok.com",
	"youtube.com",
	"instagram.com",
}

var SupportedBrowsers = []string{
	"chrome",
	"chromium",
	"edge",
	"firefox",
	"opera",
	"safari",
	"vivaldi",
	"whale",
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

func IsSupportedBrowser(browser string) bool {
	return slices.Contains(SupportedBrowsers, strings.ToLower(browser))
}

package section

type VideoConfig struct {
	Format             string   `key:"format" usage:"yt-dlp format selector"`
	CookiesFile        string   `key:"cookies_file" usage:"Path to a Netscape-format cookie jar for login-walled videos"`
	CookiesFromBrowser string   `key:"cookies_from_browser" usage:"Browser to pull cookies from, e.g. firefox"`
	Impersonate        string   `key:"impersonate" usage:"TLS client to impersonate, e.g. chrome. Often required for Instagram"`
	YtdlpPath          string   `key:"ytdlp_path" usage:"Path to a yt-dlp binary. Empty resolves from PATH, downloading one if absent"`
	YtdlpArgs          []string `key:"ytdlp_args" usage:"Extra arguments passed through to yt-dlp"`
	MaxDurationMinutes int      `key:"max_duration_minutes" usage:"Refuse videos longer than this. 0 disables the check"`
}

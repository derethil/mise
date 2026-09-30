package section

type ExtractConfig struct {
	Format             string   `key:"format" flag:"format" category:"DOWNLOAD OPTIONS" usage:"yt-dlp format selector"`
	CookiesFile        string   `key:"cookies_file" flag:"cookies-file" category:"DOWNLOAD OPTIONS" usage:"Path to a Netscape-format cookie jar for login-walled videos"`
	CookiesFromBrowser string   `key:"cookies_from_browser" flag:"cookies-from-browser" category:"DOWNLOAD OPTIONS" usage:"Browser to pull cookies from, e.g. firefox"`
	Impersonate        string   `key:"impersonate" flag:"impersonate" category:"DOWNLOAD OPTIONS" usage:"TLS client to impersonate, e.g. chrome. Often required for Instagram"`
	YtdlpPath          string   `key:"ytdlp_path" flag:"yt-dlp-path" category:"TOOL OPTIONS" usage:"Path to a yt-dlp binary. Empty resolves yt-dlp from PATH"`
	FfmpegPath         string   `key:"ffmpeg_path" flag:"ffmpeg-path" category:"TOOL OPTIONS" usage:"Path to an ffmpeg binary. Empty resolves ffmpeg from PATH"`
	YtdlpArgs          []string `key:"ytdlp_args" flag:"yt-dlp-arg" category:"TOOL OPTIONS" usage:"Comma-separated extra arguments passed through to yt-dlp"`
	MaxDurationMinutes int      `key:"max_duration_minutes" flag:"max-duration" category:"DOWNLOAD OPTIONS" usage:"Refuse videos longer than this many minutes. 0 disables the check"`
}

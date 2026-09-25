package importcmd

import (
	"net/url"
	"os"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/video"
	"github.com/urfave/cli/v3"
)

func validateUrl(cmd *cli.Command) error {
	urlArg := cmd.Args().Get(0)
	if urlArg == "" {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Missing required argument: url")
	}

	parsed, err := url.Parse(urlArg)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "%s is not a valid url", urlArg)
	}

	if !video.IsSupportedImportSource(parsed.Host) {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "URL %s is not a supported import source %v", urlArg, video.SupportedImportSources)
	}

	return nil
}

func validateCookiesFromFile(cmd *cli.Command) error {
	value := cmd.String(flagCookiesFile)
	if value == "" {
		return nil
	}

	info, err := os.Stat(value)
	if err != nil {
		if os.IsNotExist(err) {
			return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Cookies file %s does not exist", value)
		}

		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Cookies file %s could not be read: %v", value, err)
	}

	if info.IsDir() {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Cookies file %s is a directory, not a file", value)
	}

	return nil
}

func validateExclusiveCookies(cmd *cli.Command) error {
	if cmd.String(flagCookiesFromBrowser) != "" && cmd.String(flagCookiesFile) != "" {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "--%s and --%s cannot be used together", flagCookiesFromBrowser, flagCookiesFile)
	}

	return nil
}

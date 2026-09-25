package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/video"
	"github.com/urfave/cli/v3"
)

var importCmd = &cli.Command{
	Name:  "import",
	Usage: "Import a recipe from various social media video sources",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:     "url",
			Required: true,
		},
	},
	Flags: config.FlagsForCommand("import"),
	ArgValidator: func(ctx context.Context, cmd *cli.Command) error {
		validators := []func(*cli.Command) error{
			validateUrl,
			validateCookiesFromBrowser,
			validateCookiesFromFile,
			validateExclusiveCookies,
		}

		for _, validator := range validators {
			if err := validator(cmd); err != nil {
				return err
			}
		}

		return nil

	},
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		configPath := cliutil.ResolveFlag(cmd, cliutil.GlobalFlagConfig, config.DefaultConfigPath())
		cfg, err := config.Load(cmd, configPath)
		if err != nil {
			return ctx, err
		}

		return config.NewContext(ctx, cfg), nil
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		url := cmd.StringArg("url")
		cfg := config.FromContext(ctx)

		printProgress := cliutil.PrintProgress()
		opts := []video.YtdlpOption{
			video.WithConfig(cfg.Video),
			video.WithProgress(func(p video.Progress) {
				_ = printProgress(cliutil.Progress{
					Label:     "video",
					Status:    p.Status,
					Total:     p.Total,
					Completed: p.Completed,
				})
			}),
		}

		workdir, err := video.DownloadVideo(ctx, url, opts, cfg.Video.YtdlpArgs...)
		if err != nil {
			return err
		}

		fmt.Println()

		return workdir.Cleanup()
	},
}

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

func validateCookiesFromBrowser(cmd *cli.Command) error {
	value := cmd.String("video.cookies-from-browser")
	if value == "" {
		return nil
	}

	if !video.IsSupportedBrowser(value) {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Browser %s is not a supported browser %v", value, video.SupportedBrowsers)
	}

	return nil
}

func validateCookiesFromFile(cmd *cli.Command) error {
	value := cmd.String("video.cookies-file")
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
	if cmd.String("video.cookies-from-browser") != "" && cmd.String("video.cookies-file") != "" {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "--video.cookies-from-browser and --video.cookies-file cannot be used together")
	}

	return nil
}

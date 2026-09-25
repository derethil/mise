package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/video"
	"github.com/urfave/cli/v3"
)

const (
	flagCookiesFile        = "video.cookies-file"
	flagCookiesFromBrowser = "video.cookies-from-browser"
	flagDryRun             = "dry-run"
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
	Flags: append(config.FlagsForCommand("import"), &cli.BoolFlag{
		Name:  flagDryRun,
		Usage: "Fetch and print video metadata without downloading",
	}),
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
		cfg := config.FromContext(ctx)

		extraction, err := video.Extract(ctx, cmd.StringArg("url"), cfg.Video, progressPrinter())
		if err != nil {
			return cliutil.VideoUserError(err)
		}

		showExtractionInfo(*extraction)

		if cmd.Bool(flagDryRun) {
			slog.InfoContext(ctx, "Dry run complete, no import performed.")
			return nil
		}

		media, err := extraction.Download(ctx)
		if err != nil {
			return cliutil.VideoUserError(err)
		}

		fmt.Printf("\nFile:     %s\n", media.Path)
		fmt.Println("\nTranscription, extraction, and Tandoor creation aren't built yet — the download is left in the workdir above.")

		return nil
	},
}

func progressPrinter() video.ProgressFunc {
	printProgress := cliutil.PrintProgress()

	return func(p video.Progress) {
		_ = printProgress(cliutil.Progress{
			Label:     "video",
			Status:    p.Status,
			Total:     p.Total,
			Completed: p.Completed,
		})
	}
}

func showExtractionInfo(e video.Extraction) {
	source := e.Source
	workdir := e.WorkDir()

	fmt.Printf("Title:    %s\n", source.Title)

	if source.Uploader != "" {
		fmt.Printf("Uploader: %s\n", source.Uploader)
	}

	fmt.Printf("Duration: %s\n", source.Duration)
	fmt.Printf("Source:   %s\n", source.URL)
	fmt.Printf("Workdir:  %s\n", workdir)
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
	value := cmd.String(flagCookiesFromBrowser)
	if value == "" {
		return nil
	}

	if !video.IsSupportedBrowser(value) {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Browser %s is not a supported browser %v", value, video.SupportedBrowsers)
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

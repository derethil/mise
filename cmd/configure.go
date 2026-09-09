package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var configureCmd = &cli.Command{
	Name:  "configure",
	Usage: "Provide configuration values for mise",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		path := cliutil.ResolveFlag(cmd, cliutil.GlobalFlagConfig, config.DefaultConfigPath())

		cfg, err := loadConfigFromFile(cmd, path)
		if err != nil {
			return cliutil.ErrWithUserMessage(err, "failed to load existing configuration")
		}

		err = configureTandoor(ctx, &cfg)
		if err != nil {
			return err
		}

		err = configureProviders(&cfg)
		if err != nil {
			return err
		}

		if err := config.Save(cfg, path); err != nil {
			return cliutil.ErrWithUserMessage(err, "failed to save configuration")
		}

		return nil
	},
}

func loadConfigFromFile(cmd *cli.Command, path string) (config.Config, error) {
	exists, err := os.Stat(path)
	if err == nil && exists.IsDir() {
		return config.Config{}, fmt.Errorf("path %s is a directory, please provide a file path", path)
	}
	if err != nil && !os.IsNotExist(err) {
		return config.Config{}, fmt.Errorf("error checking path %s: %w", path, err)
	}

	return config.Load(cmd, path)
}

func validateBaseURL(input string) error {
	parsed, err := url.ParseRequestURI(input)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("must be a valid URL")
	}

	return nil
}

func configureTandoor(ctx context.Context, cfg *config.Config) error {
	var baseURL, apiToken string

	validateAPIToken := func(input string) error {
		if len(input) != 40 || !strings.HasPrefix(input, "tda_") {
			return fmt.Errorf("must start with tda_ and be 40 characters long")
		}

		return nil
	}

	promptTandoor := func() error {
		var err error

		baseURL, err = cliutil.PromptForInput("Tandoor Base URL", cfg.Tandoor.BaseURL, validateBaseURL)
		if err != nil {
			return err
		}

		apiToken, err = cliutil.PromptForHiddenInput("Tandoor API Token", cfg.Tandoor.Token, validateAPIToken)
		if err != nil {
			return err
		}

		client := tandoor.NewClient(baseURL, apiToken)
		err = client.TestConnection(ctx)
		if err != nil {
			return err
		}

		return nil
	}

	err := promptTandoor()
	for err != nil {
		if cliutil.IsUserAbort(err) {
			return err
		}

		slog.Error("Tandoor is not reachable with the provided configuration. Please try again.", "error", err)

		err = promptTandoor()
	}

	cfg.Tandoor.BaseURL = baseURL
	cfg.Tandoor.Token = apiToken

	return nil
}

func configureModels(cfg *config.Config) (ai.ModelRef, ai.ModelRef, error) {
	validateModel := func(input string) error {
		_, err := ai.ParseModel(input)
		return err
	}

	smallModelInput, err := cliutil.PromptForInput("Small model (e.g., ollama/qwen2.5:7b)", cfg.Models.Small, validateModel)
	if err != nil {
		return ai.ModelRef{}, ai.ModelRef{}, err
	}

	largeModelInput, err := cliutil.PromptForInput("Large model (e.g., ollama/qwen2.5:14b)", cfg.Models.Large, validateModel)
	if err != nil {
		return ai.ModelRef{}, ai.ModelRef{}, err
	}

	smallModel, _ := ai.ParseModel(smallModelInput)
	largeModel, _ := ai.ParseModel(largeModelInput)

	return smallModel, largeModel, nil

}

func configureProvider(cfg *config.Config, provider string) error {
	providerCfg, _ := cfg.Providers.Get(provider)

	baseURL, err := cliutil.PromptForInput(fmt.Sprintf("%s Base URL", provider), providerCfg.BaseURL, validateBaseURL)
	if err != nil {
		return err
	}
	providerCfg.BaseURL = baseURL

	if provider != ai.ProviderOllama {
		apiKey, err := cliutil.PromptForHiddenInput(fmt.Sprintf("%s API Key", provider), providerCfg.APIKey)
		if err != nil {
			return err
		}
		providerCfg.APIKey = apiKey
	}

	cfg.Providers.Set(provider, providerCfg)

	return nil
}

func configureProviders(cfg *config.Config) error {
	fmt.Println("Supported providers:", strings.Join(ai.SupportedProviders(), ", "))

	smallModel, largeModel, err := configureModels(cfg)
	if err != nil {
		return err
	}

	var usedProviders []string
	for _, provider := range []string{smallModel.Provider, largeModel.Provider} {
		if !slices.Contains(usedProviders, provider) {
			usedProviders = append(usedProviders, provider)
		}
	}

	for _, provider := range usedProviders {
		err := configureProvider(cfg, provider)
		if err != nil {
			return err
		}
	}

	return nil
}

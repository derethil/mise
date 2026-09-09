package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v3"
)

type ConfigureSuite struct {
	suite.Suite
}

func TestConfigureSuite(t *testing.T) {
	suite.Run(t, new(ConfigureSuite))
}

func (s *ConfigureSuite) TestValidateBaseURL_AcceptsValidURL() {
	s.NoError(validateBaseURL("https://example.com"))
}

func (s *ConfigureSuite) TestValidateBaseURL_RejectsMissingScheme() {
	s.Error(validateBaseURL("example.com"))
}

func (s *ConfigureSuite) TestValidateBaseURL_RejectsGarbage() {
	s.Error(validateBaseURL("not a url"))
}

func (s *ConfigureSuite) TestConfigureTandoor_ErrorsWithoutInteractiveInput() {
	var cfg config.Config

	err := configureTandoor(context.Background(), &cfg)

	s.Require().Error(err)
	s.True(cliutil.IsUserAbort(err))
}

func (s *ConfigureSuite) TestConfigureModels_ErrorsWithoutInteractiveInput() {
	var cfg config.Config

	_, _, err := configureModels(&cfg)

	s.Require().Error(err)
	s.True(cliutil.IsUserAbort(err))
}

func (s *ConfigureSuite) TestConfigureProvider_ErrorsWithoutInteractiveInput() {
	var cfg config.Config

	err := configureProvider(&cfg, ai.ProviderOllama)

	s.Require().Error(err)
	s.True(cliutil.IsUserAbort(err))
}

func (s *ConfigureSuite) TestConfigureProviders_ErrorsWithoutInteractiveInput() {
	var cfg config.Config

	err := configureProviders(&cfg)

	s.Require().Error(err)
	s.True(cliutil.IsUserAbort(err))
}

func (s *ConfigureSuite) TestConfigureCmd_DoesNotWriteConfigWithoutInteractiveInput() {
	path := filepath.Join(s.T().TempDir(), "config.toml")

	err := runConfigureCmd(s.T(), "--config", path)

	s.Require().Error(err)
	s.True(cliutil.IsUserAbort(err))

	_, statErr := os.Stat(path)
	s.True(os.IsNotExist(statErr), "config file should not be written when configuration is incomplete")
}

func runConfigureCmd(t *testing.T, args ...string) error {
	t.Helper()

	root := &cli.Command{
		Name:  "mise",
		Flags: globalFlags,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			configPath := cliutil.ResolveFlag(cmd, cliutil.GlobalFlagConfig, config.DefaultConfigPath())

			cfg, err := config.Load(cmd, configPath)
			if err != nil {
				return ctx, err
			}

			return config.NewContext(ctx, cfg), nil
		},
		Commands: []*cli.Command{configureCmd},
	}

	return root.Run(context.Background(), append([]string{"mise", "configure"}, args...))
}

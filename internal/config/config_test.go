package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v3"
)

type ConfigSuite struct {
	suite.Suite
	origConfigDir string
}

func (s *ConfigSuite) SetupTest() {
	s.origConfigDir = ConfigDir
	ConfigDir = s.T().TempDir()
}

func (s *ConfigSuite) TearDownTest() {
	ConfigDir = s.origConfigDir
}

func (s *ConfigSuite) writeConfigFile(contents string) {
	err := os.WriteFile(filepath.Join(ConfigDir, "config.toml"), []byte(contents), 0o644)
	s.Require().NoError(err)
}

func (s *ConfigSuite) load(args ...string) Config {
	var (
		cfg Config
		err error
	)

	cmd := &cli.Command{
		Name:  "mise",
		Flags: Flags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err = Load(cmd)
			return err
		},
	}

	runErr := cmd.Run(context.Background(), append([]string{"mise"}, args...))
	s.Require().NoError(runErr)
	s.Require().NoError(err)

	return cfg
}

func (s *ConfigSuite) TestDefaults() {
	cfg := s.load()

	s.Equal("https://tandoor.dev/api/v1", cfg.Tandoor.BaseURL)
	s.Empty(cfg.Tandoor.Token)
	s.Equal(filepath.Join(DataDir, "tandoor_backups"), cfg.Tandoor.BackupDir)
}

func (s *ConfigSuite) TestConfigFileOverridesDefaults() {
	s.writeConfigFile(`
[tandoor]
token = "from-file"
base_url = "https://from-file.example"
`)

	cfg := s.load()

	s.Equal("from-file", cfg.Tandoor.Token)
	s.Equal("https://from-file.example", cfg.Tandoor.BaseURL)
}

func (s *ConfigSuite) TestEnvOverridesConfigFile() {
	s.writeConfigFile(`
[tandoor]
base_url = "https://from-file.example"
`)
	s.T().Setenv("MISE_TANDOOR_BASE_URL", "https://from-env.example")

	cfg := s.load()

	s.Equal("https://from-env.example", cfg.Tandoor.BaseURL)
}

func (s *ConfigSuite) TestFlagOverridesEnv() {
	s.T().Setenv("MISE_TANDOOR_BASE_URL", "https://from-env.example")

	cfg := s.load("--tandoor.base_url", "https://from-flag.example")

	s.Equal("https://from-flag.example", cfg.Tandoor.BaseURL)
}

func (s *ConfigSuite) TestEnvKeyDisambiguation() {
	s.T().Setenv("MISE_TANDOOR_BASE_URL", "https://disambiguated.example")

	cfg := s.load()

	s.Equal("https://disambiguated.example", cfg.Tandoor.BaseURL)
}

func (s *ConfigSuite) TestUnknownEnvVarsAreIgnored() {
	s.T().Setenv("MISE_SOME_UNRELATED_SETTING", "should-be-ignored")

	cfg := s.load()

	s.Equal("https://tandoor.dev/api/v1", cfg.Tandoor.BaseURL)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

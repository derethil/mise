package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestGlobalHelpCategoriesAreRelevantToCommand(t *testing.T) {
	tests := []struct {
		name       string
		command    *cli.Command
		categories []string
	}{
		{name: "logs", command: rootCmd.Command("logs"), categories: []string{"GENERAL OPTIONS"}},
		{name: "recipe backup", command: rootCmd.Command("recipe").Command("backup"), categories: []string{"GENERAL OPTIONS", "TANDOOR OPTIONS"}},
		{name: "recipe keyword", command: rootCmd.Command("recipe").Command("keyword"), categories: []string{"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"}},
		{name: "models clear", command: rootCmd.Command("models").Command("clear"), categories: []string{"GENERAL OPTIONS", "PROVIDER OPTIONS"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := test.command.Metadata[visibleGlobalFlagCategoriesMetadataKey].(globalFlagCategories)
			var names []string
			for _, category := range metadata.VisibleFlagCategories() {
				names = append(names, category.Name())
			}
			assert.Equal(t, test.categories, names)
		})
	}
}

func TestEveryCommandDeclaresGlobalFlagCategories(t *testing.T) {
	var check func(*cli.Command)
	check = func(command *cli.Command) {
		if command.Name == "help" {
			return
		}

		categories, ok := command.Metadata[cliutil.GlobalFlagCategoriesMetadataKey].([]string)
		assert.True(t, ok, "%s must declare its global flag categories", command.Name)
		assert.Contains(t, categories, cliutil.GeneralOptions)

		for _, child := range command.Commands {
			check(child)
		}
	}

	check(rootCmd)
}

func TestCompletionIncludesRelevantGlobalFlags(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{"config", "verbose", "model", "ollama-url", "start-ollama", "tandoor-url"},
		flagNames(relevantGlobalFlags(rootCmd.Command("import"))),
	)
	assert.ElementsMatch(t,
		[]string{"config", "verbose", "tandoor-url"},
		flagNames(relevantGlobalFlags(rootCmd.Command("recipe").Command("backup"))),
	)
	assert.ElementsMatch(t,
		[]string{"config", "verbose"},
		flagNames(relevantGlobalFlags(rootCmd.Command("logs"))),
	)
}

func TestOptionsSectionRequiresLocalOptions(t *testing.T) {
	withoutOptions := renderCommandHelp(t, nil)
	assert.NotContains(t, withoutOptions, "\nOPTIONS:")
	assert.Contains(t, withoutOptions, "\nGLOBAL OPTIONS:")
	assert.Contains(t, withoutOptions, "mise child [options]\n\nGLOBAL OPTIONS:")

	withOptions := renderCommandHelp(t, []cli.Flag{&cli.BoolFlag{Name: "local"}})
	assert.Contains(t, withOptions, "\nOPTIONS:")
	assert.Contains(t, withOptions, "--local")
	assert.Contains(t, withOptions, "--help, -h  show help\n\nGLOBAL OPTIONS:")

	withArgument := renderCommandHelpWithArguments(t, nil, []cli.Argument{&cli.StringArg{Name: "value"}})
	assert.Contains(t, withArgument, "mise child [options] [value]\n")
	assert.NotContains(t, withArgument, "[value] \n")
}

func renderCommandHelp(t *testing.T, localFlags []cli.Flag) string {
	return renderCommandHelpWithArguments(t, localFlags, nil)
}

func renderCommandHelpWithArguments(t *testing.T, localFlags []cli.Flag, arguments []cli.Argument) string {
	t.Helper()

	var output bytes.Buffer
	command := &cli.Command{
		Name:      "child",
		Flags:     localFlags,
		Arguments: arguments,
		Metadata:  cliutil.GlobalFlagMetadata(),
	}
	root := &cli.Command{
		Name:     "mise",
		Writer:   &output,
		Flags:    []cli.Flag{&cli.BoolFlag{Name: "verbose", Category: cliutil.GeneralOptions}},
		Metadata: cliutil.GlobalFlagMetadata(),
		Commands: []*cli.Command{command},
	}
	configureGlobalHelp(root, root.Flags)

	err := root.Run(context.Background(), []string{"mise", "child", "--help"})
	assert.NoError(t, err)

	return output.String()
}

func flagNames(flags []cli.Flag) []string {
	names := make([]string, 0, len(flags))
	for _, flag := range flags {
		names = append(names, flag.Names()[0])
	}

	return names
}

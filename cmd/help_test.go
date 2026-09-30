package cmd

import (
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

func flagNames(flags []cli.Flag) []string {
	names := make([]string, 0, len(flags))
	for _, flag := range flags {
		names = append(names, flag.Names()[0])
	}

	return names
}

package cmd

import (
	"context"
	"sort"

	"github.com/urfave/cli/v3"
)

func init() {
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
}

const commandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}
GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .Metadata.globalFlagCategories}}
   See 'mise --help' for all global configuration overrides.{{end}}
`

const subcommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}
GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .Metadata.globalFlagCategories}}
   See 'mise --help' for all global configuration overrides.{{end}}
`

var commandGlobalCategories = map[string][]string{
	"":                 {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"configure":        {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"genkit":           {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"import":           {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"logs":             {"GENERAL OPTIONS"},
	"models":           {"GENERAL OPTIONS", "PROVIDER OPTIONS"},
	"models clear":     {"GENERAL OPTIONS", "PROVIDER OPTIONS"},
	"models list":      {"GENERAL OPTIONS", "PROVIDER OPTIONS"},
	"models pull":      {"GENERAL OPTIONS", "PROVIDER OPTIONS"},
	"recipe":           {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"recipe backup":    {"GENERAL OPTIONS", "TANDOOR OPTIONS"},
	"recipe keyword":   {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"recipe normalize": {"GENERAL OPTIONS", "PROVIDER OPTIONS", "TANDOOR OPTIONS"},
	"recipe restore":   {"GENERAL OPTIONS", "TANDOOR OPTIONS"},
}

type globalFlagCategories struct {
	categories []cli.VisibleFlagCategory
}

func (g globalFlagCategories) VisibleFlagCategories() []cli.VisibleFlagCategory {
	return g.categories
}

type globalFlagCategory struct {
	name  string
	flags []cli.Flag
}

func (g globalFlagCategory) Name() string {
	return g.name
}

func (g globalFlagCategory) Flags() []cli.Flag {
	return g.flags
}

func newGlobalFlagCategories(flags []cli.Flag, allowedNames ...string) globalFlagCategories {
	allowed := make(map[string]bool, len(allowedNames))
	for _, name := range allowedNames {
		allowed[name] = true
	}

	byCategory := make(map[string][]cli.Flag)
	for _, flag := range flags {
		category := flag.(cli.CategorizableFlag).GetCategory()
		if len(allowed) > 0 && !allowed[category] {
			continue
		}
		byCategory[category] = append(byCategory[category], flag)
	}

	names := make([]string, 0, len(byCategory))
	for name := range byCategory {
		names = append(names, name)
	}
	sort.Strings(names)

	categories := make([]cli.VisibleFlagCategory, 0, len(names))
	for _, name := range names {
		flags := byCategory[name]
		sort.Slice(flags, func(i, j int) bool {
			return flags[i].String() < flags[j].String()
		})
		categories = append(categories, globalFlagCategory{name: name, flags: flags})
	}

	return globalFlagCategories{categories: categories}
}

func configureGlobalHelp(cmd *cli.Command, path string, flags []cli.Flag) {
	if cmd.Metadata == nil {
		cmd.Metadata = make(map[string]any)
	}
	categories, ok := commandGlobalCategories[path]
	if !ok {
		categories = []string{"GENERAL OPTIONS"}
	}
	cmd.Metadata["globalFlagCategories"] = newGlobalFlagCategories(flags, categories...)
	if path != "" {
		cmd.ShellComplete = completeWithGlobalFlags
	}

	for _, child := range cmd.Commands {
		childPath := child.Name
		if path != "" {
			childPath = path + " " + child.Name
		}
		configureGlobalHelp(child, childPath, flags)
	}
}

func completeWithGlobalFlags(ctx context.Context, cmd *cli.Command) {
	localFlags := cmd.Flags
	cmd.Flags = append(append([]cli.Flag{}, localFlags...), relevantGlobalFlags(cmd)...)
	defer func() { cmd.Flags = localFlags }()

	cli.DefaultCompleteWithFlags(ctx, cmd)
}

func relevantGlobalFlags(cmd *cli.Command) []cli.Flag {
	categories := cmd.Metadata["globalFlagCategories"].(globalFlagCategories)
	var flags []cli.Flag
	for _, category := range categories.VisibleFlagCategories() {
		flags = append(flags, category.Flags()...)
	}

	return flags
}

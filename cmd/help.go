package cmd

import (
	"context"
	"sort"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/urfave/cli/v3"
)

func init() {
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
}

const commandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleFlags}} [options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{trim .ArgsUsage}}{{else}}{{range .Arguments}} {{.Usage}}{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .Metadata.hasLocalOptions}}

OPTIONS:{{if .VisibleFlagCategories}}{{template "visibleFlagCategoryTemplate" .}}{{else}}{{template "visibleFlagTemplate" .}}
{{end}}{{else}}
{{end}}{{if .VisiblePersistentFlags}}
GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .Metadata.visibleGlobalFlagCategories}}
   See 'mise --help' for all global configuration overrides.{{end}}
`

const subcommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleFlags}} [options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{trim .ArgsUsage}}{{else}}{{range .Arguments}} {{.Usage}}{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .Metadata.hasLocalOptions}}

OPTIONS:{{if .VisibleFlagCategories}}{{template "visibleFlagCategoryTemplate" .}}{{else}}{{template "visibleFlagTemplate" .}}
{{end}}{{else}}
{{end}}{{if .VisiblePersistentFlags}}
GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .Metadata.visibleGlobalFlagCategories}}
   See 'mise --help' for all global configuration overrides.{{end}}
`

const (
	hasLocalOptionsMetadataKey             = "hasLocalOptions"
	visibleGlobalFlagCategoriesMetadataKey = "visibleGlobalFlagCategories"
)

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

func configureGlobalHelp(cmd *cli.Command, flags []cli.Flag) {
	if cmd.Metadata == nil {
		cmd.Metadata = map[string]any{}
	}

	categories, _ := cmd.Metadata[cliutil.GlobalFlagCategoriesMetadataKey].([]string)
	cmd.Metadata[hasLocalOptionsMetadataKey] = len(cmd.Flags) > 0
	cmd.Metadata[visibleGlobalFlagCategoriesMetadataKey] = newGlobalFlagCategories(flags, categories...)
	if cmd.Name != "mise" {
		cmd.ShellComplete = completeWithGlobalFlags
	}

	for _, child := range cmd.Commands {
		configureGlobalHelp(child, flags)
	}
}

func completeWithGlobalFlags(ctx context.Context, cmd *cli.Command) {
	localFlags := cmd.Flags
	cmd.Flags = append(append([]cli.Flag{}, localFlags...), relevantGlobalFlags(cmd)...)
	defer func() { cmd.Flags = localFlags }()

	cli.DefaultCompleteWithFlags(ctx, cmd)
}

func relevantGlobalFlags(cmd *cli.Command) []cli.Flag {
	categories, _ := cmd.Metadata[visibleGlobalFlagCategoriesMetadataKey].(globalFlagCategories)
	var flags []cli.Flag
	for _, category := range categories.VisibleFlagCategories() {
		flags = append(flags, category.Flags()...)
	}

	return flags
}

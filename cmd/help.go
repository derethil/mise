package cmd

import (
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

GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .Root.Metadata.globalFlagCategories}}{{end}}
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

GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .Root.Metadata.globalFlagCategories}}{{end}}
`

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

func newGlobalFlagCategories(flags []cli.Flag) globalFlagCategories {
	byCategory := make(map[string][]cli.Flag)
	for _, flag := range flags {
		category := flag.(cli.CategorizableFlag).GetCategory()
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

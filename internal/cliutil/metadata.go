package cliutil

const (
	GlobalFlagCategoriesMetadataKey = "globalFlagCategories"
	GeneralOptions                  = "GENERAL OPTIONS"
	ProviderOptions                 = "PROVIDER OPTIONS"
	TandoorOptions                  = "TANDOOR OPTIONS"
)

func GlobalFlagMetadata(categories ...string) map[string]any {
	categories = append([]string{GeneralOptions}, categories...)

	return map[string]any{GlobalFlagCategoriesMetadataKey: categories}
}

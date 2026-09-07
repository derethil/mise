package cleaningredients

import (
	"fmt"
	"strconv"
	"strings"
)

type Change struct {
	Before string
	Row    CleanedRow
}

func (c Change) String() string {
	switch c.Row.Kind {
	case KindJunk:
		return fmt.Sprintf("  [drop]   %s", c.Before)
	case KindHeader:
		return fmt.Sprintf("  [header] %s", c.Before)
	default:
		return fmt.Sprintf("  %s -> %s", c.Before, formatIngredient(c.Row))
	}
}

func formatIngredient(row CleanedRow) string {
	var parts []string

	if row.Amount != 0 {
		parts = append(parts, fmt.Sprintf("amount: %s", strconv.FormatFloat(row.Amount, 'g', -1, 64)))
	}
	if row.Unit != "" {
		parts = append(parts, fmt.Sprintf("unit: %s", row.Unit))
	}
	if row.Food != "" {
		parts = append(parts, fmt.Sprintf("food: %s", row.Food))
	}
	if row.Note != "" {
		parts = append(parts, fmt.Sprintf("note: %s", row.Note))
	}

	return strings.Join(parts, " | ")
}

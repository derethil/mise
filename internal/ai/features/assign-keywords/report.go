package assignkeywords

import "fmt"

type Change struct {
	Name     string
	Category string
	Reason   string
	Action   string
}

func (c Change) String() string {
	line := fmt.Sprintf("  [%-6s] %s", c.Action, c.Name)

	if c.Category != "" {
		line += fmt.Sprintf(" (%s)", c.Category)
	}
	if c.Reason != "" {
		line += fmt.Sprintf(" — %s", c.Reason)
	}

	return line
}

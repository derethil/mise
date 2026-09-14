package assignkeywords

import (
	"fmt"
	"strings"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	ActionKept    = "kept"
	ActionReused  = "reuse"
	ActionNew     = "new"
	ActionRemoved = "remove"
)

type ApplyOptions struct {
	Replace bool
	Protect []string
}

func (a AssignedKeywords) Apply(rawBefore []byte, vocabulary []tandoor.Keyword, opts ApplyOptions) ([]byte, []Change, error) {
	known := indexKnown(vocabulary)
	assigned := indexAssigned(a)
	protected := indexProtected(opts.Protect)

	kept, changes, onRecipe := reconcileExisting(rawBefore, assigned, protected, opts.Replace)

	newKept, newChanges, err := applyNew(a, known, onRecipe)
	if err != nil {
		return nil, nil, err
	}
	kept = append(kept, newKept...)
	changes = append(changes, newChanges...)

	out, err := sjson.SetRawBytes(rawBefore, "keywords", fmt.Appendf(nil, "[%s]", strings.Join(kept, ",")))
	if err != nil {
		return nil, nil, err
	}

	return out, changes, nil
}

func indexKnown(vocabulary []tandoor.Keyword) map[string]tandoor.Keyword {
	known := make(map[string]tandoor.Keyword, len(vocabulary))
	for _, keyword := range vocabulary {
		known[strings.ToLower(keyword.Name)] = keyword
	}
	return known
}

func indexAssigned(a AssignedKeywords) map[string]AssignedKeyword {
	assigned := make(map[string]AssignedKeyword, len(a))
	for _, keyword := range a {
		assigned[strings.ToLower(keyword.Name)] = keyword
	}
	return assigned
}

func indexProtected(protect []string) map[string]bool {
	protected := make(map[string]bool, len(protect))
	for _, name := range protect {
		protected[strings.ToLower(strings.TrimSpace(name))] = true
	}
	return protected
}

func reconcileExisting(rawBefore []byte, assigned map[string]AssignedKeyword, protected map[string]bool, replace bool) ([]string, []Change, map[string]bool) {
	var kept []string
	var changes []Change
	onRecipe := map[string]bool{}

	for _, existing := range gjson.GetBytes(rawBefore, "keywords").Array() {
		name := cleaned(existing.Get("name").String())
		key := strings.ToLower(name)
		onRecipe[key] = true

		if replace && !protected[key] {
			if _, ok := assigned[key]; !ok {
				changes = append(changes, Change{Name: name, Action: ActionRemoved})
				continue
			}
		}

		kept = append(kept, existing.Raw)
		changes = append(changes, Change{Name: name, Action: ActionKept})
	}

	return kept, changes, onRecipe
}

func applyNew(a AssignedKeywords, known map[string]tandoor.Keyword, onRecipe map[string]bool) ([]string, []Change, error) {
	var kept []string
	var changes []Change

	for _, keyword := range a {
		key := strings.ToLower(keyword.Name)
		if onRecipe[key] {
			continue
		}

		change := Change{Name: keyword.Name, Category: keyword.Category, Reason: keyword.Reason, Action: ActionNew}

		existing, isKnown := known[key]
		if isKnown {
			change.Name = existing.Name
			change.Action = ActionReused
		}

		raw, err := sjson.Set("{}", "name", change.Name)
		if err != nil {
			return nil, nil, err
		}
		if isKnown {
			if raw, err = sjson.Set(raw, "id", existing.ID); err != nil {
				return nil, nil, err
			}
		}

		kept = append(kept, raw)
		changes = append(changes, change)
	}

	return kept, changes, nil
}

func HasChanges(changes []Change) bool {
	for _, change := range changes {
		if change.Action != ActionKept {
			return true
		}
	}

	return false
}

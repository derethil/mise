package assignkeywords

import (
	"testing"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type ApplySuite struct {
	suite.Suite
}

func TestApplySuite(t *testing.T) {
	suite.Run(t, new(ApplySuite))
}

func (s *ApplySuite) recipe() []byte {
	return []byte(`{
		"id": 1,
		"keywords": [
			{"id": 5, "name": "Dinner"},
			{"id": 6, "name": "Quick"}
		]
	}`)
}

func (s *ApplySuite) TestApply_AddsNewKeyword() {
	assigned := AssignedKeywords{{Name: "Vegetarian", Category: "Diet", Reason: "no meat"}}

	out, changes, err := assigned.Apply(s.recipe(), nil, ApplyOptions{})
	s.Require().NoError(err)

	keywords := gjson.GetBytes(out, "keywords").Array()
	s.Require().Len(keywords, 3, "existing keywords are kept and the new one is appended")

	newKeyword := keywords[2]
	s.Equal("Vegetarian", newKeyword.Get("name").String())
	s.False(newKeyword.Get("id").Exists(), "a brand new keyword has no id yet")

	s.Require().Len(changes, 3)
	s.Equal(Change{Name: "Vegetarian", Category: "Diet", Reason: "no meat", Action: ActionNew}, changes[2])
}

func (s *ApplySuite) TestApply_ReusesExistingVocabularyKeywordByNameCaseInsensitive() {
	assigned := AssignedKeywords{{Name: "vegetarian", Category: "Diet"}}
	vocabulary := []tandoor.Keyword{{ID: 9, Name: "Vegetarian"}}

	out, changes, err := assigned.Apply(s.recipe(), vocabulary, ApplyOptions{})
	s.Require().NoError(err)

	added := gjson.GetBytes(out, "keywords").Array()[2]
	s.Equal("Vegetarian", added.Get("name").String(), "reused keyword uses the vocabulary's canonical spelling")
	s.Equal(int64(9), added.Get("id").Int())

	s.Equal(ActionReused, changes[2].Action)
	s.Equal("Vegetarian", changes[2].Name)
}

func (s *ApplySuite) TestApply_KeepsExistingKeywordsNotInAssignedSetByDefault() {
	assigned := AssignedKeywords{}

	out, changes, err := assigned.Apply(s.recipe(), nil, ApplyOptions{})
	s.Require().NoError(err)

	s.Len(gjson.GetBytes(out, "keywords").Array(), 2, "add-only mode never drops existing keywords")
	s.Equal([]Change{
		{Name: "Dinner", Action: ActionKept},
		{Name: "Quick", Action: ActionKept},
	}, changes)
}

func (s *ApplySuite) TestApply_ReplaceRemovesUnassignedKeywords() {
	assigned := AssignedKeywords{{Name: "Dinner"}}

	out, changes, err := assigned.Apply(s.recipe(), nil, ApplyOptions{Replace: true})
	s.Require().NoError(err)

	keywords := gjson.GetBytes(out, "keywords").Array()
	s.Require().Len(keywords, 1, "Quick should be dropped since it wasn't reassigned")
	s.Equal("Dinner", keywords[0].Get("name").String())

	s.Equal([]Change{
		{Name: "Dinner", Action: ActionKept},
		{Name: "Quick", Action: ActionRemoved},
	}, changes)
}

func (s *ApplySuite) TestApply_ReplaceProtectsIgnoredKeywordsFromRemoval() {
	assigned := AssignedKeywords{}

	out, changes, err := assigned.Apply(s.recipe(), nil, ApplyOptions{
		Replace: true,
		Protect: []string{"  Quick  "},
	})
	s.Require().NoError(err)

	keywords := gjson.GetBytes(out, "keywords").Array()
	s.Require().Len(keywords, 1, "Dinner is removed but the protected Quick keyword stays")
	s.Equal("Quick", keywords[0].Get("name").String())

	s.Equal([]Change{
		{Name: "Dinner", Action: ActionRemoved},
		{Name: "Quick", Action: ActionKept},
	}, changes)
}

func (s *ApplySuite) TestApply_AssignedKeywordAlreadyOnRecipeIsNotDuplicated() {
	assigned := AssignedKeywords{{Name: "dinner", Category: "Meal"}}

	out, changes, err := assigned.Apply(s.recipe(), nil, ApplyOptions{})
	s.Require().NoError(err)

	s.Len(gjson.GetBytes(out, "keywords").Array(), 2, "no duplicate Dinner entry should be added")
	s.Len(changes, 2)
	s.Equal(ActionKept, changes[0].Action, "the existing entry wins, not a new/reused one")
}

func (s *ApplySuite) TestApply_NoExistingKeywordsField() {
	assigned := AssignedKeywords{{Name: "Dessert"}}

	out, changes, err := assigned.Apply([]byte(`{"id": 1}`), nil, ApplyOptions{})
	s.Require().NoError(err)

	keywords := gjson.GetBytes(out, "keywords").Array()
	s.Require().Len(keywords, 1)
	s.Equal("Dessert", keywords[0].Get("name").String())
	s.Equal([]Change{{Name: "Dessert", Action: ActionNew}}, changes)
}

func (s *ApplySuite) TestHasChanges() {
	s.False(HasChanges(nil))
	s.False(HasChanges([]Change{{Action: ActionKept}, {Action: ActionKept}}))
	s.True(HasChanges([]Change{{Action: ActionKept}, {Action: ActionNew}}))
	s.True(HasChanges([]Change{{Action: ActionRemoved}}))
}

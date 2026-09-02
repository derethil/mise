package tandoor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
)

type Recipe struct {
	ID int `json:"id"`

	raw    []byte
	loaded bool
	client *Client
}

func NewRecipe(client *Client, id int) Recipe {
	return Recipe{
		ID:     id,
		raw:    nil,
		loaded: false,
		client: client,
	}
}

func (r *Recipe) Load() error {
	body, err := r.client.Request(http.MethodGet, fmt.Sprintf("recipe/%d/", r.ID), nil)
	if err != nil {
		return err
	}

	parsed, err := parseRecipe(body)
	if err != nil {
		return err
	}

	r.ID = parsed.ID
	r.raw = parsed.raw
	r.loaded = true

	return nil
}

func (r Recipe) JSON() []byte {
	return r.raw
}

func (r Recipe) Backup() error {
	if err := os.MkdirAll(r.client.backupDir, 0o755); err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, r.JSON(), "", "    "); err != nil {
		return err
	}

	return os.WriteFile(r.backupPath(), pretty.Bytes(), 0o644)
}

func (r Recipe) Restore() error {
	data, err := os.ReadFile(r.backupPath())
	if err != nil {
		return err
	}

	parsed, err := parseRecipe(data)
	if err != nil {
		return fmt.Errorf("invalid backup file for recipe %d: %w", r.ID, err)
	}

	_, err = r.client.Request(http.MethodPut, fmt.Sprintf("recipe/%d/", r.ID), parsed.JSON())
	return err
}

func (r Recipe) backupPath() string {
	return filepath.Join(r.client.backupDir, fmt.Sprintf("%d.json", r.ID))
}

func parseRecipe(raw []byte) (Recipe, error) {
	if !gjson.ValidBytes(raw) {
		return Recipe{}, fmt.Errorf("invalid recipe JSON")
	}

	id := gjson.GetBytes(raw, "id")
	if !id.Exists() {
		return Recipe{}, fmt.Errorf("recipe is missing an id")
	}

	return Recipe{ID: int(id.Int()), raw: raw}, nil
}

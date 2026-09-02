// Package tandoor implements a minimal client around the Tandoor API.
package tandoor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Client struct {
	baseURL    string
	token      string
	backupDir  string
	httpClient *http.Client
}

func NewClient(baseURL, token, backupDir string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		backupDir:  backupDir,
		httpClient: &http.Client{},
	}
}

func (c *Client) GetRecipeByID(id int) (Recipe, error) {
	body, err := c.request(http.MethodGet, fmt.Sprintf("recipe/%d/", id), nil)
	if err != nil {
		return Recipe{}, err
	}

	return ParseRecipe(body)
}

func (c *Client) BackupRecipe(recipe Recipe) error {
	if err := os.MkdirAll(c.backupDir, 0o755); err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, recipe.JSON(), "", "    "); err != nil {
		return err
	}

	return os.WriteFile(c.recipeBackupPath(recipe.ID), pretty.Bytes(), 0o644)
}

func (c *Client) RestoreRecipe(recipeID int) error {
	data, err := os.ReadFile(c.recipeBackupPath(recipeID))
	if err != nil {
		return err
	}

	recipe, err := ParseRecipe(data)
	if err != nil {
		return fmt.Errorf("invalid backup file for recipe %d: %w", recipeID, err)
	}

	_, err = c.request(http.MethodPut, fmt.Sprintf("recipe/%d/", recipeID), recipe.JSON())
	return err
}

func (c *Client) recipeBackupPath(recipeID int) string {
	return filepath.Join(c.backupDir, fmt.Sprintf("%d.json", recipeID))
}

func (c *Client) request(method, endpoint string, payload []byte) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	url := fmt.Sprintf("%s/api/%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request to %s failed: %s", endpoint, resp.Status)
	}

	return respBody, nil
}

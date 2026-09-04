// Package tandoor implements a minimal client around the Tandoor API.
package tandoor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/derethil/mise/internal/config"
)

const defaultTimeout = 30 * time.Second

type Client struct {
	baseURL    string
	token      string
	timeout    time.Duration
	httpClient *http.Client

	Recipes *RecipeService
}

func NewClient(baseURL, token string) *Client {
	c := &Client{
		baseURL:    baseURL,
		token:      token,
		timeout:    defaultTimeout,
		httpClient: &http.Client{},
	}

	c.Recipes = &RecipeService{client: c}

	return c
}

func FromConfig(cfg config.Config) *Client {
	return NewClient(cfg.Tandoor.BaseURL, cfg.Tandoor.Token)
}

func (c *Client) Request(ctx context.Context, method, endpoint string, payload []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
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
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request to %s failed: %s", endpoint, resp.Status)
	}

	return respBody, nil
}

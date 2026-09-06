// Package tandoor implements a minimal client around the Tandoor API.
package tandoor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
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

	slog.DebugContext(ctx, "sending tandoor request", slog.String("method", method), slog.String("endpoint", endpoint))

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

	slog.DebugContext(ctx, "received tandoor response", slog.String("endpoint", endpoint), slog.Int("status", resp.StatusCode))

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("request to %s failed: %s", endpoint, resp.Status)

		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("%w: %w", ErrTandoorUnauthorized, err)
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w: %w", ErrTandoorNotFound, err)
		default:
			return nil, fmt.Errorf("%w: %w", ErrTandoorRequestFailed, err)
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf("%w: %s returned %s instead of JSON", ErrTandoorRequestFailed, resp.Request.URL, contentType)
	}

	return respBody, nil
}

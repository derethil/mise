// Package tandoor implements a minimal client around the Tandoor API.
package tandoor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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
	Foods   *FoodService
	Units   *UnitService
}

func NewClient(baseURL, token string) *Client {
	c := &Client{
		baseURL:    baseURL,
		token:      token,
		timeout:    defaultTimeout,
		httpClient: &http.Client{},
	}

	c.Recipes = &RecipeService{client: c}
	c.Foods = &FoodService{client: c}
	c.Units = &UnitService{client: c}

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

	slog.DebugContext(ctx, "sending tandoor request",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
		slog.String("request_body", string(payload)),
	)

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

	slog.DebugContext(ctx, "received tandoor response",
		slog.String("endpoint", endpoint),
		slog.Int("status", resp.StatusCode),
		slog.String("response_body", string(respBody)),
	)

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

type paginatedResponse[T any] struct {
	Results []T     `json:"results"`
	Next    *string `json:"next"`
}

func RequestAllPages[T any](ctx context.Context, c *Client, endpoint string) ([]T, error) {
	var all []T

	for endpoint != "" {
		body, err := c.Request(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}

		var page paginatedResponse[T]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("invalid paginated response from %s: %w", endpoint, err)
		}
		all = append(all, page.Results...)

		endpoint = ""
		if page.Next != nil {
			endpoint = strings.TrimPrefix(*page.Next, c.baseURL+"/")
		}
	}

	return all, nil
}

func constructURL(endpoint string, params map[string]string) string {
	qParams := url.Values{}

	for k, v := range params {
		qParams.Set(k, v)
	}

	if len(qParams) == 0 {
		return endpoint
	}

	return fmt.Sprintf("%s?%s", endpoint, qParams.Encode())
}

func constructParams(defaults map[string]string, kv ...string) map[string]string {
	for i := 0; i < len(kv)-1; i += 2 {
		defaults[kv[i]] = kv[i+1]
	}

	return defaults
}

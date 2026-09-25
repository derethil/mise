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

var skipResponseLogEndpoints = []string{"/openapi/"}

type Client struct {
	baseURL    string
	rootURL    string
	token      string
	timeout    time.Duration
	httpClient *http.Client

	Recipes  *RecipeService
	Foods    *FoodService
	Units    *UnitService
	Keywords *KeywordService
}

func NewClient(baseURL, token string) *Client {
	rootURL := strings.TrimRight(baseURL, "/")

	c := &Client{
		baseURL:    rootURL + "/api",
		rootURL:    rootURL,
		token:      token,
		timeout:    defaultTimeout,
		httpClient: &http.Client{},
	}

	c.Recipes = &RecipeService{client: c}
	c.Foods = &FoodService{ResourceService[Food]{client: c, endpoint: "food/"}}
	c.Units = &UnitService{ResourceService[Unit]{client: c, endpoint: "unit/"}}
	c.Keywords = &KeywordService{ResourceService[Keyword]{client: c, endpoint: "keyword/"}}

	return c
}

func FromConfig(cfg config.Config) *Client {
	return NewClient(cfg.Tandoor.BaseURL, cfg.Tandoor.Token)
}

func (c *Client) Request(ctx context.Context, method, endpoint string, payload []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	return c.requestURL(ctx, method, url, payload)
}

func (c *Client) requestURL(ctx context.Context, method, url string, payload []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

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
		slog.String("url", url),
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

	if shouldLogResponseBody(url) {
		slog.DebugContext(ctx, "received tandoor response",
			slog.String("url", url),
			slog.Int("status", resp.StatusCode),
			slog.String("response_body", string(respBody)),
		)
	} else {
		slog.DebugContext(ctx, "received tandoor response",
			slog.String("url", url),
			slog.Int("status", resp.StatusCode),
		)
	}

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("request to %s failed: %s", url, resp.Status)

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
	if contentType != "" && !strings.Contains(contentType, "json") {
		return nil, fmt.Errorf("%w: %s returned %s instead of JSON", ErrTandoorRequestFailed, resp.Request.URL, contentType)
	}

	return respBody, nil
}

func shouldLogResponseBody(url string) bool {
	for _, endpoint := range skipResponseLogEndpoints {
		if strings.Contains(url, endpoint) {
			return false
		}
	}

	return true
}

func (c *Client) TestConnection(ctx context.Context) error {
	endpoint := constructURL("meal-type/", map[string]string{"page_size": "1"})
	_, err := c.Request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	return nil
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

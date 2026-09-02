// Package tandoor implements a minimal client around the Tandoor API.
package tandoor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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

func (c *Client) Request(method, endpoint string, payload []byte) ([]byte, error) {
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

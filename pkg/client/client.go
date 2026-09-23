package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	u := strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL:    u,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) doJSON(method, path string, body any, dest any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s %s -> %s: %s", method, path, res.Status, raw)
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

func (c *Client) Health() error {
	var out map[string]any
	if err := c.doJSON(http.MethodGet, "/health", nil, &out); err != nil {
		return err
	}
	return nil
}

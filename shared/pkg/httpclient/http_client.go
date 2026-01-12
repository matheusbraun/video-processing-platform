package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is an interface for HTTP operations with retry logic
type HTTPClient interface {
	Get(ctx context.Context, url string, response interface{}) error
	Post(ctx context.Context, url string, body, response interface{}) error
	Put(ctx context.Context, url string, body, response interface{}) error
	Delete(ctx context.Context, url string, response interface{}) error
}

// Client implements HTTPClient with retry logic
type Client struct {
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// NewHTTPClient creates a new HTTP client with retry logic
func NewHTTPClient(timeout time.Duration, maxRetries int) HTTPClient {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		retryDelay: 1 * time.Second,
	}
}

// Get performs a GET request with retry logic
func (c *Client) Get(ctx context.Context, url string, response interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.doWithRetry(req, response)
}

// Post performs a POST request with retry logic
func (c *Client) Post(ctx context.Context, url string, body, response interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doWithRetry(req, response)
}

// Put performs a PUT request with retry logic
func (c *Client) Put(ctx context.Context, url string, body, response interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doWithRetry(req, response)
}

// Delete performs a DELETE request with retry logic
func (c *Client) Delete(ctx context.Context, url string, response interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.doWithRetry(req, response)
}

// doWithRetry executes an HTTP request with retry logic
func (c *Client) doWithRetry(req *http.Request, response interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := c.retryDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		// Read response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Check status code
		if resp.StatusCode >= 500 {
			// Retry on server errors
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			// Don't retry on client errors
			return fmt.Errorf("client error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		// Success - unmarshal response
		if response != nil && len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, response); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

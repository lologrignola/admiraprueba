package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	httpClient *http.Client
	logger     *logrus.Logger
	maxRetries int
	retryDelay time.Duration
}

type ClientConfig struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

func NewClient(config ClientConfig, logger *logrus.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:     logger,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
	}
}

func (c *Client) Get(ctx context.Context, url string, result interface{}) error {
	return c.doWithRetry(ctx, "GET", url, nil, result)
}

func (c *Client) Post(ctx context.Context, url string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	return c.doWithRetry(ctx, "POST", url, jsonBody, result)
}

func (c *Client) doWithRetry(ctx context.Context, method, url string, body []byte, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
				// Exponential backoff
			}
		}

		err := c.doRequest(ctx, method, url, body, result)
		if err == nil {
			return nil
		}

		lastErr = err
		c.logger.WithFields(logrus.Fields{
			"attempt": attempt + 1,
			"url":     url,
			"method":  method,
			"error":   err.Error(),
		}).Warn("Request failed, retrying")

		// Don't retry on client errors (4xx)
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return err
		}
	}

	return fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

func (c *Client) doRequest(ctx context.Context, method, url string, body []byte, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}


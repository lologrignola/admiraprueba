package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Get(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		Timeout:    5 * time.Second,
		MaxRetries: 2,
		RetryDelay: 100 * time.Millisecond,
	}, logger)

	var result map[string]string
	err := client.Get(context.Background(), server.URL, &result)

	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestClient_GetWithRetry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 50 * time.Millisecond,
	}, logger)

	var result map[string]string
	err := client.Get(context.Background(), server.URL, &result)

	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
	assert.Equal(t, 2, attempts)
}

func TestClient_GetClientError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 50 * time.Millisecond,
	}, logger)

	var result map[string]string
	err := client.Get(context.Background(), server.URL, &result)

	require.Error(t, err)
	httpErr, ok := err.(*HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
}

func TestClient_Post(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received": true}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		Timeout:    5 * time.Second,
		MaxRetries: 2,
		RetryDelay: 100 * time.Millisecond,
	}, logger)

	requestBody := map[string]string{"test": "data"}
	var result map[string]bool
	err := client.Post(context.Background(), server.URL, requestBody, &result)

	require.NoError(t, err)
	assert.True(t, result["received"])
}

func TestHTTPError(t *testing.T) {
	err := &HTTPError{
		StatusCode: 404,
		Message:    "Not Found",
	}

	assert.Equal(t, "HTTP 404: Not Found", err.Error())
}


// Package testutil provides testing utilities for DEparrow integration tests.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPClient provides helper methods for HTTP requests in tests.
type HTTPClient struct {
	BaseURL   string
	Token     string
	Client    *http.Client
	UserAgent string
}

// NewHTTPClient creates a new HTTP client for testing.
func NewHTTPClient(baseURL, token string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Token:   token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: "DEparrow-Test/1.0",
	}
}

// Get performs a GET request.
func (c *HTTPClient) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	return c.Client.Do(req)
}

// Post performs a POST request with JSON body.
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(jsonBody))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// Put performs a PUT request with JSON body.
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(jsonBody))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// Delete performs a DELETE request.
func (c *HTTPClient) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	return c.Client.Do(req)
}

func (c *HTTPClient) setHeaders(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
}

// ReadJSON reads JSON response body into target.
func ReadJSON(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// ReadString reads response body as string.
func ReadString(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// AssertSuccessfulResponse asserts that the response is successful (2xx).
func AssertSuccessfulResponse(t *testing.T, resp *http.Response) {
	t.Helper()
	assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
		"Expected successful status code, got %d", resp.StatusCode)
}

// AssertErrorResponse asserts that the response is an error.
func AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, resp.StatusCode,
		"Expected status code %d, got %d", expectedStatus, resp.StatusCode)
}

// WaitForHealth waits for the server to become healthy.
func WaitForHealth(ctx context.Context, client *HTTPClient) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server health: %w", ctx.Err())
		default:
			resp, err := client.Get(ctx, "/api/v1/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Retry retries a function until it succeeds or timeout.
func Retry(ctx context.Context, fn func() error, interval time.Duration) error {
	var lastErr error
	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("retry failed: %w (last error: %v)", ctx.Err(), lastErr)
			}
			return ctx.Err()
		default:
			if err := fn(); err != nil {
				lastErr = err
				time.Sleep(interval)
				continue
			}
			return nil
		}
	}
}

// Eventually asserts that a condition becomes true within a timeout.
func Eventually(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	require.Eventually(t, condition, timeout, interval, msgAndArgs...)
}

// AssertStatusCode asserts the HTTP status code.
func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	assert.Equal(t, expected, actual, "Status code mismatch")
}

// AssertJSONContains asserts that JSON response contains expected keys.
func AssertJSONContains(t *testing.T, body []byte, keys ...string) {
	t.Helper()
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse JSON response")

	for _, key := range keys {
		_, exists := result[key]
		assert.True(t, exists, "JSON response missing key: %s", key)
	}
}

// AssertJSONEquals asserts that JSON response equals expected value.
func AssertJSONEquals(t *testing.T, body []byte, key string, expected interface{}) {
	t.Helper()
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse JSON response")

	actual, exists := result[key]
	require.True(t, exists, "JSON response missing key: %s", key)
	assert.Equal(t, expected, actual, "JSON value mismatch for key: %s", key)
}

// TestLogger provides logging for tests.
type TestLogger struct {
	t      *testing.T
	prefix string
}

// NewTestLogger creates a new test logger.
func NewTestLogger(t *testing.T, prefix string) *TestLogger {
	return &TestLogger{t: t, prefix: prefix}
}

// Log logs a message.
func (l *TestLogger) Log(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Logf("[%s] %s", l.prefix, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *TestLogger) Error(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Errorf("[%s] %s", l.prefix, fmt.Sprintf(format, args...))
}

// Debug logs a debug message (only in verbose mode).
func (l *TestLogger) Debug(format string, args ...interface{}) {
	l.t.Helper()
	if testing.Verbose() {
		l.t.Logf("[%s] DEBUG: %s", l.prefix, fmt.Sprintf(format, args...))
	}
}

// Context provides a test context with cleanup.
type TestContext struct {
	context.Context
	cancel context.CancelFunc
	cleanup []func()
	t      *testing.T
}

// NewTestContext creates a new test context.
func NewTestContext(t *testing.T, timeout time.Duration) *TestContext {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &TestContext{
		Context: ctx,
		cancel:  cancel,
		cleanup: make([]func(), 0),
		t:       t,
	}
}

// AddCleanup adds a cleanup function.
func (tc *TestContext) AddCleanup(fn func()) {
	tc.cleanup = append(tc.cleanup, fn)
}

// Cleanup runs all cleanup functions.
func (tc *TestContext) Cleanup() {
	tc.cancel()
	// Run cleanup in reverse order
	for i := len(tc.cleanup) - 1; i >= 0; i-- {
		tc.cleanup[i]()
	}
}

// WaitForCondition waits for a condition to be true with timeout.
func (tc *TestContext) WaitForCondition(condition func() bool, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-tc.Done():
			return tc.Err()
		case <-ticker.C:
			if condition() {
				return nil
			}
		}
	}
}

// MockWebSocketConn provides a mock WebSocket connection for testing.
type MockWebSocketConn struct {
	messages [][]byte
	closed   bool
}

// NewMockWebSocketConn creates a new mock WebSocket connection.
func NewMockWebSocketConn() *MockWebSocketConn {
	return &MockWebSocketConn{
		messages: make([][]byte, 0),
		closed:   false,
	}
}

// SendMessage simulates receiving a message.
func (m *MockWebSocketConn) SendMessage(msg []byte) {
	m.messages = append(m.messages, msg)
}

// Close closes the connection.
func (m *MockWebSocketConn) Close() error {
	m.closed = true
	return nil
}

// IsClosed returns whether the connection is closed.
func (m *MockWebSocketConn) IsClosed() bool {
	return m.closed
}

// GenerateTestID generates a unique test ID.
func GenerateTestID() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}

// RandomPort generates a random port number for testing.
func RandomPort() int {
	return 10000 + int(time.Now().UnixNano()%55000)
}

// CompareJSON compares two JSON byte slices for equality.
func CompareJSON(t *testing.T, expected, actual []byte) {
	t.Helper()
	var expectedJSON, actualJSON interface{}
	require.NoError(t, json.Unmarshal(expected, &expectedJSON))
	require.NoError(t, json.Unmarshal(actual, &actualJSON))
	assert.Equal(t, expectedJSON, actualJSON)
}

// MustMarshal marshals to JSON or panics.
func MustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// MustUnmarshal unmarshals from JSON or panics.
func MustUnmarshal(data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}
}

// SkipIfShort skips the test if -short flag is set.
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
}

// SkipIfCI skips the test if running in CI.
func SkipIfCI(t *testing.T) {
	t.Helper()
	if IsCI() {
		t.Skip("Skipping in CI environment")
	}
}

// IsCI returns true if running in CI environment.
func IsCI() bool {
	// Check common CI environment variables
	return strings.EqualFold(os.Getenv("CI"), "true") ||
		strings.EqualFold(os.Getenv("GITHUB_ACTIONS"), "true") ||
		strings.EqualFold(os.Getenv("GITLAB_CI"), "true") ||
		strings.EqualFold(os.Getenv("TRAVIS"), "true") ||
		strings.EqualFold(os.Getenv("CIRCLECI"), "true")
}

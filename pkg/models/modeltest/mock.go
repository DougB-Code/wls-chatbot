// mock.go implements HTTP transport that replays golden file responses.
package modeltest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MockTransport replays responses from golden files.
type MockTransport struct {
	index    *GoldenIndex
	provider string
	model    string
}

// NewMockTransport creates a transport that serves responses from golden files.
func NewMockTransport(index *GoldenIndex, provider, model string) *MockTransport {
	return &MockTransport{
		index:    index,
		provider: provider,
		model:    model,
	}
}

// RoundTrip implements http.RoundTripper by matching requests to golden files.
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	capability := t.inferCapability(req)
	if capability == "" {
		return nil, fmt.Errorf("could not infer capability from request: %s %s", req.Method, req.URL.Path)
	}

	golden := t.index.Lookup(t.provider, capability, t.model)
	if golden == nil {
		return nil, fmt.Errorf("no golden file for %s/%s/%s", t.provider, capability, t.model)
	}

	// Build response from golden file
	resp := &http.Response{
		Status:     fmt.Sprintf("%d %s", golden.Response.Status, http.StatusText(golden.Response.Status)),
		StatusCode: golden.Response.Status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(golden.Response.Body)),
		Request:    req,
	}

	for k, v := range golden.Response.Headers {
		resp.Header.Set(k, v)
	}

	return resp, nil
}

// inferCapability determines the capability from the request URL.
func (t *MockTransport) inferCapability(req *http.Request) string {
	path := req.URL.Path

	// OpenAI-style endpoints
	if strings.Contains(path, "/chat/completions") {
		return "chat"
	}
	if strings.Contains(path, "/images/generations") {
		return "image_gen"
	}
	if strings.Contains(path, "/images/edits") {
		return "image_edit"
	}
	if strings.Contains(path, "/models") {
		return "test_connection"
	}

	// Gemini-style endpoints
	if strings.Contains(path, ":generateContent") {
		return "chat"
	}
	if strings.Contains(path, "imagen") || strings.Contains(path, ":predict") {
		return "image_gen"
	}

	// Anthropic-style endpoints
	if strings.Contains(path, "/messages") {
		return "chat"
	}

	return ""
}

// CapabilityMatcher allows custom capability matching rules.
type CapabilityMatcher interface {
	Match(req *http.Request) string
}

// RequestMatcher compares incoming requests to golden file requests.
type RequestMatcher struct {
	IgnoreHeaders []string // Headers to ignore when matching
	IgnoreFields  []string // JSON fields to ignore in body
}

// DefaultRequestMatcher returns a matcher with sensible defaults.
func DefaultRequestMatcher() *RequestMatcher {
	return &RequestMatcher{
		IgnoreHeaders: []string{"Authorization", "X-Api-Key", "Date", "User-Agent"},
		IgnoreFields:  []string{"user", "stream"},
	}
}

// Matches checks if an incoming request matches a golden file request.
func (m *RequestMatcher) Matches(req *http.Request, golden RecordedRequest) bool {
	// Method must match
	if req.Method != golden.Method {
		return false
	}

	// Path must match (ignoring query params for now)
	reqPath := req.URL.Path
	goldenPath := strings.Split(golden.URL, "?")[0]
	if !strings.HasSuffix(goldenPath, reqPath) && !strings.HasSuffix(reqPath, goldenPath) {
		return false
	}

	return true
}

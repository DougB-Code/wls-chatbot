// recorder.go implements HTTP round-trip recording for golden file capture.
package modeltest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// RecordedRequest contains a captured HTTP request.
type RecordedRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body,omitempty"`
}

// RecordedResponse contains a captured HTTP response.
type RecordedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body,omitempty"`
}

// Recording represents a single captured request/response pair.
type Recording struct {
	Timestamp time.Time        `json:"timestamp"`
	Request   RecordedRequest  `json:"request"`
	Response  RecordedResponse `json:"response"`
	Duration  time.Duration    `json:"duration_ms"`
	Error     string           `json:"error,omitempty"`
}

// RecordingTransport wraps an http.RoundTripper to capture request/response pairs.
type RecordingTransport struct {
	underlying http.RoundTripper
	recordings []Recording
	mu         sync.Mutex

	// Patterns to sanitize in headers and bodies
	sensitivePatterns []*regexp.Regexp
}

// NewRecordingTransport creates a transport that records HTTP exchanges.
func NewRecordingTransport(underlying http.RoundTripper) *RecordingTransport {
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	return &RecordingTransport{
		underlying: underlying,
		sensitivePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(api[_-]?key|authorization|x-api-key|bearer)\s*[:=]\s*["']?([^"'\s,}]+)`),
			regexp.MustCompile(`(?i)key=([^&\s]+)`),
		},
	}
}

// RoundTrip implements http.RoundTripper.
func (t *RecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	rec := Recording{
		Timestamp: start,
		Request:   t.captureRequest(req),
	}

	resp, err := t.underlying.RoundTrip(req)
	rec.Duration = time.Since(start)

	if err != nil {
		rec.Error = err.Error()
	} else {
		rec.Response = t.captureResponse(resp)
	}

	t.mu.Lock()
	t.recordings = append(t.recordings, rec)
	t.mu.Unlock()

	return resp, err
}

// Recordings returns all captured recordings.
func (t *RecordingTransport) Recordings() []Recording {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]Recording, len(t.recordings))
	copy(result, t.recordings)
	return result
}

// Clear removes all recordings.
func (t *RecordingTransport) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.recordings = nil
}

// captureRequest extracts request data for recording.
func (t *RecordingTransport) captureRequest(req *http.Request) RecordedRequest {
	rec := RecordedRequest{
		Method:  req.Method,
		URL:     t.sanitize(req.URL.String()),
		Headers: make(map[string]string),
	}

	// Capture and sanitize headers
	for k, v := range req.Header {
		if len(v) > 0 {
			rec.Headers[k] = t.sanitize(v[0])
		}
	}

	// Capture body if present
	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // Reset body
			sanitized := t.sanitize(string(bodyBytes))
			rec.Body = json.RawMessage(sanitized)
		}
	}

	return rec
}

// captureResponse extracts response data for recording.
func (t *RecordingTransport) captureResponse(resp *http.Response) RecordedResponse {
	rec := RecordedResponse{
		Status:  resp.StatusCode,
		Headers: make(map[string]string),
	}

	for k, v := range resp.Header {
		if len(v) > 0 {
			rec.Headers[k] = v[0]
		}
	}

	// Capture body
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // Reset body
			rec.Body = json.RawMessage(bodyBytes)
		}
	}

	return rec
}

// sanitize removes sensitive data from a string.
func (t *RecordingTransport) sanitize(s string) string {
	result := s
	for _, pattern := range t.sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Extract the prefix and replace the sensitive value
			parts := strings.SplitN(match, "=", 2)
			if len(parts) == 2 {
				return parts[0] + "=REDACTED"
			}
			parts = strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ": REDACTED"
			}
			return "REDACTED"
		})
	}
	return result
}

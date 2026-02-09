// stream_manager.go manages cancellation state for streaming chat responses.
// internal/features/ai/chat/app/chat/stream_manager.go
package chat

import (
	"context"
	"sync"
)

// streamManager controls the lifecycle of a single active stream.
type streamManager struct {
	mu             sync.Mutex
	cancel         context.CancelFunc
	conversationID string
	messageID      string
	cancelled      bool
}

// newStreamManager constructs a stream manager.
func newStreamManager() *streamManager {

	return &streamManager{}
}

// start records the active stream and stores its cancel function.
func (s *streamManager) start(conversationID, messageID string, cancel context.CancelFunc) {

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.cancel = cancel
	s.conversationID = conversationID
	s.messageID = messageID
	s.cancelled = false
}

// stop cancels the active stream if present.
func (s *streamManager) stop() {

	s.mu.Lock()
	if s.messageID != "" {
		s.cancelled = true
	}
	cancel := s.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// wasCancelled reports whether the active stream was cancelled.
func (s *streamManager) wasCancelled(conversationID, messageID string) bool {

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.messageID == "" {
		return false
	}
	if s.conversationID != conversationID || s.messageID != messageID {
		return false
	}
	return s.cancelled
}

// clear removes the active stream if it matches the given identifiers.
func (s *streamManager) clear(conversationID, messageID string) {

	var cancel context.CancelFunc
	s.mu.Lock()
	if s.conversationID == conversationID && s.messageID == messageID {
		cancel = s.cancel
		s.cancel = nil
		s.conversationID = ""
		s.messageID = ""
		s.cancelled = false
	}
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

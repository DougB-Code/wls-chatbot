// stream_manager.go manages cancellation state for streaming chat responses.
// internal/features/chat/app/chat/stream_manager.go
package chat

import (
	"context"
	"sync"
	"time"
)

// streamInfo captures metadata about the active stream.
type streamInfo struct {
	conversationID string
	messageID      string
	startedAt      time.Time
	cancelled      bool
}

// streamManager controls the lifecycle of a single active stream.
type streamManager struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	info   *streamInfo
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
	s.info = &streamInfo{
		conversationID: conversationID,
		messageID:      messageID,
		startedAt:      time.Now(),
	}
}

// stop cancels the active stream if present.
func (s *streamManager) stop() {

	s.mu.Lock()
	if s.info != nil {
		s.info.cancelled = true
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
	if s.info == nil {
		return false
	}
	if s.info.conversationID != conversationID || s.info.messageID != messageID {
		return false
	}
	return s.info.cancelled
}

// clear removes the active stream if it matches the given identifiers.
func (s *streamManager) clear(conversationID, messageID string) {

	var cancel context.CancelFunc
	s.mu.Lock()
	if s.info != nil && s.info.conversationID == conversationID && s.info.messageID == messageID {
		cancel = s.cancel
		s.cancel = nil
		s.info = nil
	}
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

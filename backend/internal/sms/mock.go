package sms

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockProvider implements SMSProvider for testing
type MockProvider struct {
	mu            sync.RWMutex
	sentMessages  []MockMessage
	shouldFail    bool
	failureReason error
}

// MockMessage represents a message sent via the mock provider
type MockMessage struct {
	ID        string
	To        string
	Message   string
	Timestamp time.Time
}

// NewMockProvider creates a new mock SMS provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		sentMessages: make([]MockMessage, 0),
		shouldFail:   false,
	}
}

// SendSMS sends a mock SMS (just logs it)
func (m *MockProvider) SendSMS(ctx context.Context, to, message string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		if m.failureReason != nil {
			return "", m.failureReason
		}
		return "", ErrSendFailed
	}

	messageID := uuid.New().String()
	msg := MockMessage{
		ID:        messageID,
		To:        to,
		Message:   message,
		Timestamp: time.Now(),
	}

	m.sentMessages = append(m.sentMessages, msg)

	// Log for debugging in test/dev environments
	log.Printf("[MOCK SMS] To: %s, Message: %s, ID: %s", to, message, messageID)

	return messageID, nil
}

// GetProviderName returns the provider name
func (m *MockProvider) GetProviderName() string {
	return string(ProviderMock)
}

// ValidateConfig validates the mock configuration (always valid)
func (m *MockProvider) ValidateConfig() error {
	return nil
}

// GetSentMessages returns all messages sent via this provider
func (m *MockProvider) GetSentMessages() []MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]MockMessage, len(m.sentMessages))
	copy(messages, m.sentMessages)
	return messages
}

// GetLastMessage returns the last sent message
func (m *MockProvider) GetLastMessage() *MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.sentMessages) == 0 {
		return nil
	}
	return &m.sentMessages[len(m.sentMessages)-1]
}

// Clear clears all sent messages
func (m *MockProvider) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = make([]MockMessage, 0)
}

// SetShouldFail sets whether the provider should fail
func (m *MockProvider) SetShouldFail(shouldFail bool, reason error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failureReason = reason
}

// GetMessageByID finds a message by its ID
func (m *MockProvider) GetMessageByID(id string) *MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := range m.sentMessages {
		if m.sentMessages[i].ID == id {
			return &m.sentMessages[i]
		}
	}
	return nil
}

// GetMessageCount returns the number of messages sent
func (m *MockProvider) GetMessageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sentMessages)
}

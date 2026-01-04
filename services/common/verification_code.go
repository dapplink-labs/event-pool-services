package common

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type VerificationCodeEntry struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Type      string    `json:"type"`
}

type VerificationCodeManager struct {
	codes   map[string]*VerificationCodeEntry
	mutex   sync.RWMutex
	timeout time.Duration
}

func NewVerificationCodeManager() *VerificationCodeManager {
	manager := &VerificationCodeManager{
		codes:   make(map[string]*VerificationCodeEntry),
		timeout: 60 * time.Second,
	}
	go manager.cleanupExpiredCodes()

	return manager
}

func (m *VerificationCodeManager) GenerateCode() string {
	min := big.NewInt(100000)
	max := big.NewInt(999999)

	// max - min + 1
	diff := new(big.Int).Sub(max, min)
	diff.Add(diff, big.NewInt(1))

	n, err := rand.Int(rand.Reader, diff)
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}

	result := new(big.Int).Add(n, min)
	return result.String()
}

func (m *VerificationCodeManager) SendCode(target, codeType string) (string, error) {
	code := m.GenerateCode()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.codes[target] = &VerificationCodeEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(m.timeout),
		Type:      codeType,
	}

	return code, nil
}

func (m *VerificationCodeManager) VerifyCode(target, inputCode string) (bool, error) {
	m.mutex.RLock()
	entry, exists := m.codes[target]
	m.mutex.RUnlock()

	if !exists {
		return false, fmt.Errorf("verification code not found")
	}

	if time.Now().After(entry.ExpiresAt) {
		m.mutex.Lock()
		delete(m.codes, target)
		m.mutex.Unlock()
		return false, fmt.Errorf("verification code expired")
	}

	if entry.Code == inputCode {
		m.mutex.Lock()
		delete(m.codes, target)
		m.mutex.Unlock()
		return true, nil
	}

	return false, fmt.Errorf("invalid verification code")
}

func (m *VerificationCodeManager) GetRemainingTime(target string) int {
	m.mutex.RLock()
	entry, exists := m.codes[target]
	m.mutex.RUnlock()

	if !exists {
		return 0
	}

	remaining := entry.ExpiresAt.Sub(time.Now()).Seconds()
	if remaining <= 0 {
		return 0
	}

	return int(remaining)
}

func (m *VerificationCodeManager) cleanupExpiredCodes() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()
		for target, entry := range m.codes {
			if now.After(entry.ExpiresAt) {
				delete(m.codes, target)
			}
		}
		m.mutex.Unlock()
	}
}

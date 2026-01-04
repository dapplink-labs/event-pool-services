package backend

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateCompactUUID generates a 32-character UUID without dashes
func GenerateCompactUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// CurrentTimestamp returns the current Unix timestamp in seconds
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}

package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateServiceID generates a unique service ID
func GenerateServiceID() string {
	// Use timestamp + random number for uniqueness
	timestamp := time.Now().UnixNano()
	random := rand.Intn(10000)
	return fmt.Sprintf("svc_%d%d", timestamp, random)
}


package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Deal eveyting in bytes to keep things generic and simple
func SHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Hash/secret Verification
func VerifyHash(data []byte, expected string) bool {
	return SHA256Hash(data) == expected
}

func HashCertificate(cert []byte) string {
	return SHA256Hash(cert)
}

func VerifySecret(secret string, expected string) bool {
	return VerifyHash([]byte(secret), expected)
}

// CreateCompositeKey creates a deterministic composite key
// Useful for creating unique identifiers
func CreateCompositeKey(components ...string) string {
	var combined string
	for i, component := range components {
		if i > 0 {
			combined += ":"
		}
		combined += component
	}
	return SHA256Hash([]byte(combined))
}

func LogHashOperation(operation string, input string, output string) {
	fmt.Printf("[CRYPTO] %s: %s -> %s\n", operation, input[:min(len(input), 10)]+"...", output[:16]+"...")
}

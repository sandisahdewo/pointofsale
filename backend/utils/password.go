package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64MB
	argonThreads = 4
	argonKeyLen  = 32
	saltLength   = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}

func VerifyPassword(encodedHash, password string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var memory uint32
	var time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1, nil
}

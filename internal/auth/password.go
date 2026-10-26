package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength = 16
	keyLength  = 32

	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, fmt.Errorf("invalid password hash format")
	}

	var parsedMemory, parsedIterations, parsedParallelism uint64

	for _, parameter := range strings.Split(parts[3], ",") {
		keyValue := strings.SplitN(parameter, "=", 2)
		if len(keyValue) != 2 {
			return false, fmt.Errorf("invalid Argon2 parameter")
		}

		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return false, fmt.Errorf("invalid Argon2 parameter: %w", err)
		}

		switch keyValue[0] {
		case "m":
			parsedMemory = value
		case "t":
			parsedIterations = value
		case "p":
			parsedParallelism = value
		default:
			return false, fmt.Errorf("unknown Argon2 parameter")
		}
	}

	// Do not trust parameters read from the database: malformed values
	// must not cause excessive memory or CPU consumption.
	if parsedMemory < 8*1024 || parsedMemory > 256*1024 ||
		parsedIterations < 1 || parsedIterations > 10 ||
		parsedParallelism < 1 || parsedParallelism > 8 {
		return false, fmt.Errorf("Argon2 parameters out of range")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 {
		return false, fmt.Errorf("invalid password salt")
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedHash) != keyLength {
		return false, fmt.Errorf("invalid password hash")
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(parsedIterations),
		uint32(parsedMemory),
		uint8(parsedParallelism),
		uint32(len(expectedHash)),
	)

	if len(actualHash) != len(expectedHash) {
		return false, nil
	}

	var diff byte
	for i := range actualHash {
		diff |= actualHash[i] ^ expectedHash[i]
	}

	return diff == 0, nil
}

package config

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	DBPath       string
	CSRFKey      []byte
	SessionKey   []byte
	CookieDomain string
	CookieSecure bool
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:         getEnv("PORT", "8585"),
		DBPath:       getEnv("DB_PATH", "./crochet.db"),
		CookieDomain: getEnv("COOKIE_DOMAIN", ""),
		CookieSecure: getEnv("COOKIE_SECURE", "false") == "true",
	}

	// CSRF Key (critical for security)
	csrfKeyStr := os.Getenv("CSRF_KEY")
	if csrfKeyStr == "" {
		slog.Warn("CSRF_KEY environment variable not set. Generating a random key for development. This key will change on each restart. PLEASE SET CSRF_KEY IN PRODUCTION!")
		cfg.CSRFKey = generateRandomBytes(32)
	} else {
		decodedKey, err := base64.StdEncoding.DecodeString(csrfKeyStr)
		if err != nil || len(decodedKey) < 32 {
			slog.Warn("CSRF_KEY is invalid or too short (min 32 bytes recommended). Generating a random key for development. PLEASE SET A SECURE CSRF_KEY IN PRODUCTION!")
			cfg.CSRFKey = generateRandomBytes(32)
		} else {
			cfg.CSRFKey = decodedKey
		}
	}

	// Session Key (critical for security)
	sessionKeyStr := os.Getenv("SESSION_KEY")
	if sessionKeyStr == "" {
		slog.Warn("SESSION_KEY environment variable not set. Generating a random key for development. Sessions will be invalid on restart. PLEASE SET SESSION_KEY IN PRODUCTION!")
		cfg.SessionKey = generateRandomBytes(32)
	} else {
		decodedKey, err := base64.StdEncoding.DecodeString(sessionKeyStr)
		if err != nil || len(decodedKey) < 32 {
			slog.Warn("SESSION_KEY is invalid or too short (min 32 bytes recommended). Generating a random key for development. PLEASE SET A SECURE SESSION_KEY IN PRODUCTION!")
			cfg.SessionKey = generateRandomBytes(32)
		} else {
			cfg.SessionKey = decodedKey
		}
	}

	// Make sure port is valid
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		slog.Error("Invalid PORT environment variable. Falling back to default.", "PORT", os.Getenv("PORT"))
		cfg.Port = "8585"
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// generateRandomBytes generates a random byte slice of specified length
// Uses crypto/rand for secure random numbers.
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil { // Use crypto/rand
		slog.Error("Failed to read random bytes", "error", err)
		// Fallback to a less secure random string if crypto/rand fails
		// This fallback is only for panic prevention, not for production use
		fallbackKey := "fallback-insecure-key-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		// Ensure the fallback key is at least n bytes long
		if len(fallbackKey) < n {
			paddedKey := make([]byte, n)
			copy(paddedKey, fallbackKey)
			return paddedKey
		}
		return []byte(fallbackKey)[:n]
	}
	return b
}
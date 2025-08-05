package bridge

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SecretHandler manages secrets with encryption and leak detection
type SecretHandler struct {
	leakDetector *LeakDetector
	key          []byte
	mu           sync.RWMutex
}

// LeakDetector tracks secrets to prevent leaks in logs
type LeakDetector struct {
	patterns map[string]string // secret value -> secret name
	mu       sync.RWMutex
}

// NewSecretHandler creates a new secret handler
func NewSecretHandler() (*SecretHandler, error) {
	// Generate encryption key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	return &SecretHandler{
		leakDetector: NewLeakDetector(),
		key:          key,
	}, nil
}

// NewLeakDetector creates a new leak detector
func NewLeakDetector() *LeakDetector {
	return &LeakDetector{
		patterns: make(map[string]string),
	}
}

// PrepareSecrets prepares secrets for workflow execution
func (sh *SecretHandler) PrepareSecrets(workspace *Workspace, secrets map[string]string) (string, error) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	// Track all secrets for leak detection
	for name, value := range secrets {
		sh.leakDetector.Track(name, value)
	}

	// Write secrets file (act format)
	secretsPath := filepath.Join(workspace.SecretDir, ".secrets")

	file, err := os.OpenFile(secretsPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("create secrets file: %w", err)
	}
	defer file.Close()

	// Write each secret
	for key, value := range secrets {
		// Validate secret
		if err := sh.validateSecret(key, value); err != nil {
			return "", fmt.Errorf("invalid secret %s: %w", key, err)
		}

		fmt.Fprintf(file, "%s=%s\n", key, value)
	}

	return secretsPath, nil
}

// EncryptSecrets encrypts secrets for storage
func (sh *SecretHandler) EncryptSecrets(secrets map[string]string) (map[string]string, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	encrypted := make(map[string]string)

	for key, value := range secrets {
		encValue, err := sh.encrypt([]byte(value))
		if err != nil {
			return nil, fmt.Errorf("encrypt %s: %w", key, err)
		}
		encrypted[key] = base64.StdEncoding.EncodeToString(encValue)
	}

	return encrypted, nil
}

// DecryptSecrets decrypts secrets from storage
func (sh *SecretHandler) DecryptSecrets(encrypted map[string]string) (map[string]string, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	decrypted := make(map[string]string)

	for key, value := range encrypted {
		encValue, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}

		decValue, err := sh.decrypt(encValue)
		if err != nil {
			return nil, fmt.Errorf("decrypt %s: %w", key, err)
		}

		decrypted[key] = string(decValue)
	}

	return decrypted, nil
}

// SanitizeLogs removes secrets from log output
func (sh *SecretHandler) SanitizeLogs(logs []string) []string {
	return sh.leakDetector.SanitizeLogs(logs)
}

// validateSecret validates a secret key and value
func (sh *SecretHandler) validateSecret(key, value string) error {
	// Key validation
	if key == "" {
		return fmt.Errorf("empty key")
	}

	if strings.ContainsAny(key, " \t\n\r=") {
		return fmt.Errorf("invalid characters in key")
	}

	// Value validation
	if value == "" {
		return fmt.Errorf("empty value")
	}

	if len(value) > 64*1024 { // 64KB limit
		return fmt.Errorf("value too large")
	}

	return nil
}

// encrypt encrypts data using AES-256-GCM
func (sh *SecretHandler) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sh.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt decrypts data using AES-256-GCM
func (sh *SecretHandler) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sh.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Track adds a secret to the leak detector
func (ld *LeakDetector) Track(name, value string) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	// Track the exact value
	ld.patterns[value] = name

	// Also track common encodings
	ld.patterns[base64.StdEncoding.EncodeToString([]byte(value))] = name + "_base64"
	ld.patterns[fmt.Sprintf("%x", sha256.Sum256([]byte(value)))] = name + "_sha256"
}

// SanitizeLogs removes tracked secrets from logs
func (ld *LeakDetector) SanitizeLogs(logs []string) []string {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	sanitized := make([]string, len(logs))

	for i, log := range logs {
		sanitized[i] = ld.sanitizeLine(log)
	}

	return sanitized
}

// sanitizeLine sanitizes a single log line
func (ld *LeakDetector) sanitizeLine(line string) string {
	// Replace all tracked patterns
	for pattern, name := range ld.patterns {
		if pattern != "" && strings.Contains(line, pattern) {
			replacement := fmt.Sprintf("***%s***", name)
			line = strings.ReplaceAll(line, pattern, replacement)
		}
	}

	return line
}

// ParseSecretsFile parses a secrets file in act format
func ParseSecretsFile(path string) (map[string]string, error) {
	secrets := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		secrets[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return secrets, nil
}

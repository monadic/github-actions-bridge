package leakdetector

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
)

// Detector tracks and detects secret leaks in logs and outputs
type Detector struct {
	patterns map[string]string // secret value -> secret name
	mu       sync.RWMutex
}

// New creates a new leak detector
func New() *Detector {
	return &Detector{
		patterns: make(map[string]string),
	}
}

// Track adds a secret to track for leaks
func (d *Detector) Track(name, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Skip empty values
	if value == "" {
		return
	}
	
	// Track the exact value
	d.patterns[value] = name
	
	// Track common encodings
	d.patterns[base64.StdEncoding.EncodeToString([]byte(value))] = name + "_base64"
	d.patterns[base64.URLEncoding.EncodeToString([]byte(value))] = name + "_base64url"
	
	// Track hex encoding
	d.patterns[fmt.Sprintf("%x", sha256.Sum256([]byte(value)))] = name + "_sha256"
	
	// Track URL encoded
	d.patterns[strings.ReplaceAll(value, " ", "%20")] = name + "_urlencoded"
	d.patterns[strings.ReplaceAll(value, " ", "+")] = name + "_urlencoded_plus"
}

// SanitizeLogs removes tracked secrets from logs
func (d *Detector) SanitizeLogs(logs []string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	sanitized := make([]string, len(logs))
	for i, log := range logs {
		sanitized[i] = d.sanitizeLine(log)
	}
	
	return sanitized
}

// SanitizeString removes tracked secrets from a single string
func (d *Detector) SanitizeString(s string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	return d.sanitizeLine(s)
}

// sanitizeLine sanitizes a single line
func (d *Detector) sanitizeLine(line string) string {
	// Replace all tracked patterns
	for pattern, name := range d.patterns {
		if pattern != "" && strings.Contains(line, pattern) {
			replacement := fmt.Sprintf("***%s***", name)
			line = strings.ReplaceAll(line, pattern, replacement)
		}
	}
	
	return line
}

// CheckForLeaks checks if any tracked secrets appear in the text
func (d *Detector) CheckForLeaks(text string) (bool, []string) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	var leaks []string
	found := false
	
	for pattern, name := range d.patterns {
		if pattern != "" && strings.Contains(text, pattern) {
			found = true
			leaks = append(leaks, name)
		}
	}
	
	return found, leaks
}

// Clear removes all tracked patterns
func (d *Detector) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.patterns = make(map[string]string)
}

// Count returns the number of tracked patterns
func (d *Detector) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	return len(d.patterns)
}

// GetTrackedNames returns all secret names being tracked
func (d *Detector) GetTrackedNames() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	names := make(map[string]bool)
	for _, name := range d.patterns {
		// Remove encoding suffixes
		baseName := strings.Split(name, "_")[0]
		names[baseName] = true
	}
	
	var result []string
	for name := range names {
		result = append(result, name)
	}
	
	return result
}
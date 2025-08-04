package leakdetector

import (
	"fmt"
	"strings"
	"sync"
)

// Detector prevents secrets from appearing in logs and output
type Detector struct {
	mu      sync.RWMutex
	secrets map[string]string // key -> secret value
	masks   map[string]string // secret value -> mask
}

// New creates a new leak detector
func New() *Detector {
	return &Detector{
		secrets: make(map[string]string),
		masks:   make(map[string]string),
	}
}

// Track registers a secret for leak detection
func (d *Detector) Track(key, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.secrets[key] = value
	// Create a mask for this secret
	mask := fmt.Sprintf("***%s***", key)
	d.masks[value] = mask
}

// ScanAndMask checks content for secrets and returns masked version
func (d *Detector) ScanAndMask(content string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	masked := content
	for secret, mask := range d.masks {
		if secret != "" && len(secret) > 3 { // Only mask non-trivial secrets
			masked = strings.ReplaceAll(masked, secret, mask)
		}
	}

	return masked
}

// CheckForLeaks returns true if content contains any tracked secrets
func (d *Detector) CheckForLeaks(content string) (bool, []string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var leakedKeys []string
	for key, secret := range d.secrets {
		if secret != "" && strings.Contains(content, secret) {
			leakedKeys = append(leakedKeys, key)
		}
	}

	return len(leakedKeys) > 0, leakedKeys
}

// Clear removes all tracked secrets
func (d *Detector) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.secrets = make(map[string]string)
	d.masks = make(map[string]string)
}

// Count returns the number of tracked secrets
func (d *Detector) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.secrets)
}

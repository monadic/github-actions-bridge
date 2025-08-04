package bridge

import (
	"fmt"
	"os"
	"path/filepath"
)

// prepareSecrets creates a secrets file for act (Jesper's file approach)
func (b *ActionsBridge) prepareSecrets(ws *Workspace, secrets map[string]string) (string, error) {
	// Act-compatible secrets file
	secretsPath := filepath.Join(ws.SecretDir, ".secrets")
	
	// Create with restrictive permissions
	file, err := os.OpenFile(secretsPath, 
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("create secrets file: %w", err)
	}
	defer file.Close()

	// Track all secrets for leak detection
	for key, value := range secrets {
		// Register with leak detector
		b.leakDetector.Track(key, value)
		
		// Write in act format (KEY=value)
		if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
			return "", fmt.Errorf("write secret %s: %w", key, err)
		}
	}

	// Also create individual secret files for flexibility
	for key, value := range secrets {
		secretFile := filepath.Join(ws.SecretDir, key)
		if err := ws.WriteSecret(key, []byte(value)); err != nil {
			return "", fmt.Errorf("write individual secret %s: %w", key, err)
		}
	}

	return secretsPath, nil
}

// SecretProvider defines how secrets are provided to workflows
type SecretProvider interface {
	GetSecrets(context map[string]interface{}) (map[string]string, error)
}

// ConfigHubSecretProvider retrieves secrets from ConfigHub
type ConfigHubSecretProvider struct {
	// In a real implementation, this would connect to ConfigHub
}

func (p *ConfigHubSecretProvider) GetSecrets(context map[string]interface{}) (map[string]string, error) {
	// Placeholder implementation
	// In production, this would fetch from ConfigHub's secret store
	return map[string]string{
		"EXAMPLE_SECRET": "example-value",
	}, nil
}

// EnvironmentSecretProvider reads secrets from environment variables
type EnvironmentSecretProvider struct {
	prefix string
}

func NewEnvironmentSecretProvider(prefix string) *EnvironmentSecretProvider {
	return &EnvironmentSecretProvider{prefix: prefix}
}

func (p *EnvironmentSecretProvider) GetSecrets(context map[string]interface{}) (map[string]string, error) {
	secrets := make(map[string]string)
	
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, p.prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], p.prefix)
				secrets[key] = parts[1]
			}
		}
	}
	
	return secrets, nil
}

// FileSecretProvider reads secrets from files
type FileSecretProvider struct {
	directory string
}

func NewFileSecretProvider(directory string) *FileSecretProvider {
	return &FileSecretProvider{directory: directory}
}

func (p *FileSecretProvider) GetSecrets(context map[string]interface{}) (map[string]string, error) {
	secrets := make(map[string]string)
	
	entries, err := os.ReadDir(p.directory)
	if err != nil {
		return nil, fmt.Errorf("read secrets directory: %w", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			content, err := os.ReadFile(filepath.Join(p.directory, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("read secret %s: %w", entry.Name(), err)
			}
			secrets[entry.Name()] = strings.TrimSpace(string(content))
		}
	}
	
	return secrets, nil
}

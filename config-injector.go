package bridge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// injectConfigs writes configuration values to the workspace
func (b *ActionsBridge) injectConfigs(ws *Workspace, configs map[string]interface{}) error {
	// Write individual config files as JSON
	for key, value := range configs {
		// Create JSON file for each config
		jsonPath := filepath.Join(ws.ConfigDir, key+".json")
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal config %s: %w", key, err)
		}
		
		if err := ws.WriteConfig(key+".json", data); err != nil {
			return fmt.Errorf("write config %s: %w", jsonPath, err)
		}
	}

	// Create a combined config file
	combinedPath := filepath.Join(ws.ConfigDir, "config.json")
	combinedData, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal combined config: %w", err)
	}
	
	if err := ws.WriteConfig("config.json", combinedData); err != nil {
		return fmt.Errorf("write combined config: %w", err)
	}

	// Also create an env file for simple string values
	if err := b.createEnvFile(ws, configs); err != nil {
		return fmt.Errorf("create env file: %w", err)
	}

	// Create a YAML version for convenience
	if err := b.createYAMLConfig(ws, configs); err != nil {
		// Non-fatal, just log
		fmt.Printf("Warning: failed to create YAML config: %v\n", err)
	}

	return nil
}

// createEnvFile creates a .env file with string configuration values
func (b *ActionsBridge) createEnvFile(ws *Workspace, configs map[string]interface{}) error {
	envPath := filepath.Join(ws.ConfigDir, ".env")
	envFile, err := os.Create(envPath)
	if err != nil {
		return err
	}
	defer envFile.Close()

	// Write header
	fmt.Fprintln(envFile, "# ConfigHub configuration values")
	fmt.Fprintln(envFile, "# Generated for GitHub Actions workflow")
	fmt.Fprintln(envFile)

	// Process configs
	b.writeEnvVars(envFile, "CONFIG", configs)

	return nil
}

// writeEnvVars recursively writes configuration values as environment variables
func (b *ActionsBridge) writeEnvVars(file *os.File, prefix string, data map[string]interface{}) {
	for key, value := range data {
		envKey := fmt.Sprintf("%s_%s", prefix, strings.ToUpper(key))
		
		switch v := value.(type) {
		case string:
			fmt.Fprintf(file, "%s=%s\n", envKey, v)
		case int, int64, float64, bool:
			fmt.Fprintf(file, "%s=%v\n", envKey, v)
		case map[string]interface{}:
			// Recurse for nested objects
			b.writeEnvVars(file, envKey, v)
		case []interface{}:
			// For arrays, create indexed variables
			for i, item := range v {
				if str, ok := item.(string); ok {
					fmt.Fprintf(file, "%s_%d=%s\n", envKey, i, str)
				}
			}
			// Also store the count
			fmt.Fprintf(file, "%s_COUNT=%d\n", envKey, len(v))
		default:
			// For complex types, marshal to JSON
			if jsonData, err := json.Marshal(v); err == nil {
				fmt.Fprintf(file, "%s=%s\n", envKey, string(jsonData))
			}
		}
	}
}

// createYAMLConfig creates a YAML version of the configuration
func (b *ActionsBridge) createYAMLConfig(ws *Workspace, configs map[string]interface{}) error {
	// Note: This is a simplified YAML writer
	// In production, use a proper YAML library
	yamlPath := filepath.Join(ws.ConfigDir, "config.yaml")
	yamlFile, err := os.Create(yamlPath)
	if err != nil {
		return err
	}
	defer yamlFile.Close()

	fmt.Fprintln(yamlFile, "# ConfigHub configuration values")
	fmt.Fprintln(yamlFile, "# Generated for GitHub Actions workflow")
	fmt.Fprintln(yamlFile)
	
	b.writeSimpleYAML(yamlFile, configs, 0)
	
	return nil
}

// writeSimpleYAML writes a simplified YAML representation
func (b *ActionsBridge) writeSimpleYAML(file *os.File, data map[string]interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)
	
	for key, value := range data {
		switch v := value.(type) {
		case string:
			fmt.Fprintf(file, "%s%s: %q\n", indentStr, key, v)
		case int, int64, float64, bool:
			fmt.Fprintf(file, "%s%s: %v\n", indentStr, key, v)
		case map[string]interface{}:
			fmt.Fprintf(file, "%s%s:\n", indentStr, key)
			b.writeSimpleYAML(file, v, indent+1)
		case []interface{}:
			fmt.Fprintf(file, "%s%s:\n", indentStr, key)
			for _, item := range v {
				fmt.Fprintf(file, "%s  - %v\n", indentStr, item)
			}
		default:
			// Complex types as JSON
			if jsonData, err := json.Marshal(v); err == nil {
				fmt.Fprintf(file, "%s%s: %s\n", indentStr, key, string(jsonData))
			}
		}
	}
}

// ConfigInjectionMode defines how configs are injected
type ConfigInjectionMode string

const (
	ConfigInjectionModeFiles ConfigInjectionMode = "files"
	ConfigInjectionModeEnv   ConfigInjectionMode = "env"
	ConfigInjectionModeBoth  ConfigInjectionMode = "both"
)

// ConfigFormat defines the format for configuration files
type ConfigFormat string

const (
	ConfigFormatJSON ConfigFormat = "json"
	ConfigFormatYAML ConfigFormat = "yaml"
	ConfigFormatEnv  ConfigFormat = "env"
	ConfigFormatAll  ConfigFormat = "all"
)

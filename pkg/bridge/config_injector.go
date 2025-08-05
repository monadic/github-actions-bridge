package bridge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v3"
)

// ConfigInjector handles configuration injection into workspaces
type ConfigInjector struct {
	workspace *Workspace
}

// NewConfigInjector creates a new config injector
func NewConfigInjector(workspace *Workspace) *ConfigInjector {
	return &ConfigInjector{
		workspace: workspace,
	}
}

// InjectConfigs injects configurations into the workspace
func (ci *ConfigInjector) InjectConfigs(configs map[string]interface{}) error {
	// Write individual config files
	for key, value := range configs {
		if err := ci.writeConfigFile(key, value); err != nil {
			return fmt.Errorf("write config %s: %w", key, err)
		}
	}
	
	// Create consolidated config files
	if err := ci.writeConsolidatedConfigs(configs); err != nil {
		return fmt.Errorf("write consolidated configs: %w", err)
	}
	
	// Create environment file for simple values
	if err := ci.createEnvFile(configs); err != nil {
		return fmt.Errorf("create env file: %w", err)
	}
	
	return nil
}

// writeConfigFile writes a single configuration file
func (ci *ConfigInjector) writeConfigFile(key string, value interface{}) error {
	// Determine format based on value type
	var data []byte
	var ext string
	var err error
	
	switch v := value.(type) {
	case string:
		// Plain text file
		data = []byte(v)
		ext = ".txt"
	case map[string]interface{}, []interface{}:
		// JSON file
		data, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		ext = ".json"
	default:
		// Default to JSON
		data, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		ext = ".json"
	}
	
	filename := fmt.Sprintf("%s%s", key, ext)
	path := filepath.Join(ci.workspace.ConfigDir, filename)
	
	return os.WriteFile(path, data, 0644)
}

// writeConsolidatedConfigs creates consolidated config files
func (ci *ConfigInjector) writeConsolidatedConfigs(configs map[string]interface{}) error {
	// Write as JSON
	jsonPath := filepath.Join(ci.workspace.ConfigDir, "config.json")
	jsonData, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	
	// Write as YAML
	yamlPath := filepath.Join(ci.workspace.ConfigDir, "config.yaml")
	yamlData, err := yaml.Marshal(configs)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	if err := os.WriteFile(yamlPath, yamlData, 0644); err != nil {
		return fmt.Errorf("write yaml: %w", err)
	}
	
	return nil
}

// createEnvFile creates an environment variable file
func (ci *ConfigInjector) createEnvFile(configs map[string]interface{}) error {
	envPath := filepath.Join(ci.workspace.ConfigDir, ".env")
	file, err := os.Create(envPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write header
	fmt.Fprintln(file, "# ConfigHub injected configuration")
	fmt.Fprintln(file, "# Generated automatically - DO NOT EDIT")
	fmt.Fprintln(file)
	
	// Convert configs to environment variables
	flatConfigs := ci.flattenConfigs(configs, "CONFIG")
	
	for key, value := range flatConfigs {
		fmt.Fprintf(file, "%s=%s\n", key, value)
	}
	
	return nil
}

// flattenConfigs flattens nested configs into environment variables
func (ci *ConfigInjector) flattenConfigs(configs map[string]interface{}, prefix string) map[string]string {
	flat := make(map[string]string)
	
	for key, value := range configs {
		envKey := fmt.Sprintf("%s_%s", prefix, strings.ToUpper(key))
		
		switch v := value.(type) {
		case string:
			flat[envKey] = v
		case int, int32, int64, float32, float64:
			flat[envKey] = fmt.Sprintf("%v", v)
		case bool:
			flat[envKey] = fmt.Sprintf("%t", v)
		case map[string]interface{}:
			// Recursively flatten nested maps
			nested := ci.flattenConfigs(v, envKey)
			for k, val := range nested {
				flat[k] = val
			}
		default:
			// Convert complex types to JSON
			if data, err := json.Marshal(v); err == nil {
				flat[envKey] = string(data)
			}
		}
	}
	
	return flat
}

// InjectWorkflowConfig injects configuration directly into a workflow
func (ci *ConfigInjector) InjectWorkflowConfig(workflowPath string, configs map[string]interface{}) error {
	// Read existing workflow
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	
	// Parse workflow
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return fmt.Errorf("parse workflow: %w", err)
	}
	
	// Inject environment variables at job level
	if jobs, ok := workflow["jobs"].(map[string]interface{}); ok {
		for _, job := range jobs {
			if jobMap, ok := job.(map[string]interface{}); ok {
				// Add or merge env section
				if env, exists := jobMap["env"].(map[string]interface{}); exists {
					// Merge with existing env
					for k, v := range ci.flattenConfigs(configs, "CONFIG") {
						env[k] = v
					}
				} else {
					// Create new env section
					jobMap["env"] = ci.flattenConfigs(configs, "CONFIG")
				}
			}
		}
	}
	
	// Write back modified workflow
	modifiedData, err := yaml.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}
	
	return os.WriteFile(workflowPath, modifiedData, 0644)
}

// CreateConfigScript creates a script that exports all configs
func (ci *ConfigInjector) CreateConfigScript(configs map[string]interface{}) error {
	scriptPath := filepath.Join(ci.workspace.ConfigDir, "load-configs.sh")
	
	file, err := os.Create(scriptPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write shebang and header
	fmt.Fprintln(file, "#!/bin/bash")
	fmt.Fprintln(file, "# ConfigHub configuration loader")
	fmt.Fprintln(file, "# Source this file to load all configurations")
	fmt.Fprintln(file)
	
	// Export all configs
	flat := ci.flattenConfigs(configs, "CONFIG")
	for key, value := range flat {
		// Escape single quotes in value
		escaped := strings.ReplaceAll(value, "'", "'\"'\"'")
		fmt.Fprintf(file, "export %s='%s'\n", key, escaped)
	}
	
	// Add helper functions
	fmt.Fprintln(file, "\n# Helper functions")
	fmt.Fprintln(file, "config_get() {")
	fmt.Fprintln(file, "  local key=\"CONFIG_${1^^}\"")
	fmt.Fprintln(file, "  echo \"${!key}\"")
	fmt.Fprintln(file, "}")
	
	fmt.Fprintln(file, "\nconfig_has() {")
	fmt.Fprintln(file, "  local key=\"CONFIG_${1^^}\"")
	fmt.Fprintln(file, "  [[ -n \"${!key}\" ]]")
	fmt.Fprintln(file, "}")
	
	// Make executable
	return os.Chmod(scriptPath, 0755)
}

// ValidateConfigs validates configuration data
func (ci *ConfigInjector) ValidateConfigs(configs map[string]interface{}) error {
	for key, value := range configs {
		// Validate key
		if key == "" {
			return fmt.Errorf("empty configuration key")
		}
		
		if strings.ContainsAny(key, " \t\n\r/\\") {
			return fmt.Errorf("invalid characters in key: %s", key)
		}
		
		// Validate value size
		if data, err := json.Marshal(value); err == nil {
			if len(data) > 1024*1024 { // 1MB limit per config
				return fmt.Errorf("configuration %s too large: %d bytes", key, len(data))
			}
		}
	}
	
	return nil
}
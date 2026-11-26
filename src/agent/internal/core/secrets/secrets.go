// Package secrets handles secure API key management using SOPS encryption.
// API keys are stored in a SOPS-encrypted YAML file and decrypted at runtime
// using an age key retrieved from 1Password.
package secrets

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Default age recipient for SOPS encryption
const defaultAgeRecipient = "age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354"

// Config represents the structure of the secrets file
type Config struct {
	APIKeys map[string]string `json:"apiKeys"`
}

// Manager handles loading and accessing secrets
type Manager struct {
	mu           sync.RWMutex
	config       *Config
	configPath   string
	ageKey       string
	stopWatcher  chan struct{}
	lastModified time.Time
}

// NewManager creates a new secrets manager
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
	}

	// Load age key from environment or 1Password
	if err := m.loadAgeKey(); err != nil {
		return nil, fmt.Errorf("failed to load age key: %w", err)
	}

	// Load and decrypt config
	if err := m.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return m, nil
}

// NewManagerWithKey creates a manager with a provided age key
func NewManagerWithKey(configPath, ageKey string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
		ageKey:     ageKey,
	}

	if err := m.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return m, nil
}

// loadAgeKey tries to get the age key from environment or 1Password
func (m *Manager) loadAgeKey() error {
	// First try environment variable
	if key := os.Getenv("SOPS_AGE_KEY"); key != "" {
		m.ageKey = key
		return nil
	}

	// Try 1Password CLI
	if key, err := m.getKeyFrom1Password(); err == nil {
		m.ageKey = key
		return nil
	}

	return errors.New("no SOPS age key found: set SOPS_AGE_KEY env var or ensure 1Password CLI is configured")
}

// getKeyFrom1Password retrieves the SOPS age key from 1Password
func (m *Manager) getKeyFrom1Password() (string, error) {
	// Try to read from 1Password using op CLI
	cmd := exec.Command("op", "read", "op://Private/SOPS/password")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("1Password CLI error: %s", stderr.String())
	}

	key := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(key, "AGE-SECRET-KEY-") {
		return "", errors.New("invalid SOPS key format from 1Password")
	}

	return key, nil
}

// Load decrypts and loads the secrets file
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create empty config if file doesn't exist
		m.config = &Config{APIKeys: make(map[string]string)}
		return nil
	}

	// Decrypt using sops CLI
	cmd := exec.Command("sops", "-d", m.configPath)
	cmd.Env = append(os.Environ(), "SOPS_AGE_KEY="+m.ageKey)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to decrypt secrets: %s - %s", err, stderr.String())
	}

	// Parse YAML to extract apiKeys
	config, err := parseYAML(stdout.String())
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.config = config

	// Update last modified time
	if info, err := os.Stat(m.configPath); err == nil {
		m.lastModified = info.ModTime()
	}

	return nil
}

// Reload decrypts and reloads the secrets file
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil // No file to reload
	}

	// Check if file has changed
	info, err := os.Stat(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config: %w", err)
	}
	if info.ModTime().Equal(m.lastModified) {
		return nil // No change
	}

	// Decrypt using sops CLI
	cmd := exec.Command("sops", "-d", m.configPath)
	cmd.Env = append(os.Environ(), "SOPS_AGE_KEY="+m.ageKey)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to decrypt secrets: %s - %s", err, stderr.String())
	}

	// Parse YAML to extract apiKeys
	config, err := parseYAML(stdout.String())
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.config = config
	m.lastModified = info.ModTime()

	return nil
}

// StartWatcher starts a background goroutine that watches for file changes
// and reloads secrets automatically. Returns a stop channel.
func (m *Manager) StartWatcher(intervalSeconds int) chan struct{} {
	if intervalSeconds <= 0 {
		intervalSeconds = 5 // Default 5 seconds
	}

	stop := make(chan struct{})
	m.mu.Lock()
	m.stopWatcher = stop
	m.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := m.Reload(); err != nil {
					fmt.Printf("Secrets reload error: %v\n", err)
				}
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// StopWatcher stops the background watcher
func (m *Manager) StopWatcher() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopWatcher != nil {
		close(m.stopWatcher)
		m.stopWatcher = nil
	}
}

// Get returns an API key by provider name
func (m *Manager) Get(provider string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil || m.config.APIKeys == nil {
		return ""
	}

	return m.config.APIKeys[provider]
}

// Set stores an API key in memory (call Save to persist)
func (m *Manager) Set(provider, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		m.config = &Config{APIKeys: make(map[string]string)}
	}
	if m.config.APIKeys == nil {
		m.config.APIKeys = make(map[string]string)
	}

	m.config.APIKeys[provider] = key
}

// Save encrypts and saves the config to file
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return errors.New("no config to save")
	}

	// Get age recipient from environment or use default
	ageRecipient := os.Getenv("SOPS_AGE_RECIPIENT")
	if ageRecipient == "" {
		ageRecipient = defaultAgeRecipient
	}

	// Convert config to YAML
	yamlContent, err := configToYAML(m.config)
	if err != nil {
		return fmt.Errorf("failed to create yaml: %w", err)
	}

	// Write to temp file
	tmpFile := m.configPath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Encrypt using sops CLI
	cmd := exec.Command("sops", "--encrypt", "--age", ageRecipient, tmpFile)
	cmd.Env = append(os.Environ(), "SOPS_AGE_KEY="+m.ageKey)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to encrypt: %s - %s", err, stderr.String())
	}

	// Write encrypted content
	if err := os.WriteFile(m.configPath, stdout.Bytes(), 0600); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	_ = os.Remove(tmpFile)
	return nil
}

// DefaultConfigPath returns the default secrets file location
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agent", "secrets.enc.yaml")
}

// EnsureDir ensures the config directory exists
func EnsureDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".config", "agent")
	return os.MkdirAll(dir, 0700)
}

// SupportedProviders defines which providers can be stored in secrets
var SupportedProviders = []string{
	"openai", "anthropic", "google", "deepseek", "agentgateway", "minimax",
}

// sopsFileConfig represents the SOPS-encrypted file structure
type sopsFileConfig struct {
	Data string `yaml:"data"`
}

// yamlConfig represents the inner YAML structure with apiKeys
type yamlConfig struct {
	APIKeys map[string]string `yaml:"apiKeys"`
}

func parseYAML(content string) (*Config, error) {
	// Try to parse as SOPS format first (has 'data' field with inner YAML)
	var sopsConf sopsFileConfig
	if err := yaml.Unmarshal([]byte(content), &sopsConf); err == nil && sopsConf.Data != "" {
		// Parse the inner YAML from the 'data' field
		content = sopsConf.Data
	}

	// Parse the actual YAML content
	var yamlConf yamlConfig
	if err := yaml.Unmarshal([]byte(content), &yamlConf); err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}

	// Validate and filter to only known providers
	apiKeys := make(map[string]string)
	for k, v := range yamlConf.APIKeys {
		if slices.Contains(SupportedProviders, k) {
			apiKeys[k] = v
		}
	}

	return &Config{APIKeys: apiKeys}, nil
}

func configToYAML(config *Config) (string, error) {
	// Build YAML manually with proper formatting
	var lines []string
	lines = append(lines, "# DO NOT EDIT - This file is encrypted with SOPS")
	lines = append(lines, "# Managed by plan/agent/internal/core/secrets")
	lines = append(lines, "")
	lines = append(lines, "apiKeys:")

	for key, value := range config.APIKeys {
		lines = append(lines, fmt.Sprintf("  %s: %q", key, value))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

// --- JSON support for programmatic access ---

// ConfigJSON returns the config as JSON
func (m *Manager) ConfigJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return []byte(`{"apiKeys":{}}`), nil
	}

	return json.MarshalIndent(m.config, "", "  ")
}

// ProviderList returns a list of configured providers
func (m *Manager) ProviderList() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil || m.config.APIKeys == nil {
		return nil
	}

	providers := make([]string, 0, len(m.config.APIKeys))
	for k := range m.config.APIKeys {
		providers = append(providers, k)
	}
	return providers
}

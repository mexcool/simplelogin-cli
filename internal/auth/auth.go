package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"go.yaml.in/yaml/v3"
)

// Config represents the config file structure.
type Config struct {
	APIKey  string `yaml:"api_key"`
	APIBase string `yaml:"api_base"`
	OPRef   string `yaml:"op_ref"`
}

var configDir string

func init() {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		configDir = filepath.Join(xdg, "simplelogin")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			configDir = filepath.Join(".config", "simplelogin")
		} else {
			configDir = filepath.Join(home, ".config", "simplelogin")
		}
	}
}

// ConfigDir returns the config directory path.
func ConfigDir() string {
	return configDir
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(configDir, "config.yml")
}

// loadConfig reads and parses the config file. Returns an empty Config if the file doesn't exist.
func loadConfig() Config {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return Config{}
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}
	}
	return cfg
}

// saveConfig writes the config to disk, creating the directory if needed.
func saveConfig(cfg Config) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0600); err != nil {
		return err
	}
	return nil
}

// GetAPIKey returns the API key following the priority chain:
// 1. SIMPLELOGIN_API_KEY env var
// 2. SL_API_KEY env var
// 3. 1Password via op CLI (if op_ref is set in config)
// 4. api_key from config file
func GetAPIKey() (string, error) {
	// 1. SIMPLELOGIN_API_KEY env var
	if key := os.Getenv("SIMPLELOGIN_API_KEY"); key != "" {
		return key, nil
	}

	// 2. SL_API_KEY env var
	if key := os.Getenv("SL_API_KEY"); key != "" {
		return key, nil
	}

	// Load config file
	cfg := loadConfig()

	// 3. 1Password via op CLI
	if cfg.OPRef != "" {
		if opPath, err := exec.LookPath("op"); err == nil && opPath != "" {
			cmd := exec.Command("op", "read", cfg.OPRef)
			out, err := cmd.Output()
			if err == nil {
				key := strings.TrimSpace(string(out))
				if key != "" {
					return key, nil
				}
			}
		}
	}

	// 4. Config file api_key
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	return "", fmt.Errorf("no API key found. Run: sl auth login")
}

// SaveAPIKey stores the API key in the config file.
func SaveAPIKey(key string) error {
	cfg := loadConfig()
	cfg.APIKey = key
	// Remove op_ref if setting direct key
	cfg.OPRef = ""
	return saveConfig(cfg)
}

// SaveOPRef stores the 1Password reference in the config file.
func SaveOPRef(vault, item string) error {
	ref := fmt.Sprintf("op://%s/%s/credential", vault, item)
	cfg := loadConfig()
	cfg.OPRef = ref
	// Remove direct api_key if setting 1Password ref
	cfg.APIKey = ""
	return saveConfig(cfg)
}

// ClearConfig removes the API key and op_ref from the config file.
func ClearConfig() error {
	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		return nil
	}
	cfg := loadConfig()
	cfg.APIKey = ""
	cfg.OPRef = ""
	return saveConfig(cfg)
}

// MaskKey masks an API key for display, showing only the first 4 and last 4 characters.
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// GetOPRef returns the stored 1Password reference, if any.
func GetOPRef() string {
	cfg := loadConfig()
	return cfg.OPRef
}

// GetAPIBase returns the configured API base URL, or the default if none is set.
func GetAPIBase() string {
	cfg := loadConfig()
	if cfg.APIBase != "" {
		return cfg.APIBase
	}
	return api.BaseURL
}

// SaveAPIBase stores the API base URL in the config file.
func SaveAPIBase(baseURL string) error {
	cfg := loadConfig()
	cfg.APIBase = strings.TrimRight(baseURL, "/")
	return saveConfig(cfg)
}

package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

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
	loadConfig()

	// 3. 1Password via op CLI
	if opRef := viper.GetString("op_ref"); opRef != "" {
		if opPath, err := exec.LookPath("op"); err == nil && opPath != "" {
			cmd := exec.Command("op", "read", opRef)
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
	if key := viper.GetString("api_key"); key != "" {
		return key, nil
	}

	return "", fmt.Errorf("no API key found. Run: sl auth login")
}

// SaveAPIKey stores the API key in the config file.
func SaveAPIKey(key string) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	loadConfig()
	viper.Set("api_key", key)
	// Remove op_ref if setting direct key
	viper.Set("op_ref", "")
	if err := viper.WriteConfigAs(ConfigPath()); err != nil {
		return err
	}
	return os.Chmod(ConfigPath(), 0600)
}

// SaveOPRef stores the 1Password reference in the config file.
func SaveOPRef(vault, item string) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	ref := fmt.Sprintf("op://%s/%s/credential", vault, item)
	loadConfig()
	viper.Set("op_ref", ref)
	// Remove direct api_key if setting 1Password ref
	viper.Set("api_key", "")
	if err := viper.WriteConfigAs(ConfigPath()); err != nil {
		return err
	}
	return os.Chmod(ConfigPath(), 0600)
}

// ClearConfig removes the API key and op_ref from the config file.
func ClearConfig() error {
	loadConfig()
	viper.Set("api_key", "")
	viper.Set("op_ref", "")
	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		return nil
	}
	if err := viper.WriteConfigAs(ConfigPath()); err != nil {
		return err
	}
	return os.Chmod(ConfigPath(), 0600)
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
	loadConfig()
	return viper.GetString("op_ref")
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configDir)
	_ = viper.ReadInConfig()
}

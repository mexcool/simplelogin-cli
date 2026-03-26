package auth

import (
	"os"
	"path/filepath"
	"testing"
)

// setConfigDir overrides the package-level configDir to point at a temp directory.
// It returns a cleanup function that restores the original value.
func setConfigDir(t *testing.T, dir string) {
	t.Helper()
	orig := configDir
	configDir = dir
	t.Cleanup(func() { configDir = orig })
}

// ---------------------------------------------------------------------------
// MaskKey
// ---------------------------------------------------------------------------

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"abc", "****"},
		{"12345678", "****"},       // exactly 8 chars -> masked
		{"123456789", "1234...6789"}, // 9 chars -> first 4 + last 4
		{"abcdefghijklmnop", "abcd...mnop"},
		{"sk_live_xxxxxxxxxxxxxxxxxxxx", "sk_l...xxxx"},
	}
	for _, tt := range tests {
		got := MaskKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Config load/save round-trip
// ---------------------------------------------------------------------------

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	cfg := Config{
		APIKey:  "test-key-123",
		APIBase: "https://custom.example.com",
		OPRef:   "op://vault/item/credential",
	}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	// Verify file exists with correct permissions
	info, err := os.Stat(ConfigPath())
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("config file permissions = %o, want 0600", perm)
	}

	// Load back
	loaded := loadConfig()
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, cfg.APIKey)
	}
	if loaded.APIBase != cfg.APIBase {
		t.Errorf("APIBase = %q, want %q", loaded.APIBase, cfg.APIBase)
	}
	if loaded.OPRef != cfg.OPRef {
		t.Errorf("OPRef = %q, want %q", loaded.OPRef, cfg.OPRef)
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, filepath.Join(dir, "nonexistent"))

	cfg := loadConfig()
	if cfg.APIKey != "" || cfg.OPRef != "" || cfg.APIBase != "" {
		t.Errorf("expected empty Config for missing file, got: %+v", cfg)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	// Write garbage to config file
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ConfigPath(), []byte(":::not yaml:::"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := loadConfig()
	if cfg.APIKey != "" {
		t.Errorf("expected empty Config for invalid YAML, got APIKey=%q", cfg.APIKey)
	}
}

// ---------------------------------------------------------------------------
// SaveAPIKey / SaveOPRef / ClearConfig
// ---------------------------------------------------------------------------

func TestSaveAPIKey(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	// First set an op_ref
	if err := saveConfig(Config{OPRef: "op://v/i/c"}); err != nil {
		t.Fatal(err)
	}

	// SaveAPIKey should clear op_ref
	if err := SaveAPIKey("my-new-key"); err != nil {
		t.Fatalf("SaveAPIKey: %v", err)
	}

	cfg := loadConfig()
	if cfg.APIKey != "my-new-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "my-new-key")
	}
	if cfg.OPRef != "" {
		t.Errorf("OPRef should be cleared after SaveAPIKey, got %q", cfg.OPRef)
	}
}

func TestSaveOPRef(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	// First set an api key
	if err := saveConfig(Config{APIKey: "old-key"}); err != nil {
		t.Fatal(err)
	}

	if err := SaveOPRef("my-vault", "my-item"); err != nil {
		t.Fatalf("SaveOPRef: %v", err)
	}

	cfg := loadConfig()
	expectedRef := "op://my-vault/my-item/credential"
	if cfg.OPRef != expectedRef {
		t.Errorf("OPRef = %q, want %q", cfg.OPRef, expectedRef)
	}
	if cfg.APIKey != "" {
		t.Errorf("APIKey should be cleared after SaveOPRef, got %q", cfg.APIKey)
	}
}

func TestClearConfig(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	// Set up a full config
	if err := saveConfig(Config{APIKey: "key", OPRef: "ref"}); err != nil {
		t.Fatal(err)
	}

	if err := ClearConfig(); err != nil {
		t.Fatalf("ClearConfig: %v", err)
	}

	cfg := loadConfig()
	if cfg.APIKey != "" || cfg.OPRef != "" {
		t.Errorf("expected cleared config, got: %+v", cfg)
	}
}

func TestClearConfig_NoFile(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, filepath.Join(dir, "nonexistent"))

	// Should not error when config doesn't exist
	if err := ClearConfig(); err != nil {
		t.Errorf("ClearConfig with no file should not error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetAPIKey priority chain
// ---------------------------------------------------------------------------

func TestGetAPIKey_EnvVarPriority(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	// Save a key to the config file
	if err := SaveAPIKey("config-key"); err != nil {
		t.Fatal(err)
	}

	// SIMPLELOGIN_API_KEY should take priority
	t.Setenv("SIMPLELOGIN_API_KEY", "env-key-1")
	t.Setenv("SL_API_KEY", "env-key-2")

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "env-key-1" {
		t.Errorf("expected SIMPLELOGIN_API_KEY to take priority, got %q", key)
	}
}

func TestGetAPIKey_SLAPIKeyFallback(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	if err := SaveAPIKey("config-key"); err != nil {
		t.Fatal(err)
	}

	// Unset primary, set secondary
	t.Setenv("SIMPLELOGIN_API_KEY", "")
	t.Setenv("SL_API_KEY", "env-key-2")

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "env-key-2" {
		t.Errorf("expected SL_API_KEY fallback, got %q", key)
	}
}

func TestGetAPIKey_ConfigFileFallback(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	if err := SaveAPIKey("config-key"); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SIMPLELOGIN_API_KEY", "")
	t.Setenv("SL_API_KEY", "")

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "config-key" {
		t.Errorf("expected config file key, got %q", key)
	}
}

func TestGetAPIKey_NoKeyAnywhere(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, filepath.Join(dir, "empty"))

	t.Setenv("SIMPLELOGIN_API_KEY", "")
	t.Setenv("SL_API_KEY", "")

	_, err := GetAPIKey()
	if err == nil {
		t.Error("expected error when no API key is configured")
	}
}

// ---------------------------------------------------------------------------
// ConfigDir / ConfigPath
// ---------------------------------------------------------------------------

func TestConfigPath(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	if got := ConfigDir(); got != dir {
		t.Errorf("ConfigDir() = %q, want %q", got, dir)
	}
	want := filepath.Join(dir, "config.yml")
	if got := ConfigPath(); got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// GetOPRef
// ---------------------------------------------------------------------------

func TestGetOPRef(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, dir)

	if err := saveConfig(Config{OPRef: "op://v/i/c"}); err != nil {
		t.Fatal(err)
	}

	got := GetOPRef()
	if got != "op://v/i/c" {
		t.Errorf("GetOPRef() = %q, want %q", got, "op://v/i/c")
	}
}

func TestGetOPRef_Empty(t *testing.T) {
	dir := t.TempDir()
	setConfigDir(t, filepath.Join(dir, "empty"))

	got := GetOPRef()
	if got != "" {
		t.Errorf("GetOPRef() for missing config = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Config directory creation
// ---------------------------------------------------------------------------

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	setConfigDir(t, nested)

	if err := saveConfig(Config{APIKey: "k"}); err != nil {
		t.Fatalf("saveConfig with nested dir: %v", err)
	}

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}

	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("directory permissions = %o, want 0700", perm)
	}
}

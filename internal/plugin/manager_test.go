package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAllPlugins(t *testing.T) {
	// Setup test environment
	cwd, _ := os.Getwd()
	testDir := filepath.Join(cwd, "testdata", "plugins")
	os.Setenv("AVM_PLUGIN_DIR", testDir)
	defer os.Unsetenv("AVM_PLUGIN_DIR")

	aliases, err := loadAllPluginsInternal(cwd)
	if err != nil {
		t.Fatalf("loadAllPluginsInternal returned error: %v", err)
	}

	// Assertions

	// 1. Healthy plugin should return "test-healthy" alias
	if val, ok := aliases["test-healthy"]; !ok {
		t.Errorf("Expected 'test-healthy' alias, not found")
	} else if val.Command != "echo healthy" {
		t.Errorf("Expected command 'echo healthy', got %q", val.Command)
	} else if val.PluginName != "healthy" {
		t.Errorf("Expected PluginName 'healthy', got %q", val.PluginName)
	}

	// 2. Slow plugin should return timeout or be skipped depending on logic (it sleeps for 2s, timeout is 500ms).
	// Because of 500ms timeout, the slow plugin should fail context timeout.
	if _, ok := aliases["test-slow"]; ok {
		t.Errorf("Expected 'test-slow' alias to be skipped due to timeout")
	}

	// 3. Malformed plugin should fail JSON decode and be skipped.
	if _, ok := aliases["test-malformed"]; ok {
		t.Errorf("Expected 'test-malformed' alias to be skipped")
	}

	// 4. No-trigger plugin should fail health-check and be skipped.
	if _, ok := aliases["test-no-trigger"]; ok {
		t.Errorf("Expected 'test-no-trigger' alias to be skipped due to health-check")
	}
}

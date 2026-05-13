package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s, t     string
		distance int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "abd", 1},
		{"run:tv", "run-tv", 1},
		{"tv-run", "run-tv", 6},
	}

	for _, test := range tests {
		d := levenshteinDistance(test.s, test.t)
		if d != test.distance {
			t.Errorf("levenshteinDistance(%q, %q) = %d; want %d", test.s, test.t, d, test.distance)
		}
	}
}

func TestNormalizeForComparison(t *testing.T) {
	tests := []struct {
		s        string
		expected string
	}{
		{"run-tv", "run-tv"},
		{"tv-run", "run-tv"},
		{"run:tv", "run-tv"},
		{"run_tv", "run-tv"},
		{"tv.run", "run-tv"},
		{"a-b:c", "a-b-c"},
		{"c:b-a", "a-b-c"},
	}

	for _, test := range tests {
		actual := normalizeForComparison(test.s)
		if actual != test.expected {
			t.Errorf("normalizeForComparison(%q) = %q; want %q", test.s, actual, test.expected)
		}
	}
}

func TestSuggestAliases(t *testing.T) {
	root := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.WriteFile(filepath.Join(root, ".avm.json"), []byte(`{"run-tv": "echo tv", "build": "go build", "deploy": "sh deploy.sh"}`), 0644); err != nil {
		t.Fatal(err)
	}

	loadOnce = sync.Once{}
	loadErr = nil
	local = nil
	global = nil
	pluginAliases = nil
	loadErr = nil

	tests := []struct {
		query    string
		expected []string
	}{
		{"run:tv", []string{"run-tv"}},
		{"tv-run", []string{"run-tv"}},
		{"buil", []string{"build"}},
		{"deply", []string{"deploy"}},
		{"run", []string{"run-tv"}}, // because it's a substring and len > 3
		{"foo", nil},
	}

	for _, test := range tests {
		actual := SuggestAliases(test.query)
		// Sort both to compare easily
		if len(actual) == 0 && len(test.expected) == 0 {
			continue
		}

		// This is a simple check, since order might differ or multiple suggestions might occur
		foundAll := true
		for _, exp := range test.expected {
			found := false
			for _, act := range actual {
				if act == exp {
					found = true
					break
				}
			}
			if !found {
				foundAll = false
				break
			}
		}

		if !foundAll {
			t.Errorf("SuggestAliases(%q) = %v; want to contain %v", test.query, actual, test.expected)
		}
	}
}

func TestLoadFileLegacyMap(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	content := []byte(`{"build": "go build", "test": "go test"}`)
	if err := os.WriteFile(file, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	aliases, env, tools, err := LoadFileWithEnv(root, ".avm.json")
	if err != nil {
		t.Fatalf("LoadFileWithEnv() error = %v", err)
	}

	if len(aliases) != 2 || env != nil || tools != nil {
		t.Fatalf("expected legacy aliases with nil env/tools, got aliases=%v env=%v tools=%v", aliases, env, tools)
	}

	if aliases["build"] != "go build" {
		t.Fatalf("expected alias %q, got %q", "go build", aliases["build"])
	}
}

func TestLoadFileStructured(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	content := []byte(`{"aliases": {"start": "npm run start"}, "env": {"NODE_ENV": "test", "API_URL": "https://example.com"}}`)
	if err := os.WriteFile(file, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	aliases, env, tools, err := LoadFileWithEnv(root, ".avm.json")
	if err != nil {
		t.Fatalf("LoadFileWithEnv() error = %v", err)
	}

	if aliases == nil || aliases["start"] != "npm run start" {
		t.Fatalf("expected alias not found in structured config: %v", aliases)
	}

	if env["NODE_ENV"] != "test" || env["API_URL"] != "https://example.com" {
		t.Fatalf("expected env values not found in structured config: %v", env)
	}

	if tools == nil || len(tools) != 0 {
		t.Fatalf("expected empty tools map in legacy structured config fixture, got %v", tools)
	}
}

func TestLoadFileStructuredWithTools(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	content := []byte(`{"aliases": {"start": "npm run start"}, "env": {"NODE_ENV": "test"}, "tools": {"node": "20.11.1", "ruby": "3.3.0"}}`)
	if err := os.WriteFile(file, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	aliases, env, tools, err := LoadFileWithEnv(root, ".avm.json")
	if err != nil {
		t.Fatalf("LoadFileWithEnv() error = %v", err)
	}

	if aliases["start"] != "npm run start" {
		t.Fatalf("expected alias not found in structured config: %v", aliases)
	}

	if env["NODE_ENV"] != "test" {
		t.Fatalf("expected env values not found in structured config: %v", env)
	}

	if tools["node"] != "20.11.1" || tools["ruby"] != "3.3.0" {
		t.Fatalf("expected tools values not found in structured config: %v", tools)
	}
}

func TestResolveEnvPrecedence(t *testing.T) {
	root := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	localFile := filepath.Join(root, ".avm.json")
	if err := os.WriteFile(localFile, []byte(`{"aliases": {"start": "node local.js"}, "env": {"NODE_ENV": "local", "API_URL": "https://local.example.com"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".avm.json"), []byte(`{"env": {"NODE_ENV": "global", "GLOBAL_ONLY": "yes"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	loadOnce = sync.Once{}
	loadErr = nil
	local = nil
	global = nil
	localEnv = nil
	globalEnv = nil

	result, err := ResolveEnv()
	if err != nil {
		t.Fatalf("ResolveEnv() error = %v", err)
	}

	expected := map[string]string{
		"NODE_ENV":    "local",
		"API_URL":     "https://local.example.com",
		"GLOBAL_ONLY": "yes",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestResolveToolsPrecedence(t *testing.T) {
	root := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	localFile := filepath.Join(root, ".avm.json")
	if err := os.WriteFile(localFile, []byte(`{"tools": {"node": "21.0.0", "go": "1.22.0"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".avm.json"), []byte(`{"tools": {"node": "20.11.1", "python": "3.12.0"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	loadOnce = sync.Once{}
	loadErr = nil
	local = nil
	global = nil
	localTools = nil
	globalTools = nil

	resolved, err := ResolveTools()
	if err != nil {
		t.Fatalf("ResolveTools() error = %v", err)
	}

	expected := map[string]string{
		"node":   "21.0.0",
		"go":     "1.22.0",
		"python": "3.12.0",
	}
	if !reflect.DeepEqual(resolved, expected) {
		t.Fatalf("expected %v, got %v", expected, resolved)
	}
}

func TestLoadFileLegacyMapAutoMigrates(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	if err := os.WriteFile(file, []byte(`{"build":"go build","test":"go test"}`), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, _, _, err := LoadFileWithEnv(root, ".avm.json"); err != nil {
		t.Fatalf("LoadFileWithEnv() error = %v", err)
	}

	migrated, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read migrated file: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(migrated, &raw); err != nil {
		t.Fatalf("invalid migrated json: %v", err)
	}

	if _, ok := raw["aliases"]; !ok {
		t.Fatalf("expected migrated file to contain aliases key, got: %s", string(migrated))
	}

	if _, ok := raw["build"]; ok {
		t.Fatalf("expected legacy keys to be migrated out, got: %s", string(migrated))
	}
}

func TestMigrateLegacyConfigFunction(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	if err := os.WriteFile(file, []byte(`{"build":"go build"}`), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := MigrateLegacyConfig(root, ".avm.json"); err != nil {
		t.Fatalf("MigrateLegacyConfig() error = %v", err)
	}

	migrated, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read migrated file: %v", err)
	}

	if !json.Valid(migrated) {
		t.Fatalf("migrated file is not valid json: %s", string(migrated))
	}
}

func TestUseToolWritesStructuredConfig(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	oldWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.WriteFile(file, []byte(`{"build":"go build","test":"go test"}`), 0644); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	loadOnce = sync.Once{}
	local = nil
	global = nil
	localEnv = nil
	globalEnv = nil
	localTools = nil
	globalTools = nil
	pluginAliases = nil
	loadErr = nil

	alias := &Alias{
		Root:      root,
		LocalFile: ".avm.json",
	}
	if err := UseTool(alias, "node", "20.11.1"); err != nil {
		t.Fatalf("UseTool() error = %v", err)
	}

	rawBytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read updated file: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(rawBytes, &data); err != nil {
		t.Fatalf("invalid updated json: %v", err)
	}

	if _, ok := data["aliases"]; !ok {
		t.Fatalf("expected aliases key to exist after migration, got: %s", string(rawBytes))
	}

	toolsValue, ok := data["tools"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tools key to be a map, got: %v", data["tools"])
	}

	if toolsValue["node"] != "20.11.1" {
		t.Fatalf("expected node tool version in tools map, got: %v", data["tools"])
	}
}

func TestLoadFileRejectsInvalidEnvKey(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	if err := os.WriteFile(file, []byte(`{"env":{"BAD;echo hacked":"1"}}`), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, _, _, err := LoadFileWithEnv(root, ".avm.json"); err == nil {
		t.Fatal("expected invalid env key error")
	}
}

func TestLoadFileStructuredParseErrorDoesNotFallbackToLegacy(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".avm.json")

	if err := os.WriteFile(file, []byte(`{"aliases":{"start":123}}`), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, _, _, err := LoadFileWithEnv(root, ".avm.json"); err == nil {
		t.Fatal("expected structured parse error")
	}
}

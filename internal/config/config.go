package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Alias struct {
	Root      string
	LocalFile string
	Global    bool
}

type ConfigFile struct {
	Aliases map[string]string `json:"aliases"`
	Env     map[string]string `json:"env"`
	Tools   map[string]string `json:"tools"`
}

func LoadFile(root string, localFile string) (map[string]string, error) {
	aliases, _, _, err := LoadFileWithEnv(root, localFile)
	if err != nil {
		return nil, err
	}

	return aliases, nil
}

func LoadFileWithEnv(root string, localFile string) (map[string]string, map[string]string, map[string]string, error) {
	file := filepath.Join(root, localFile)
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	aliases, env, tools, structured, err := parseConfigFile(data)
	if err != nil {
		return nil, nil, nil, err
	}

	// Backward-compatible auto-migration:
	// legacy flat maps are transparently rewritten as structured configs.
	if !structured && aliases != nil {
		if err := migrateLegacyConfig(root, localFile, aliases); err != nil {
			// Keep behavior resilient if migration is not possible (e.g., read-only config dir).
			// Alias resolution will still work from in-memory parse.
		}
	}

	return aliases, env, tools, nil
}

func parseConfigFile(data []byte) (map[string]string, map[string]string, map[string]string, bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, nil, false, fmt.Errorf("invalid config json: %w", err)
	}

	_, hasAliases := raw["aliases"]
	_, hasEnv := raw["env"]
	_, hasTools := raw["tools"]
	if hasAliases || hasEnv || hasTools {
		cfg, err := parseStructuredConfig(data)
		if err != nil {
			return nil, nil, nil, true, err
		}
		normalizeConfigFile(cfg)
		if err := validateConfigFile(cfg); err != nil {
			return nil, nil, nil, true, err
		}
		return cfg.Aliases, cfg.Env, cfg.Tools, true, nil
	}

	var aliases map[string]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, nil, nil, false, err
	}

	return aliases, nil, nil, false, nil
}

func parseStructuredConfig(data []byte) (*ConfigFile, error) {
	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func IsStructuredConfig(root string, localFile string) (bool, error) {
	file := filepath.Join(root, localFile)
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, err
	}

	_, hasAliases := raw["aliases"]
	_, hasEnv := raw["env"]
	_, hasTools := raw["tools"]
	return hasAliases || hasEnv || hasTools, nil
}

func MigrateLegacyConfig(root string, localFile string) error {
	file := filepath.Join(root, localFile)
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	aliases, _, _, structured, err := parseConfigFile(data)
	if err != nil {
		return err
	}

	if structured || aliases == nil {
		return nil
	}

	return migrateLegacyConfig(root, localFile, aliases)
}

func migrateLegacyConfig(root string, localFile string, aliases map[string]string) error {
	// Migrate legacy flat-map config to structured format.
	return SaveConfig(root, localFile, aliases, nil, nil, true)
}

func normalizeConfigFile(cfg *ConfigFile) {
	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}
	if cfg.Env == nil {
		cfg.Env = map[string]string{}
	}
	if cfg.Tools == nil {
		cfg.Tools = map[string]string{}
	}
}

func validateConfigFile(cfg *ConfigFile) error {
	for key := range cfg.Env {
		if !IsValidEnvKey(key) {
			return fmt.Errorf("invalid env key %q: must match [A-Za-z_][A-Za-z0-9_]*", key)
		}
	}
	return nil
}

func IsValidEnvKey(key string) bool {
	if key == "" {
		return false
	}

	for i, r := range key {
		if i == 0 {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_' {
				continue
			}
			return false
		}

		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func SaveConfig(root string, localFile string, aliases map[string]string, env map[string]string, tools map[string]string, structured bool) error {
	file := filepath.Join(root, localFile)

	if aliases == nil {
		aliases = map[string]string{}
	}

	if !structured {
		data, err := json.MarshalIndent(aliases, "", "  ")
		if err != nil {
			return err
		}
		return writeFileAtomic(file, data, 0644)
	}

	cfg := ConfigFile{
		Aliases: aliases,
	}

	if len(env) > 0 {
		if err := validateConfigFile(&ConfigFile{Env: env}); err != nil {
			return err
		}
		cfg.Env = env
	}
	if len(tools) > 0 {
		cfg.Tools = tools
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(file, data, 0644)
}

func SaveAliases(root string, localFile string, aliases map[string]string) error {
	file := filepath.Join(root, localFile)
	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(file, data, 0644)
}

func GetConfigPath(root string, localFile string) string {
	return filepath.Join(root, localFile)
}

func IsFileExists(root string, localFile string) (bool, error) {
	file := filepath.Join(root, localFile)
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateDefaultConfig(root string, localFile string) error {
	file := filepath.Join(root, localFile)
	data := []byte("{}\n")
	return writeFileAtomic(file, data, 0644)
}

func writeFileAtomic(file string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if info, err := os.Stat(file); err == nil {
		perm = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(file)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, bytesReader(data)); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, file); err != nil {
		return err
	}
	cleanup = false
	return syncDir(dir)
}

func bytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer d.Close()
	_ = d.Sync()
	return nil
}

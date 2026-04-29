package plugin

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	cache     map[string]ResolvedAlias
	cacheOnce sync.Once
	mu        sync.Mutex
)

const (
	PluginTimeout = 500 * time.Millisecond
	GlobalTimeout = 1 * time.Second
	WorkerCount   = 4
)

func GetPluginDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".avm", "plugins")
}

func LoadAllPlugins(cwd string) (map[string]ResolvedAlias, error) {
	mu.Lock()
	if cache != nil {
		defer mu.Unlock()
		return cache, nil
	}
	mu.Unlock()

	pluginDir := GetPluginDir()
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, nil
	}

	dirs, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	results := make(map[string]ResolvedAlias)
	var wg sync.WaitGroup
	jobs := make(chan string, len(dirs))
	resultsChan := make(chan map[string]ResolvedAlias, len(dirs))

	ctx, cancel := context.WithTimeout(context.Background(), GlobalTimeout)
	defer cancel()

	// Worker pool
	for w := 1; w <= WorkerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dirName := range jobs {
				pluginPath := filepath.Join(pluginDir, dirName)
				aliases := loadPlugin(ctx, pluginPath, cwd)
				if aliases != nil {
					resultsChan <- aliases
				}
			}
		}()
	}

	for _, d := range dirs {
		if d.IsDir() || (d.Type()&os.ModeSymlink != 0) {
			jobs <- d.Name()
		}
	}
	close(jobs)

	wg.Wait()
	close(resultsChan)

	for res := range resultsChan {
		for k, v := range res {
			// Precedence check: if multiple plugins provide the same key,
			// the first one processed wins (or we could define a priority).
			if _, exists := results[k]; !exists {
				results[k] = v
			}
		}
	}

	mu.Lock()
	cache = results
	mu.Unlock()

	return results, nil
}

func loadPlugin(ctx context.Context, pluginPath, cwd string) map[string]ResolvedAlias {
	manifestPath := filepath.Join(pluginPath, "plugin.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil
	}

	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil
	}

	exportHook := filepath.Join(pluginPath, "bin", "export-aliases")
	if _, err := os.Stat(exportHook); err != nil {
		return nil
	}

	// health-check
	healthHook := filepath.Join(pluginPath, "bin", "health-check")
	if _, err := os.Stat(healthHook); err == nil {
		hCtx, hCancel := context.WithTimeout(ctx, PluginTimeout)
		defer hCancel()
		cmd := exec.CommandContext(hCtx, healthHook, "--dir", cwd)
		if err := cmd.Run(); err != nil {
			return nil
		}
	}

	eCtx, eCancel := context.WithTimeout(ctx, PluginTimeout)
	defer eCancel()

	cmd := exec.CommandContext(eCtx, exportHook, "--dir", cwd)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var response ExportResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil
	}

	aliases := make(map[string]ResolvedAlias)
	sectionLabel := manifest.SectionLabel
	if sectionLabel == "" {
		sectionLabel = manifest.Name
	}

	for k, v := range response.Aliases {
		resolved := ResolvedAlias{
			PluginName:  manifest.Name,
			SectionName: sectionLabel,
		}

		switch val := v.(type) {
		case string:
			resolved.Command = val
		case map[string]interface{}:
			if cmd, ok := val["command"].(string); ok {
				resolved.Command = cmd
			}
			if desc, ok := val["description"].(string); ok {
				resolved.Description = desc
			}
		}

		if resolved.Command != "" {
			aliases[k] = resolved
		}
	}

	return aliases
}

func GetManifest(pluginName string) (*Manifest, error) {
	pluginPath := filepath.Join(GetPluginDir(), pluginName)
	
	// Try bin/describe first
	describeHook := filepath.Join(pluginPath, "bin", "describe")
	if _, err := os.Stat(describeHook); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), PluginTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, describeHook)
		output, err := cmd.Output()
		if err == nil {
			var manifest Manifest
			if err := json.Unmarshal(output, &manifest); err == nil {
				return &manifest, nil
			}
		}
	}

	// Fallback to plugin.json
	manifestPath := filepath.Join(pluginPath, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func InstallPlugin(source string) error {
	pluginDir := GetPluginDir()
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// Simple implementation: if it starts with http or git@, it's git.
	// Otherwise, check if it's a local directory and symlink it.
	
	if isGitURL(source) {
		name := filepath.Base(source)
		name = strings.TrimSuffix(name, ".git")
		target := filepath.Join(pluginDir, name)
		cmd := exec.Command("git", "clone", source, target)
		return cmd.Run()
	}

	// Local path
	absPath, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	name := filepath.Base(absPath)
	target := filepath.Join(pluginDir, name)
	
	return os.Symlink(absPath, target)
}

func isGitURL(s string) bool {
	return strings.HasPrefix(s, "http") || strings.HasPrefix(s, "git@")
}

func UpdatePlugin(name string) error {
	pluginPath := filepath.Join(GetPluginDir(), name)
	if _, err := os.Stat(filepath.Join(pluginPath, ".git")); err == nil {
		cmd := exec.Command("git", "-C", pluginPath, "pull")
		return cmd.Run()
	}
	return nil
}

func RemovePlugin(name string) error {
	pluginPath := filepath.Join(GetPluginDir(), name)
	return os.RemoveAll(pluginPath)
}

package tooling

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/PrajaNova/avm/internal/config"
)

type ToolProvider interface {
	Name() string
	IsInstalled(tool string, version string) bool
	InstalledVersions(tool string) ([]string, error)
	Install(tool string, version string) error
	Uninstall(tool string, version string) error
	ToolExecutablePath(tool string, version string) (string, error)
}

type ResolvedTool = config.ResolvedTool

type ToolEnv map[string]string

var providerRegistry = map[string]ToolProvider{}

func RegisterProvider(provider ToolProvider) {
	providerRegistry[provider.Name()] = provider
}

func GetProvider(tool string) (ToolProvider, bool) {
	provider, ok := providerRegistry[tool]
	return provider, ok
}

func ResolveTools(cwd string) (map[string]ResolvedTool, error) {
	return config.ResolveToolsWithSourceFor(cwd)
}

func InstallTool(tool string, version string) error {
	provider, ok := providerRegistry[tool]
	if !ok {
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	return provider.Install(tool, version)
}

func UninstallTool(tool string, version string) error {
	provider, ok := providerRegistry[tool]
	if !ok {
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	return provider.Uninstall(tool, version)
}

func ResolveToolEnv(cwd string) (ToolEnv, error) {
	_ = cwd
	resolved, err := ResolveTools(cwd)
	if err != nil {
		return nil, err
	}

	existingPath := os.Getenv("PATH")
	existingPathParts := splitPath(existingPath)
	seen := map[string]bool{}
	for _, part := range existingPathParts {
		if part == "" {
			continue
		}
		seen[filepath.Clean(part)] = true
	}

	var injectedParts []string
	var tools []string
	for tool := range resolved {
		tools = append(tools, tool)
	}
	sort.Strings(tools)

	for _, tool := range tools {
		selection := resolved[tool]

		provider, ok := providerRegistry[tool]
		if !ok {
			continue
		}

		if !provider.IsInstalled(tool, selection.Version) {
			continue
		}

		execPath, err := provider.ToolExecutablePath(tool, selection.Version)
		if err != nil {
			continue
		}

		binPath := filepath.Dir(execPath)
		if binPath == "" {
			continue
		}

		normalized := filepath.Clean(binPath)
		if seen[normalized] {
			continue
		}
		injectedParts = append(injectedParts, normalized)
		seen[normalized] = true
	}

	if len(injectedParts) == 0 {
		return ToolEnv{}, nil
	}

	newPath := strings.Join(append(injectedParts, existingPathParts...), string(os.PathListSeparator))
	return ToolEnv{"PATH": newPath}, nil
}

func InstalledVersions(tool string) ([]string, error) {
	provider, ok := providerRegistry[tool]
	if !ok {
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}
	return provider.InstalledVersions(tool)
}

func KnownTools() []string {
	tools := make([]string, 0, len(providerRegistry))
	for name := range providerRegistry {
		tools = append(tools, name)
	}
	sort.Strings(tools)
	return tools
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, string(os.PathListSeparator))
	return parts
}

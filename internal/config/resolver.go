package config

import (
	"avm/internal/plugin"
	"os"
	"sort"
	"strings"
	"sync"
)

var local map[string]string
var global map[string]string

var localEnv map[string]string
var globalEnv map[string]string
var localTools map[string]string
var globalTools map[string]string

var pluginAliases map[string]plugin.ResolvedAlias
var loadOnce sync.Once
var loadErr error

func GetAliases() error {
	loadOnce.Do(func() {
		loadErr = loadAliasesInternal()
	})
	return loadErr
}

func loadAliasesInternal() error {
	var err error

	local, localEnv, localTools, err = LoadFileWithEnv(".", ".avm.json")
	if err != nil {
		return err
	}

	home := os.Getenv("HOME")
	global, globalEnv, globalTools, err = LoadFileWithEnv(home, ".avm.json")
	if err != nil {
		return err
	}

	// Load plugins
	cwd, _ := os.Getwd()
	pluginAliases, err = plugin.LoadAllPlugins(cwd)
	if err != nil {
		// Log error but don't fail core resolution
		// fmt.Fprintf(os.Stderr, "Error loading plugins: %v\n", err)
	}

	return nil
}

func ResolveEnv() (map[string]string, error) {
	if err := GetAliases(); err != nil {
		return nil, err
	}

	resolved := map[string]string{}

	for key, value := range globalEnv {
		resolved[key] = value
	}

	for key, value := range localEnv {
		resolved[key] = value
	}

	return resolved, nil
}

func ResolveTools() (map[string]string, error) {
	if err := GetAliases(); err != nil {
		return nil, err
	}

	resolved := map[string]string{}

	for tool, version := range globalTools {
		resolved[tool] = version
	}

	for tool, version := range localTools {
		resolved[tool] = version
	}

	return resolved, nil
}

type ResolvedTool struct {
	Version string
	Source  string
}

func ResolveToolsWithSource() (map[string]ResolvedTool, error) {
	if err := GetAliases(); err != nil {
		return nil, err
	}

	resolved := map[string]ResolvedTool{}

	for tool, version := range globalTools {
		resolved[tool] = ResolvedTool{
			Version: version,
			Source:  "global",
		}
	}

	for tool, version := range localTools {
		resolved[tool] = ResolvedTool{
			Version: version,
			Source:  "local",
		}
	}

	return resolved, nil
}

func ResolveToolWithSource(key string) (string, bool, string, error) {
	if err := GetAliases(); err != nil {
		return "", false, "", err
	}

	if localTools != nil {
		if version, exists := localTools[key]; exists {
			return version, true, "local", nil
		}
	}

	if globalTools != nil {
		if version, exists := globalTools[key]; exists {
			return version, true, "global", nil
		}
	}

	return "", false, "", nil
}

func ResolveWithSource(key string) (string, bool, string, error) {
	if err := GetAliases(); err != nil {
		return "", false, "", err
	}

	if local != nil {
		if val, exists := local[key]; exists {
			return val, true, "local", nil
		}
	}

	if global != nil {
		if val, exists := global[key]; exists {
			return val, true, "global", nil
		}
	}

	if pluginAliases != nil {
		if res, exists := pluginAliases[key]; exists {
			return res.Command, true, "plugin:" + res.PluginName, nil
		}
	}

	return "", false, "", nil
}

func SuggestAliases(query string) []string {
	if err := GetAliases(); err != nil {
		return nil
	}

	allKeys := make(map[string]bool)
	if local != nil {
		for k := range local {
			allKeys[k] = true
		}
	}
	if global != nil {
		for k := range global {
			allKeys[k] = true
		}
	}
	if pluginAliases != nil {
		for k := range pluginAliases {
			allKeys[k] = true
		}
	}

	var suggestions []string
	queryNormalized := normalizeForComparison(query)

	for k := range allKeys {
		// Exact distance check
		if levenshteinDistance(query, k) <= 2 {
			suggestions = append(suggestions, k)
			continue
		}

		// Normalized comparison (handles different separators and permutations)
		kNormalized := normalizeForComparison(k)
		if queryNormalized == kNormalized {
			suggestions = append(suggestions, k)
			continue
		}

		// If one is a substring of another (using normalized versions)
		if (strings.Contains(kNormalized, queryNormalized) || strings.Contains(queryNormalized, kNormalized)) && (len(queryNormalized) > 3 || len(kNormalized) > 3) {
			suggestions = append(suggestions, k)
			continue
		}

		// If one is a substring of another and they are long enough
		if (strings.Contains(k, query) || strings.Contains(query, k)) && (len(query) > 3 || len(k) > 3) {
			suggestions = append(suggestions, k)
			continue
		}
	}

	return suggestions
}

func normalizeForComparison(s string) string {
	// Replace common separators with spaces, then split and sort parts
	f := func(r rune) bool {
		return r == '-' || r == ':' || r == '_' || r == '.'
	}
	parts := strings.FieldsFunc(s, f)
	sort.Strings(parts)
	return strings.Join(parts, "-")
}

func levenshteinDistance(s, t string) int {
	m := len(s)
	n := len(t)
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for j := 1; j <= n; j++ {
		for i := 1; i <= m; i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j] + 1
				if d[i][j-1]+1 < min {
					min = d[i][j-1] + 1
				}
				if d[i-1][j-1]+1 < min {
					min = d[i-1][j-1] + 1
				}
				d[i][j] = min
			}
		}
	}
	return d[m][n]
}

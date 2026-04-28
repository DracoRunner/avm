package config

import (
	"os"
	"sort"
	"strings"
)

var local map[string]string
var global map[string]string
var loaded bool

func GetAliases() error {
	if loaded {
		return nil
	}

	var err error

	local, err = LoadFile(".", ".avm.json")
	if err != nil {
		return err
	}

	home := os.Getenv("HOME")
	global, err = LoadFile(home, ".avm.json")
	if err != nil {
		return err
	}

	loaded = true
	return nil
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

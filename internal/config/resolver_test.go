package config

import (
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
	// Mock local and global aliases
	local = map[string]string{
		"run-tv": "echo tv",
		"build":  "go build",
	}
	global = map[string]string{
		"deploy": "sh deploy.sh",
	}
	loaded = true

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

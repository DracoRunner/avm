package config

import (
	"os"
)

var local map[string]string
var global map[string]string

func GetAliases() error {
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

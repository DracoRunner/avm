package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Alias struct {
	Root      string
	LocalFile string
	Global    bool
}

func LoadFile(root string, localFile string) (map[string]string, error) {
	file := filepath.Join(root, localFile)
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var aliases map[string]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, err
	}

	return aliases, nil
}

func SaveAliases(root string, localFile string, aliases map[string]string) error {
	file := filepath.Join(root, localFile)
	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
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
	return os.WriteFile(file, data, 0644)
}

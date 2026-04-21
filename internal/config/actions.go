package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Init(alias *Alias) error {
	root := alias.Root
	localFile := alias.LocalFile

	exists, err := IsFileExists(root, localFile)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("%s already exists", GetConfigPath(root, localFile))
	}

	return CreateDefaultConfig(root, localFile)
}

func Add(alias *Alias, key string, value string) error {
	root := alias.Root
	localFile := alias.LocalFile

	if !alias.Global {
		exists, err := IsFileExists(root, localFile)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("no %s found in current directory. Run 'avm init' first", localFile)
		}
	} else {
		// Ensure global file exists
		exists, err := IsFileExists(root, localFile)
		if err != nil {
			return err
		}
		if !exists {
			if err := CreateDefaultConfig(root, localFile); err != nil {
				return err
			}
		}
	}

	aliases, err := LoadFile(root, localFile)
	if err != nil {
		return err
	}

	if aliases == nil {
		aliases = map[string]string{}
	}

	aliases[key] = value

	return SaveAliases(root, localFile, aliases)
}

func List(alias *Alias) error {
	if err := GetAliases(); err != nil {
		return err
	}

	hasLocal := local != nil && len(local) > 0
	hasGlobal := global != nil && len(global) > 0

	if !hasLocal && !hasGlobal {
		fmt.Println("No aliases configured.")
		fmt.Println()
		fmt.Println("Get started:")
		fmt.Println("  avm init                          # Create local .avm.json")
		fmt.Println("  avm add start \"npm run dev\"       # Add local alias")
		fmt.Println("  avm add -g cleanup \"rm -rf tmp\"   # Add global alias")
		return nil
	}

	fmt.Println()

	if hasLocal {
		fmt.Println(color.CyanString("Local aliases (.avm.json):"))
		for key, value := range local {
			if global != nil {
				if _, ok := global[key]; ok {
					fmt.Printf("  %s → %s  %s\n", color.GreenString(key), value, color.YellowString("[override]"))
					continue
				}
			}
			fmt.Printf("  %s → %s\n", color.GreenString(key), value)
		}
		fmt.Println()
	}

	if hasGlobal {
		fmt.Println(color.CyanString("Global aliases (~/.avm.json):"))
		for key, value := range global {
			if local != nil {
				if _, ok := local[key]; ok {
					continue
				}
			}
			fmt.Printf("  %s → %s\n", color.GreenString(key), value)
		}
		fmt.Println()
	}

	return nil
}

func Remove(alias *Alias, key string) error {
	root := alias.Root
	localFile := alias.LocalFile

	exists, err := IsFileExists(root, localFile)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no %s found", localFile)
	}

	aliases, err := LoadFile(root, localFile)
	if err != nil {
		return err
	}

	if aliases == nil {
		aliases = map[string]string{}
	}

	if _, exists := aliases[key]; !exists {
		return fmt.Errorf("alias '%s' not found", key)
	}

	delete(aliases, key)

	return SaveAliases(root, localFile, aliases)
}

func Which(alias *Alias, key string) error {
	val, found, source, err := ResolveWithSource(key)
	if err != nil {
		return err
	}

	if !found {
		fmt.Fprintf(os.Stderr, "No alias found for '%s', will run as normal command\n", key)
		return nil
	}

	if source == "local" {
		fmt.Fprintf(os.Stderr, "%s alias '%s':\n", color.GreenString("Local"), key)
	} else {
		fmt.Fprintf(os.Stderr, "%s alias '%s':\n", color.BlueString("Global"), key)
	}
	fmt.Printf("Command: %s\n", val)

	return nil
}

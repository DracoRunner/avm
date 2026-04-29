package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

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
	hasPlugins := pluginAliases != nil && len(pluginAliases) > 0

	if !hasLocal && !hasGlobal && !hasPlugins {
		fmt.Println("No aliases configured.")
		fmt.Println()
		fmt.Println("Get started:")
		fmt.Println("  avm init                          # Create local .avm.json")
		fmt.Println("  avm add start \"npm run dev\"       # Add local alias")
		fmt.Println("  avm add -g cleanup \"rm -rf tmp\"   # Add global alias")
		fmt.Println("  avm plugin add <url>              # Add a plugin")
		return nil
	}

	fmt.Println()

	if hasLocal {
		fmt.Println(color.CyanString("Local aliases (.avm.json):"))
		var localKeys []string
		for key := range local {
			localKeys = append(localKeys, key)
		}
		sort.Strings(localKeys)
		for _, key := range localKeys {
			value := local[key]
			if global != nil {
				if _, ok := global[key]; ok {
					fmt.Printf("  %s → %s  %s\n", color.GreenString(key), value, color.YellowString("[override global]"))
					continue
				}
			}
			fmt.Printf("  %s → %s\n", color.GreenString(key), value)
		}
		fmt.Println()
	}

	if hasGlobal {
		fmt.Println(color.CyanString("Global aliases (~/.avm.json):"))
		var globalKeys []string
		for key := range global {
			globalKeys = append(globalKeys, key)
		}
		sort.Strings(globalKeys)
		for _, key := range globalKeys {
			value := global[key]
			if local != nil {
				if _, ok := local[key]; ok {
					continue
				}
			}
			fmt.Printf("  %s → %s\n", color.GreenString(key), value)
		}
		fmt.Println()
	}

	if hasPlugins {
		// Group by section
		sections := make(map[string]map[string]string)
		for key, res := range pluginAliases {
			// Precedence check: don't show if shadowed by local or global
			if _, exists := local[key]; exists {
				continue
			}
			if _, exists := global[key]; exists {
				continue
			}

			if _, ok := sections[res.SectionName]; !ok {
				sections[res.SectionName] = make(map[string]string)
			}
			sections[res.SectionName][key] = res.Command
		}

		var sectionNames []string
		for section := range sections {
			sectionNames = append(sectionNames, section)
		}
		sort.Strings(sectionNames)

		for _, section := range sectionNames {
			aliases := sections[section]
			fmt.Println(color.CyanString(fmt.Sprintf("Plugin: %s:", section)))
			var pluginKeys []string
			for key := range aliases {
				pluginKeys = append(pluginKeys, key)
			}
			sort.Strings(pluginKeys)
			for _, key := range pluginKeys {
				value := aliases[key]
				fmt.Printf("  %s → %s\n", color.GreenString(key), value)
			}
			fmt.Println()
		}
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
	} else if strings.HasPrefix(source, "plugin:") {
		pluginName := strings.TrimPrefix(source, "plugin:")
		fmt.Fprintf(os.Stderr, "%s plugin alias '%s':\n", color.BlueString(strings.Title(pluginName)), key)
	} else {
		fmt.Fprintf(os.Stderr, "%s alias '%s':\n", color.BlueString("Global"), key)
	}
	fmt.Printf("Command: %s\n", val)

	return nil
}

package cmd

import (
	"os"
	"sort"
	"strings"

	"github.com/dannyben/config/format"

	"github.com/spf13/cobra"
)

type keyCompletionMode int

const (
	completeLeafKeys keyCompletionMode = iota
	completeContainerKeys
	completeScalarArrayKeys
	completeAnyKeys
)

func completeKeyArg(configFile *string, position int, mode keyCompletionMode) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != position {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		keys, err := completionKeys(*configFile, mode)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		matches := make([]string, 0, len(keys))
		for _, key := range keys {
			if strings.HasPrefix(key, toComplete) {
				matches = append(matches, key)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	}
}

func completionKeys(configFile string, mode keyCompletionMode) ([]string, error) {
	leafKeys, err := listKeys(configFile, "")
	if err != nil {
		return nil, err
	}

	switch mode {
	case completeLeafKeys:
		return leafKeys, nil
	case completeContainerKeys:
		return parentKeys(leafKeys), nil
	case completeScalarArrayKeys:
		return scalarArrayKeys(leafKeys), nil
	case completeAnyKeys:
		return unionKeys(leafKeys, parentKeys(leafKeys)), nil
	default:
		return leafKeys, nil
	}
}

func listKeys(configFile, key string) ([]string, error) {
	entries, err := listEntries(configFile, key)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return keys, nil
}

func parentKeys(keys []string) []string {
	seen := make(map[string]bool)
	for _, key := range keys {
		parts, ok := splitListKey(key)
		if !ok {
			continue
		}
		for i := 1; i < len(parts); i++ {
			seen[joinListKey(parts[:i])] = true
		}
	}
	return sortedKeys(seen)
}

func scalarArrayKeys(keys []string) []string {
	seen := make(map[string]bool)
	for _, key := range keys {
		parts, ok := splitListKey(key)
		if !ok || len(parts) < 2 {
			continue
		}
		if !isNonNegativeInteger(parts[len(parts)-1]) {
			continue
		}
		seen[joinListKey(parts[:len(parts)-1])] = true
	}
	return sortedKeys(seen)
}

func unionKeys(first, second []string) []string {
	seen := make(map[string]bool)
	for _, key := range first {
		seen[key] = true
	}
	for _, key := range second {
		seen[key] = true
	}
	return sortedKeys(seen)
}

func sortedKeys(seen map[string]bool) []string {
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func splitListKey(key string) ([]string, bool) {
	var parts []string
	var current strings.Builder
	for i := 0; i < len(key); i++ {
		if key[i] != '.' {
			current.WriteByte(key[i])
			continue
		}
		if i+1 < len(key) && key[i+1] == '.' {
			current.WriteByte('.')
			i++
			continue
		}
		if current.Len() == 0 {
			return nil, false
		}
		parts = append(parts, current.String())
		current.Reset()
	}
	if current.Len() == 0 {
		return nil, false
	}
	parts = append(parts, current.String())
	return parts, true
}

func joinListKey(parts []string) string {
	escaped := make([]string, len(parts))
	for i, part := range parts {
		escaped[i] = strings.ReplaceAll(part, ".", "..")
	}
	return strings.Join(escaped, ".")
}

func isNonNegativeInteger(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func listEntries(configFile, key string) ([]format.Entry, error) {
	resolved, err := resolveConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	source, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}
	doc, _, err := format.ResolveSource(resolved, source)
	if err != nil {
		return nil, err
	}
	return doc.List(string(source), key)
}

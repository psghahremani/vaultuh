package command

import (
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/psghahremani/vaultuh/encryption"
	"github.com/psghahremani/vaultuh/utility"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"path/filepath"
	"sort"
	"strings"
)

var readCommand = &cobra.Command{
	Use:  "read [path]",
	Args: cobra.ExactArgs(1),
	Run:  read,
}

var mustPrintValues bool

func init() {
	rootCommand.AddCommand(readCommand)
	readCommand.Flags().BoolVarP(&mustPrintValues, "print", "p", false, "Print the values instead of copying them into the clipboard")
}

func read(_ *cobra.Command, arguments []string) {
	vaultFilePath, err := filepath.Abs(arguments[0])
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not find the absolute path for the vault file: %w", err)))
		return
	}
	password, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your password")
	gpg := encryption.GPG{
		FilePath: vaultFilePath,
		Password: password,
	}

	var vaultContent map[string]any
	data, err := gpg.DecryptFromFileIntoMemory()
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not decrypt the vault file: %w", err)))
		return
	}
	err = json.Unmarshal(data, &vaultContent)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not decode the decrypted vault file (expecting JSON): %w", err)))
		return
	}
	err = validateObject(vaultContent)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("the decrypted vault file has invalid content: %w", err)))
		return
	}

	fields := extractFieldsAsPaths(vaultContent, false)
	keys := make([]string, 0, len(fields))
	for field := range fields {
		keys = append(keys, field)
	}
	sort.SliceStable(
		keys,
		func(i, j int) bool {
			return keys[i] < keys[j]
		},
	)
	keys = append([]string{"_"}, keys...)

	for {
		selectedField, _ := pterm.DefaultInteractiveSelect.WithOptions(keys).Show("Select a field")
		value := fields[selectedField]

		if selectedField == "_" {
			temp, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your password again to show the hidden fields")
			if temp != password {
				pterm.Error.Println("The entered password is invalid.")
				continue
			}

			fields = extractFieldsAsPaths(vaultContent, true)
			keys = make([]string, 0, len(fields))
			for key := range fields {
				keys = append(keys, key)
			}
			sort.SliceStable(
				keys,
				func(i, j int) bool {
					return keys[i] < keys[j]
				},
			)
			continue
		}

		if mustPrintValues {
			pterm.Info.Printfln("Here's the value for the field you selected: %s", value)
			// TODO: Handle INTERRUPT signals.
			break
		}

		err = clipboard.WriteAll(fmt.Sprint(value))
		if err != nil {
			pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not write into the clipboard: %w", err)))
			return
		}
		pterm.Info.Println("The value has been copied into your clipboard.")

		// TODO: Handle INTERRUPT signals.
		break
	}
}

func extractFieldsAsPaths(object map[string]any, listHiddenFields bool) map[string]any {
	paths := map[string]any{}
	extractPathsRecursively(object, "", listHiddenFields, paths)
	return paths
}

func extractPathsRecursively(object map[string]any, path string, listHiddenFields bool, allPaths map[string]any) {
	for key, value := range object {
		if strings.HasPrefix(key, "_") {
			if !listHiddenFields {
				continue
			}
			key = key[1:]
		}
		currentPath := path
		if currentPath == "" {
			currentPath = key
		} else {
			currentPath = fmt.Sprintf(
				"%s ☐ %s",
				currentPath,
				key,
			)
		}
		innerObject, isObject := value.(map[string]any)
		if isObject {
			extractPathsRecursively(innerObject, currentPath, listHiddenFields, allPaths)
		} else {
			allPaths[currentPath] = value
		}
	}
}

package command

import (
	"encoding/json"
	"fmt"
	"github.com/psghahremani/vaultuh/encryption"
	"github.com/psghahremani/vaultuh/utility"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"path/filepath"
)

var createCommand = &cobra.Command{
	Use: "create",
	Run: create,
}

func init() {
	rootCommand.AddCommand(createCommand)
}

func create(_ *cobra.Command, _ []string) {
	vaultFilePath, err := filepath.Abs("./vault")
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not find the absolute path for the vault file: %w", err)))
		return
	}
	password, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter the desired password for your vault")
	gpg := encryption.GPG{
		FilePath: vaultFilePath,
		Password: password,
	}

	sampleContent, _ := json.Marshal(
		map[string]any{
			"Personal": map[string]any{
				"My Favorite Website": map[string]any{
					"Username": "poop",
					"Email":    "me@myself.com",
					"Password": "p@$$w0rd",
				},
				"My Favorite Website 2": map[string]any{
					"_Secret Social Security Number": "1234",
					"Password":                       "p@$$w0rd",
				},
			},
			"Work": map[string]any{
				"GitLab": map[string]any{
					"URL":      "gitlab.private_company.org",
					"Email":    "me@work.com",
					"Password": "c0de4life",
				},
			},
			"_Secret Naughty Things": map[string]any{
				"HotMommies": map[string]any{
					"Email":    "me@xxx.com",
					"Password": "w@nk3r69",
				},
			},
		},
	)

	err = gpg.EncryptFromMemoryIntoFile(sampleContent)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not encrypt the vault content: %w", err)))
		return
	}
	pterm.Info.Printfln("A new vault file (with some sample content) was created (%s).", vaultFilePath)
}

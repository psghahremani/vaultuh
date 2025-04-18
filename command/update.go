package command

import (
	"encoding/json"
	"fmt"
	"github.com/psghahremani/vaultuh/encryption"
	"github.com/psghahremani/vaultuh/utility"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var updateCommand = &cobra.Command{
	Use: "update",
	Run: nil,
}

func init() {
	updateCommand.AddCommand(updateDataCommand)
	updateCommand.AddCommand(updatePasswordCommand)
	rootCommand.AddCommand(updateCommand)
}

var updateDataCommand = &cobra.Command{
	Use:  "data [path]",
	Args: cobra.ExactArgs(1),
	Run:  updateData,
}

func updateData(_ *cobra.Command, arguments []string) {
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

	data, err := gpg.DecryptFromFileIntoMemory()
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not decrypt the vault file: %w", err)))
		return
	}
	var contents map[string]any
	err = json.Unmarshal(data, &contents)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not decode the decrypted vault file (expecting JSON): %w", err)))
		return
	}
	data, err = json.MarshalIndent(contents, "", "  ")
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("the decrypted vault file has invalid content: %w", err)))
		return
	}

	tempFilePath := fmt.Sprintf("%s/vaultuh.data", os.TempDir())
	tempFile, err := os.OpenFile(
		tempFilePath,
		os.O_RDWR|os.O_CREATE,
		0b110000000,
	)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not create a temporary file: %w", err)))
		return
	}
	defer func() {
		_, err := utility.DeleteFile(tempFilePath)
		if err != nil {
			pterm.Warning.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not remove the previously created temporary file: %w", err)))
			pterm.Warning.Printfln("Please remove it manually (%s).", tempFilePath)
		}
	}()

	_, err = tempFile.Write(data)
	if err != nil {
		_ = tempFile.Close()
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not write into the temporary file: %w", err)))
		return
	}

	editorName := os.Getenv("EDITOR")
	if editorName == "" {
		editorName = "vi"
	}
	for {
		command := exec.Command(editorName, tempFilePath)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err := command.Run()
		if err != nil {
			pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not run the editor's command: %w", err)))
			return
		}

		_, err = tempFile.Seek(0, 0)
		if err != nil {
			pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not seek to the beginning of the temporary file: %w", err)))
			return
		}
		data, err = io.ReadAll(tempFile)
		if err != nil {
			pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not read from the temporary file: %w", err)))
			return
		}

		contents = nil
		pTermInput := pterm.DefaultInteractiveTextInput
		pTermInput.Delimiter = ""
		err = json.Unmarshal(data, &contents)
		if err != nil {
			pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not decode the updated vault (expecting JSON): %w", err)))
			_, _ = pTermInput.Show("Press enter to update the vault again...")
			continue
		}
		err = validateObject(contents)
		if err != nil {
			pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("the updated vault has invalid content: %w", err)))
			_, _ = pTermInput.Show("Press enter to update the vault again...")
			continue
		}
		break
	}

	data, _ = json.Marshal(contents)
	err = gpg.EncryptFromMemoryIntoFile(data)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not encrypt the vault content: %w", err)))
		return
	}
	pterm.Info.Println("Vault was updated.")
}

func validateObject(object map[string]any) error {
	if len(object) == 0 {
		return fmt.Errorf("the object has no keys")
	}
	for key, value := range object {
		innerObject, isObject := value.(map[string]any)
		if isObject {
			err := validateObject(innerObject)
			if err != nil {
				return fmt.Errorf("the key \"%s\" has an invalid value object: %w", key, err)
			}
			continue
		}
		_, isString := value.(string)
		if !isString {
			return fmt.Errorf("the key \"%s\" has a value that is neither an object nor a string", key)
		}
	}
	return nil
}

var updatePasswordCommand = &cobra.Command{
	Use:  "password [path]",
	Args: cobra.ExactArgs(1),
	Run:  updatePassword,
}

func updatePassword(_ *cobra.Command, arguments []string) {
	vaultFilePath, err := filepath.Abs(arguments[0])
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not find the absolute path for the vault file: %w", err)))
		return
	}
	oldPassword, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your old password")
	gpg := encryption.GPG{
		FilePath: vaultFilePath,
		Password: oldPassword,
	}
	data, err := gpg.DecryptFromFileIntoMemory()
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not decrypt the vault file: %w", err)))
		return
	}
	newPassword, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your new password")
	confirmation, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your new password again")
	if newPassword != confirmation {
		pterm.Error.Println("The passwords don't match. Take your time and make sure no accidental typos are made. You could permanently lose access to your vault.")
		return
	}
	gpg.Password = newPassword
	err = gpg.EncryptFromMemoryIntoFile(data)
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not encrypt the vault content: %w", err)))
		return
	}
	pterm.Info.Println("The vault password was updated.")
}

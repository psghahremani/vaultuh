package main

import (
	"fmt"
	"github.com/psghahremani/vaultuh/command"
	"github.com/psghahremani/vaultuh/encryption"
	"github.com/psghahremani/vaultuh/locking"
	"github.com/psghahremani/vaultuh/utility"
	"github.com/pterm/pterm"
)

func main() {
	err := encryption.CheckForGPG()
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not check for GPG: %w", err)))
		return
	}

	processLock := locking.ProcessLock{}
	isSuccessful, err := processLock.Hold()
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not hold a process lock: %w", err)))
		return
	}
	if !isSuccessful {
		pterm.Error.Println("Another instance of Vaultuh is running.")
		return
	}

	err = command.Execute()
	if err != nil {
		pterm.Error.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not execute the command: %w", err)))
		return
	}

	err = processLock.Release()
	if err != nil {
		pterm.Warning.Print(utility.GetFormattedErrorMessage(fmt.Errorf("could not release the process lock: %w", err)))
		return
	}
}

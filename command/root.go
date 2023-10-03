package command

import (
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Version: "1.0.0",
	Use:     "vaultuh",
}

func Execute() error {
	return rootCommand.Execute()
}

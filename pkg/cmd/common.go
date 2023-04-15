package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// ExpandArgs expands environment variables in a slice of arguments for a cmd.
func ExpandArgs(cmd *cobra.Command, args []string) error {
	for i, arg := range args {
		args[i] = os.ExpandEnv(arg)
	}
	return cmd.Flags().Parse(args)
}

package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func ExpandArgs(cmd *cobra.Command, args []string) error {
	for i := 0; i < len(args); i++ {
		args[i] = os.ExpandEnv(args[i])
	}
	err := cmd.Flags().Parse(args)
	if err != nil {
		return err
	}
	return nil
}

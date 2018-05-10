package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

type LoadType int

const (
	ConfigMap LoadType = iota
	Secret
	Folder
)

// ExpandArgs expands environment variables in a slice of arguments for a cmd
func ExpandArgs(cmd *cobra.Command, args []string) error {
	for i, arg := range args {
		args[i] = os.ExpandEnv(arg)
	}
	return cmd.Flags().Parse(args)
}

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Version variable is set at build time
var Version = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

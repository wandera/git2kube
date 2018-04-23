package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "git2kube",
	DisableAutoGenTag: true,
	Short:             "Git to ConfigMap conversion tool",
	Long:              `Commandline tool for loading files from git repository into K8s ConfigMap`,
}

func init() {
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(genDocCmd)
}

//Execute run root command (main entrypoint)
func Execute() error {
	return rootCmd.Execute()
}

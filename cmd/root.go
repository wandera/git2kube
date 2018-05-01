package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"fmt"
)

var loglevel string

var rootCmd = &cobra.Command{
	Use:               "git2kube",
	DisableAutoGenTag: true,
	Short:             "Git to ConfigMap conversion tool",
	Long:              `Commandline tool for loading files from git repository into K8s ConfigMap`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		lvl, err := log.ParseLevel(loglevel)
		if err != nil {
			return err
		}

		log.SetLevel(lvl)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&loglevel, "log-level", "l", "info", fmt.Sprintf("command log level (options: %s)", log.AllLevels))

	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(genDocCmd)
}

//Execute run root command (main entrypoint)
func Execute() error {
	return rootCmd.Execute()
}

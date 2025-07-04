package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	loglevel  string
	logformat string
)

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

		switch logformat {
		case "json":
			log.SetFormatter(&log.JSONFormatter{})
		case "logfmt":
			fallthrough
		default:
			log.SetFormatter(&log.TextFormatter{})
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&loglevel, "log-level", "l", "info", fmt.Sprintf("command log level (options: %s)", log.AllLevels))
	rootCmd.PersistentFlags().StringVar(&logformat, "log-format", "logfmt", "log output format (options: logfmt, json)")

	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(genDocCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute run root command (main entrypoint).
func Execute() error {
	return rootCmd.Execute()
}

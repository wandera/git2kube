package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docDestination string

var genDocCmd = &cobra.Command{
	Use:   "gendoc",
	Short: "Generates documentation for this tool in Markdown format",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeGenDoc(args)
	},
}

func executeGenDoc(args []string) error {
	err := doc.GenMarkdownTree(rootCmd, docDestination)
	return err
}

func init() {
	genDocCmd.Flags().StringVarP(&docDestination, "destination", "d", "", "destination for documentation")
	genDocCmd.MarkFlagRequired("destination")
}

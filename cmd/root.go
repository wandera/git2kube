package cmd

import (
	"github.com/spf13/cobra"
)

var rootParams = struct {
	kubeconfig string
	git        string
	branch     string
	folder     string
	mapname    string
	namespace  string
	verbose    bool
}{}

var rootCmd = &cobra.Command{
	Use:               "git2kube",
	DisableAutoGenTag: true,
	Short:             "Git to ConfigMap conversion tool",
	Long:              `Commandline tool for loading files from git repository into K8s ConfigMap`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootParams.verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&rootParams.kubeconfig, "kubeconfig", "k", "", "path to kubeconfig file InCluster config will be used if not specified")
	rootCmd.PersistentFlags().StringVarP(&rootParams.git, "git", "g", "", "git repository address, either http(s) or ssh protocol has to be specified")
	rootCmd.PersistentFlags().StringVarP(&rootParams.branch, "branch", "b", "master", "branch name to pull")
	rootCmd.PersistentFlags().StringVarP(&rootParams.folder, "cache-folder", "c", "/tmp/git2kube/data/", "destination on filesystem where cache of repository will be stored")
	rootCmd.PersistentFlags().StringVarP(&rootParams.namespace, "namespace", "n", "default", "target namespace for resulting ConfigMap")
	rootCmd.PersistentFlags().StringVarP(&rootParams.mapname, "configmap", "m", "", "target namespace for resulting ConfigMap")

	rootCmd.MarkFlagFilename("kubeconfig")
	rootCmd.MarkPersistentFlagRequired("git")
	rootCmd.MarkPersistentFlagRequired("configmap")

	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(genDocCmd)
}

//Execute run root command (main entrypoint)
func Execute() error {
	return rootCmd.Execute()
}

package cmd

import (
	"github.com/WanderaOrg/git2kube/pkg/fetch"
	"github.com/WanderaOrg/git2kube/pkg/upload"
	"github.com/spf13/cobra"
	"os"
)

var lp = struct {
	kubeconfig bool
	git        string
	branch     string
	folder     string
	mapname    string
	namespace  string
	mergetype  string
	verbose    bool
}{}

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Loads files from git repository into ConfigMap",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeLoad(args)
	},
}

func executeLoad(args []string) error {
	if err := os.MkdirAll(lp.folder, 755); err != nil {
		return err
	}

	auth, err := fetch.NewAuth(lp.git)
	if err != nil {
		return err
	}

	fetcher := fetch.NewFetcher(lp.git, lp.folder, lp.branch, auth)

	uploader, err := upload.NewUploader(lp.kubeconfig, lp.mapname, lp.namespace, upload.MergeType(lp.mergetype))
	if err != nil {
		return err
	}

	c, err := fetcher.Fetch()
	if err != nil {
		return err
	}

	iter, err := c.Files()
	if err != nil {
		return err
	}

	err = uploader.Upload(c.ID().String(), iter)
	if err != nil {
		return err
	}

	return err
}

func init() {
	loadCmd.Flags().BoolVarP(&lp.verbose, "verbose", "v", false, "verbose output")
	loadCmd.Flags().BoolVarP(&lp.kubeconfig, "kubeconfig", "k", false, "if locally stored ~/.kube/config should be used, InCluster config will be used if false")
	loadCmd.Flags().StringVarP(&lp.mergetype, "merge-type", "", "delete", "how to merge ConfigMap data whether to also delete missing values or just upsert new (options: delete|upsert)")
	loadCmd.Flags().StringVarP(&lp.git, "git", "g", "", "git repository address, either http(s) or ssh protocol has to be specified")
	loadCmd.Flags().StringVarP(&lp.branch, "branch", "b", "master", "branch name to pull")
	loadCmd.Flags().StringVarP(&lp.folder, "cache-folder", "c", "/tmp/git2kube/data/", "destination on filesystem where cache of repository will be stored")
	loadCmd.Flags().StringVarP(&lp.namespace, "namespace", "n", "default", "target namespace for resulting ConfigMap")
	loadCmd.Flags().StringVarP(&lp.mapname, "configmap", "m", "", "target namespace for resulting ConfigMap")

	loadCmd.MarkFlagFilename("kubeconfig")
	loadCmd.MarkFlagRequired("git")
	loadCmd.MarkFlagRequired("configmap")
}

package cmd

import (
	"github.com/WanderaOrg/git2kube/pkg/fetch"
	"github.com/WanderaOrg/git2kube/pkg/upload"
	"github.com/spf13/cobra"
	"os"
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Loads files from git repository into ConfigMap",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeLoad(args)
	},
}

func executeLoad(args []string) error {
	if err := os.MkdirAll(rootParams.folder, 755); err != nil {
		return err
	}

	auth, err := fetch.NewAuth(rootParams.git)
	if err != nil {
		return err
	}

	fetcher := fetch.NewFetcher(rootParams.git, rootParams.folder, rootParams.branch, auth)

	uploader, err := upload.NewUploader(rootParams.kubeconfig, rootParams.mapname, rootParams.namespace)
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

	err = uploader.Upload(iter)
	if err != nil {
		return err
	}

	return err
}

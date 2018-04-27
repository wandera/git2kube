package cmd

import (
	"github.com/WanderaOrg/git2kube/pkg/cmd"
	"github.com/WanderaOrg/git2kube/pkg/fetch"
	"github.com/WanderaOrg/git2kube/pkg/upload"
	"github.com/spf13/cobra"
	"os"
)

var lp = struct {
	kubeconfig  bool
	git         string
	branch      string
	folder      string
	mapname     string
	namespace   string
	mergetype   string
	verbose     bool
	includes    []string
	excludes    []string
	labels      []string
	annotations []string
}{}

var loadCmd = &cobra.Command{
	Use:     "load",
	Short:   "Loads files from git repository into ConfigMap",
	PreRunE: cmd.ExpandArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeLoad()
	},
}

func executeLoad() error {
	if err := os.MkdirAll(lp.folder, 755); err != nil {
		return err
	}

	auth, err := fetch.NewAuth(lp.git)
	if err != nil {
		return err
	}

	fetcher := fetch.NewFetcher(lp.git, lp.folder, lp.branch, auth)

	uploader, err := upload.NewUploader(&upload.UploaderOptions{
		Kubeconfig:    lp.kubeconfig,
		ConfigMapName: lp.mapname,
		Namespace:     lp.namespace,
		MergeType:     upload.MergeType(lp.mergetype),
		Includes:      lp.includes,
		Excludes:      lp.excludes,
		Annotations:   lp.annotations,
		Labels:        lp.labels,
	})
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
	loadCmd.Flags().BoolVarP(&lp.kubeconfig, "kubeconfig", "k", false, "true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)")
	loadCmd.Flags().StringVarP(&lp.mergetype, "merge-type", "", "delete", "how to merge ConfigMap data whether to also delete missing values or just upsert new (options: delete|upsert)")
	loadCmd.Flags().StringVarP(&lp.git, "git", "g", "", "git repository address, either http(s) or ssh protocol has to be specified")
	loadCmd.Flags().StringVarP(&lp.branch, "branch", "b", "master", "branch name to pull")
	loadCmd.Flags().StringVarP(&lp.folder, "cache-folder", "c", "/tmp/git2kube/data/", "destination on filesystem where cache of repository will be stored")
	loadCmd.Flags().StringVarP(&lp.namespace, "namespace", "n", "default", "target namespace for the resulting ConfigMap")
	loadCmd.Flags().StringVarP(&lp.mapname, "configmap", "m", "", "name for the resulting ConfigMap")
	loadCmd.Flags().StringSliceVar(&lp.includes, "include", []string{".*"}, "regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	loadCmd.Flags().StringSliceVar(&lp.excludes, "exclude", []string{"^\\..*"}, "regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	loadCmd.Flags().StringSliceVar(&lp.labels, "label", []string{}, "label to add to K8s ConfigMap (format NAME=VALUE)")
	loadCmd.Flags().StringSliceVar(&lp.annotations, "annotation", []string{}, "annotation to add to K8s ConfigMap (format NAME=VALUE)")

	loadCmd.MarkFlagFilename("kubeconfig")
	loadCmd.MarkFlagRequired("git")
	loadCmd.MarkFlagRequired("configmap")
}

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
	target      string
	namespace   string
	mergetype   string
	includes    []string
	excludes    []string
	sshkey      string
	labels      []string
	annotations []string
}{}

var loadCmd = &cobra.Command{
	Use:                "load",
	Short:              "Loads files from git repository into target",
	DisableFlagParsing: true,
	PersistentPreRunE:  cmd.ExpandArgs,
}

var loadConfigmapCmd = &cobra.Command{
	Use:                "configmap",
	Short:              "Loads files from git repository into ConfigMap",
	DisableFlagParsing: true,
	RunE: func(c *cobra.Command, args []string) error {
		return executeLoad(cmd.ConfigMap)
	},
}

var loadSecretCmd = &cobra.Command{
	Use:                "secret",
	Short:              "Loads files from git repository into Secret",
	DisableFlagParsing: true,
	RunE: func(c *cobra.Command, args []string) error {
		return executeLoad(cmd.Secret)
	},
}

var loadFolderCmd = &cobra.Command{
	Use:                "folder",
	Short:              "Loads files from git repository into Folder",
	DisableFlagParsing: true,
	RunE: func(c *cobra.Command, args []string) error {
		return executeLoad(cmd.Folder)
	},
}

func executeLoad(lt cmd.LoadType) error {
	if err := os.MkdirAll(lp.folder, os.ModePerm); err != nil {
		return err
	}

	auth, err := fetch.NewAuth(lp.git, lp.sshkey)
	if err != nil {
		return err
	}

	fetcher := fetch.NewFetcher(lp.git, lp.folder, lp.branch, auth)

	c, err := fetcher.Fetch()
	if err != nil {
		return err
	}

	iter, err := c.Files()
	if err != nil {
		return err
	}

	var up upload.Uploader
	switch lt {
	case cmd.ConfigMap:
		uploader, err := upload.NewConfigMapUploader(&upload.UploaderOptions{
			Kubeconfig:  lp.kubeconfig,
			Target:      lp.target,
			Namespace:   lp.namespace,
			MergeType:   upload.MergeType(lp.mergetype),
			Includes:    lp.includes,
			Excludes:    lp.excludes,
			Annotations: lp.annotations,
			Labels:      lp.labels,
		})
		if err != nil {
			return err
		}
		up = uploader
	case cmd.Secret:
		uploader, err := upload.NewSecretUploader(&upload.UploaderOptions{
			Kubeconfig:  lp.kubeconfig,
			Target:      lp.target,
			Namespace:   lp.namespace,
			MergeType:   upload.MergeType(lp.mergetype),
			Includes:    lp.includes,
			Excludes:    lp.excludes,
			Annotations: lp.annotations,
			Labels:      lp.labels,
		})
		if err != nil {
			return err
		}
		up = uploader
	case cmd.Folder:
		uploader, err := upload.NewFolderUploader(&upload.UploaderOptions{
			Source:   lp.folder,
			Target:   lp.target,
			Includes: lp.includes,
			Excludes: lp.excludes,
		})
		if err != nil {
			return err
		}
		up = uploader
	}

	err = up.Upload(c.ID().String(), iter)
	if err != nil {
		return err
	}

	return err
}

func init() {
	loadCmd.PersistentFlags().StringVarP(&lp.git, "git", "g", "", "git repository address, either http(s) or ssh protocol has to be specified")
	loadCmd.PersistentFlags().StringVarP(&lp.branch, "branch", "b", "master", "branch name to pull")
	loadCmd.PersistentFlags().StringVarP(&lp.folder, "cache-folder", "c", "/tmp/git2kube/data/", "destination on filesystem where cache of repository will be stored")
	loadCmd.PersistentFlags().StringSliceVar(&lp.includes, "include", []string{".*"}, "regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	loadCmd.PersistentFlags().StringSliceVar(&lp.excludes, "exclude", []string{"^\\..*"}, "regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	loadCmd.PersistentFlags().StringVarP(&lp.sshkey, "ssh-key", "p", "", "path to the SSH private key (git repository address should be 'git@<address>', example: git@github.com:WanderaOrg/git2kube.git)")

	loadCmd.MarkPersistentFlagRequired("git")
	loadCmd.MarkPersistentFlagFilename("cache-folder")

	loadConfigmapCmd.Flags().BoolVarP(&lp.kubeconfig, "kubeconfig", "k", false, "true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)")
	loadConfigmapCmd.Flags().StringVarP(&lp.namespace, "namespace", "n", "default", "target namespace for the resulting ConfigMap")
	loadConfigmapCmd.Flags().StringVarP(&lp.target, "configmap", "m", "", "name for the resulting ConfigMap")
	loadConfigmapCmd.Flags().StringSliceVar(&lp.labels, "label", []string{}, "label to add to K8s ConfigMap (format NAME=VALUE)")
	loadConfigmapCmd.Flags().StringSliceVar(&lp.annotations, "annotation", []string{}, "annotation to add to K8s ConfigMap (format NAME=VALUE)")
	loadConfigmapCmd.Flags().StringVarP(&lp.mergetype, "merge-type", "", "delete", "how to merge ConfigMap data whether to also delete missing values or just upsert new (options: delete|upsert)")
	loadConfigmapCmd.MarkFlagFilename("kubeconfig")
	loadConfigmapCmd.MarkFlagRequired("configmap")

	loadSecretCmd.Flags().BoolVarP(&lp.kubeconfig, "kubeconfig", "k", false, "true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)")
	loadSecretCmd.Flags().StringVarP(&lp.namespace, "namespace", "n", "default", "target namespace for the resulting ConfigMap")
	loadSecretCmd.Flags().StringVarP(&lp.target, "secret", "s", "", "name for the resulting Secret")
	loadSecretCmd.Flags().StringSliceVar(&lp.labels, "label", []string{}, "label to add to K8s Secret (format NAME=VALUE)")
	loadSecretCmd.Flags().StringSliceVar(&lp.annotations, "annotation", []string{}, "annotation to add to K8s Secret (format NAME=VALUE)")
	loadSecretCmd.Flags().StringVarP(&lp.mergetype, "merge-type", "", "delete", "how to merge Secret data whether to also delete missing values or just upsert new (options: delete|upsert)")
	loadSecretCmd.MarkFlagFilename("kubeconfig")
	loadSecretCmd.MarkFlagRequired("secret")

	loadFolderCmd.Flags().StringVarP(&lp.target, "target-folder", "t", "", "path to target folder")
	loadFolderCmd.MarkFlagRequired("target-folder")
	loadFolderCmd.MarkFlagFilename("target-folder")

	loadCmd.AddCommand(loadConfigmapCmd)
	loadCmd.AddCommand(loadSecretCmd)
	loadCmd.AddCommand(loadFolderCmd)
}

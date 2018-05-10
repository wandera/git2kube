package cmd

import (
	"github.com/WanderaOrg/git2kube/pkg/cmd"
	"github.com/WanderaOrg/git2kube/pkg/fetch"
	"github.com/WanderaOrg/git2kube/pkg/upload"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var wp = struct {
	kubeconfig  bool
	git         string
	branch      string
	folder      string
	target      string
	namespace   string
	mergetype   string
	interval    int
	includes    []string
	excludes    []string
	labels      []string
	annotations []string
}{}

var watchCmd = &cobra.Command{
	Use:                "watch",
	Short:              "Runs watcher that periodically check the provided repository",
	DisableFlagParsing: true,
	PersistentPreRunE:  cmd.ExpandArgs,
}

var watchConfigmapCmd = &cobra.Command{
	Use:   "configmap",
	Short: "Runs watcher that periodically check the provided repository and updates K8s ConfigMap accordingly",
	DisableFlagParsing: true,
	RunE: func(c *cobra.Command, args []string) error {
		return executeWatch(cmd.ConfigMap)
	},
}

var watchSecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Runs watcher that periodically check the provided repository and updates K8s Secret accordingly",
	DisableFlagParsing: true,
	RunE: func(c *cobra.Command, args []string) error {
		return executeWatch(cmd.Secret)
	},
}

var watchFolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "Runs watcher that periodically check the provided repository and updates target folder accordingly",
	DisableFlagParsing: true,
	RunE: func(c *cobra.Command, args []string) error {
		return executeWatch(cmd.Folder)
	},
}

func executeWatch(lt cmd.LoadType) error {
	if err := os.MkdirAll(wp.folder, os.ModePerm); err != nil {
		return err
	}

	auth, err := fetch.NewAuth(wp.git)
	if err != nil {
		return err
	}

	fetcher := fetch.NewFetcher(wp.git, wp.folder, wp.branch, auth)

	var up upload.Uploader
	switch lt {
	case cmd.ConfigMap:
		uploader, err := upload.NewConfigMapUploader(&upload.UploaderOptions{
			Kubeconfig:  wp.kubeconfig,
			Target:      wp.target,
			Namespace:   wp.namespace,
			MergeType:   upload.MergeType(wp.mergetype),
			Includes:    wp.includes,
			Excludes:    wp.excludes,
			Annotations: wp.annotations,
			Labels:      wp.labels,
		})
		if err != nil {
			return err
		}
		up = uploader
	case cmd.Secret:
		uploader, err := upload.NewSecretUploader(&upload.UploaderOptions{
			Kubeconfig:  wp.kubeconfig,
			Target:      wp.target,
			Namespace:   wp.namespace,
			MergeType:   upload.MergeType(wp.mergetype),
			Includes:    wp.includes,
			Excludes:    wp.excludes,
			Annotations: wp.annotations,
			Labels:      wp.labels,
		})
		if err != nil {
			return err
		}
		up = uploader
	case cmd.Folder:
		uploader, err := upload.NewFolderUploader(&upload.UploaderOptions{
			Source:   wp.folder,
			Target:   wp.target,
			Includes: wp.includes,
			Excludes: wp.excludes,
		})
		if err != nil {
			return err
		}
		up = uploader
	}

	ticker := time.NewTicker(time.Duration(wp.interval) * time.Second)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				err := refresh(fetcher, up)
				if err != nil {
					log.Warn(err)
				}
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Started watcher")
	<-signalChan
	log.Info("Shutdown signal received, exiting...")
	close(stop)

	return nil
}

func refresh(fetcher fetch.Fetcher, uploader upload.Uploader) error {
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
	watchCmd.PersistentFlags().IntVarP(&wp.interval, "interval", "i", 10, "interval in seconds in which to try refreshing ConfigMap from git")
	watchCmd.PersistentFlags().StringVarP(&wp.git, "git", "g", "", "git repository address, either http(s) or ssh protocol has to be specified")
	watchCmd.PersistentFlags().StringVarP(&wp.branch, "branch", "b", "master", "branch name to pull")
	watchCmd.PersistentFlags().StringVarP(&wp.folder, "cache-folder", "c", "/tmp/git2kube/data/", "destination on filesystem where cache of repository will be stored")
	watchCmd.PersistentFlags().StringSliceVar(&wp.includes, "include", []string{".*"}, "regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	watchCmd.PersistentFlags().StringSliceVar(&wp.excludes, "exclude", []string{"^\\..*"}, "regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	watchCmd.MarkPersistentFlagRequired("git")

	watchConfigmapCmd.Flags().BoolVarP(&wp.kubeconfig, "kubeconfig", "k", false, "true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)")
	watchConfigmapCmd.Flags().StringVarP(&wp.namespace, "namespace", "n", "default", "target namespace for the resulting ConfigMap")
	watchConfigmapCmd.Flags().StringVarP(&wp.target, "configmap", "m", "", "name for the resulting ConfigMap")
	watchConfigmapCmd.Flags().StringSliceVar(&wp.labels, "label", []string{}, "label to add to K8s ConfigMap (format NAME=VALUE)")
	watchConfigmapCmd.Flags().StringSliceVar(&wp.annotations, "annotation", []string{}, "annotation to add to K8s ConfigMap (format NAME=VALUE)")
	watchConfigmapCmd.Flags().StringVarP(&wp.mergetype, "merge-type", "", "delete", "how to merge ConfigMap data whether to also delete missing values or just upsert new (options: delete|upsert)")
	watchConfigmapCmd.MarkFlagFilename("kubeconfig")
	watchConfigmapCmd.MarkFlagRequired("configmap")

	watchSecretCmd.Flags().BoolVarP(&wp.kubeconfig, "kubeconfig", "k", false, "true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)")
	watchSecretCmd.Flags().StringVarP(&wp.namespace, "namespace", "n", "default", "target namespace for the resulting ConfigMap")
	watchSecretCmd.Flags().StringVarP(&wp.target, "secret", "s", "", "name for the resulting Secret")
	watchSecretCmd.Flags().StringSliceVar(&wp.labels, "label", []string{}, "label to add to K8s Secret (format NAME=VALUE)")
	watchSecretCmd.Flags().StringSliceVar(&wp.annotations, "annotation", []string{}, "annotation to add to K8s Secret (format NAME=VALUE)")
	watchSecretCmd.Flags().StringVarP(&wp.mergetype, "merge-type", "", "delete", "how to merge Secret data whether to also delete missing values or just upsert new (options: delete|upsert)")
	watchSecretCmd.MarkFlagFilename("kubeconfig")
	watchSecretCmd.MarkFlagRequired("secret")

	watchFolderCmd.Flags().StringVarP(&wp.target, "target-folder", "t", "", "path to target folder")
	watchFolderCmd.MarkFlagRequired("target-folder")

	watchCmd.AddCommand(watchConfigmapCmd)
	watchCmd.AddCommand(watchSecretCmd)
	watchCmd.AddCommand(watchFolderCmd)
}

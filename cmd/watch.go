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
	mapname     string
	namespace   string
	mergetype   string
	verbose     bool
	interval    int
	includes    []string
	excludes    []string
	labels      []string
	annotations []string
}{}

var watchCmd = &cobra.Command{
	Use:                "watch",
	Short:              "Runs watcher that periodically check the provided repository and updates K8s configmap accordingly",
	DisableFlagParsing: true,
	PreRunE:            cmd.ExpandArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeWatch()
	},
}

func executeWatch() error {
	if err := os.MkdirAll(wp.folder, 755); err != nil {
		return err
	}

	auth, err := fetch.NewAuth(wp.git)
	if err != nil {
		return err
	}

	fetcher := fetch.NewFetcher(wp.git, wp.folder, wp.branch, auth)

	uploader, err := upload.NewConfigMapUploader(&upload.UploaderOptions{
		Kubeconfig:  wp.kubeconfig,
		Target:      wp.mapname,
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

	ticker := time.NewTicker(time.Duration(wp.interval) * time.Second)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				err := refresh(fetcher, uploader)
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
	watchCmd.Flags().IntVarP(&wp.interval, "interval", "i", 10, "interval in seconds in which to try refreshing ConfigMap from git")
	watchCmd.Flags().StringVarP(&wp.mergetype, "merge-type", "", "delete", "how to merge ConfigMap data whether to also delete missing values or just upsert new (options: delete|upsert)")
	watchCmd.Flags().BoolVarP(&wp.verbose, "verbose", "v", false, "verbose output")
	watchCmd.Flags().BoolVarP(&wp.kubeconfig, "kubeconfig", "k", false, "true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)")
	watchCmd.Flags().StringVarP(&wp.git, "git", "g", "", "git repository address, either http(s) or ssh protocol has to be specified")
	watchCmd.Flags().StringVarP(&wp.branch, "branch", "b", "master", "branch name to pull")
	watchCmd.Flags().StringVarP(&wp.folder, "cache-folder", "c", "/tmp/git2kube/data/", "destination on filesystem where cache of repository will be stored")
	watchCmd.Flags().StringVarP(&wp.namespace, "namespace", "n", "default", "target namespace for the resulting ConfigMap")
	watchCmd.Flags().StringVarP(&wp.mapname, "configmap", "m", "", "name for the resulting ConfigMap")
	watchCmd.Flags().StringSliceVar(&wp.includes, "include", []string{".*"}, "regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	watchCmd.Flags().StringSliceVar(&wp.excludes, "exclude", []string{"^\\..*"}, "regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder")
	watchCmd.Flags().StringSliceVar(&wp.labels, "label", []string{}, "label to add to K8s ConfigMap (format NAME=VALUE)")
	watchCmd.Flags().StringSliceVar(&wp.annotations, "annotation", []string{}, "annotation to add to K8s ConfigMap (format NAME=VALUE)")

	watchCmd.MarkFlagFilename("kubeconfig")
	watchCmd.MarkFlagRequired("git")
	watchCmd.MarkFlagRequired("configmap")
}

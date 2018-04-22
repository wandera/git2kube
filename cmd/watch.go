package cmd

import (
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
	interval int
}{}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Decrypt the value server-side and prints the response",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeWatch(args)
	},
}

func executeWatch(strings []string) error {
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

	ticker := time.NewTicker(time.Duration(wp.interval) * time.Second)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				err := refresh(fetcher, uploader)
				if err != nil {
					log.Println(err)
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

	err = uploader.Upload(iter)
	if err != nil {
		return err
	}

	return err
}

func init() {
	watchCmd.Flags().IntVarP(&wp.interval, "interval", "i", 10, "interval in seconds in which to try refreshing ConfigMap from git")
}

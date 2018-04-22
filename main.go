package main

import (
	"github.com/WanderaOrg/git2kube/cmd"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {

	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}

}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

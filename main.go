package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/wandera/git2kube/cmd"
	"os"
)

func main() {

	if err := cmd.Execute(); err != nil {
		log.Errorf("Command failed: %v", err)
		os.Exit(-1)
	}

}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

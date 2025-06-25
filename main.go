package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wandera/git2kube/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Errorf("Command failed: %v", err)
		os.Exit(-1)
	}
}

func init() {
	log.SetOutput(os.Stdout)
}

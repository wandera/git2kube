package main

import (
	"github.com/WanderaOrg/git2kube/cmd"
	"os"
)

func main() {

	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}

}

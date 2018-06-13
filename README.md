# git2kube - Git to K8s ConfigMap
           
[![Build Status](https://travis-ci.org/WanderaOrg/git2kube.svg?branch=master)](https://travis-ci.org/WanderaOrg/git2kube)
[![Docker Build Status](https://img.shields.io/docker/build/wanderadock/git2kube.svg)](https://hub.docker.com/r/wanderadock/git2kube/)
[![Go Report Card](https://goreportcard.com/badge/github.com/WanderaOrg/git2kube)](https://goreportcard.com/report/github.com/WanderaOrg/git2kube)
[![GitHub release](https://img.shields.io/github/release/WanderaOrg/git2kube.svg)](https://github.com/WanderaOrg/git2kube/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/WanderaOrg/scccmd/blob/master/LICENSE)

Tool for syncing git with Kubernetes.

### Features
* Synchronisation of Git repository with Kubernetes ConfigMap/Secret
  * One shot or periodic
  * Configurable healthcheck
  * Configurable labels and annotations
* Configurable include/exclude rules for filtering files that should be synchronised
* Ability to synchronise git into target folder using symlinks (suitable for sidecar deployments)
* SSH key and Basic auth

### Docker repository
The tool is released as docker image as well, check the [repository](https://hub.docker.com/r/wanderadock/git2kube/).

### Documentation
* [docs](docs/git2kube.md) - Generated documentation for the tool
* [example](example) - Kubernetes deployment examples

### How to develop
* Checkout into your GOROOT directory (e.g. /go/src/github.com/WanderaOrg/git2kube)
* `cd` into the folder and run `dep ensure --vendor-only`
* Tests are started by `go test -v ./...`
* Or if you dont want to setup your local go env just use the provided Dockerfile
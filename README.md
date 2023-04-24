# git2kube - From Git to Kubernetes

[![Test](https://github.com/wandera/git2kube/actions/workflows/test.yml/badge.svg)](https://github.com/wandera/git2kube/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wandera/git2kube)](https://goreportcard.com/report/github.com/wandera/git2kube)
[![GitHub release](https://img.shields.io/github/release/wandera/git2kube.svg)](https://github.com/wandera/git2kube/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/wandera/scccmd/blob/master/LICENSE)

Tool for syncing git with Kubernetes.

### Features
* Synchronisation of Git repository with Kubernetes ConfigMap/Secret
  * One shot or periodic
  * Configurable healthcheck
  * Configurable labels and annotations
* Configurable include/exclude rules for filtering files that should be synchronised
* Ability to synchronise git into target folder using symlinks (suitable for sidecar deployments)
* SSH key and Basic auth

### Quickstart
Check out [example](example) folder that should get you started. 

### Docker repository
The tool is released as docker image as well, check the [repository](https://github.com/wandera/git2kube/pkgs/container/git2kube).

### Documentation
* [docs](docs/git2kube.md) - Generated documentation for the tool
* [example](example) - Kubernetes deployment examples

### How to develop
* Tests are started by `go test -v ./...`
* Or if you dont want to setup your local go env just use the provided Dockerfile

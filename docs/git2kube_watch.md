## git2kube watch

Runs watcher that periodically check the provided repository

### Synopsis

Runs watcher that periodically check the provided repository

### Options

```
  -b, --branch string             branch name to pull (default "master")
  -c, --cache-folder string       destination on filesystem where cache of repository will be stored (default "/tmp/git2kube/data/")
      --exclude strings           regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder (default [^\..*])
  -g, --git string                git repository address, either http(s) or ssh protocol has to be specified
      --healthcheck-file string   path to file where each refresh writes if it was successful or not, useful for K8s liveness/readiness probe
  -h, --help                      help for watch
      --include strings           regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder (default [.*])
  -i, --interval int              interval in seconds in which to try refreshing ConfigMap from git (default 10)
  -p, --ssh-key string            path to the SSH private key (git repository address should be 'git@<address>', example: git@github.com:wandera/git2kube.git)
```

### Options inherited from parent commands

```
  -l, --log-level string   command log level (options: [panic fatal error warning info debug trace]) (default "info")
```

### SEE ALSO

* [git2kube](git2kube.md)	 - Git to ConfigMap conversion tool
* [git2kube watch configmap](git2kube_watch_configmap.md)	 - Runs watcher that periodically check the provided repository and updates K8s ConfigMap accordingly
* [git2kube watch folder](git2kube_watch_folder.md)	 - Runs watcher that periodically check the provided repository and updates target folder accordingly
* [git2kube watch secret](git2kube_watch_secret.md)	 - Runs watcher that periodically check the provided repository and updates K8s Secret accordingly


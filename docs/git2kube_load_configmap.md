## git2kube load configmap

Loads files from git repository into ConfigMap

### Synopsis

Loads files from git repository into ConfigMap

```
git2kube load configmap [flags]
```

### Options

```
      --annotation strings   annotation to add to K8s ConfigMap (format NAME=VALUE)
  -m, --configmap string     name for the resulting ConfigMap
  -h, --help                 help for configmap
  -k, --kubeconfig           true if locally stored ~/.kube/config should be used, InCluster config will be used if false (options: true|false) (default: false)
      --label strings        label to add to K8s ConfigMap (format NAME=VALUE)
      --merge-type string    how to merge ConfigMap data whether to also delete missing values or just upsert new (options: delete|upsert) (default "delete")
  -n, --namespace string     target namespace for the resulting ConfigMap (default "default")
```

### Options inherited from parent commands

```
  -b, --branch string         branch name to pull (default "master")
  -c, --cache-folder string   destination on filesystem where cache of repository will be stored (default "/tmp/git2kube/data/")
      --exclude strings       regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder (default [^\..*])
  -g, --git string            git repository address, either http(s) or ssh protocol has to be specified
      --include strings       regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder (default [.*])
  -l, --log-level string      command log level (options: [panic fatal error warning info debug]) (default "info")
  -p, --ssh-key string        path to the SSH private key (git repository address should be 'git@<address>', example: git@github.com:wandera/git2kube.git)
```

### SEE ALSO

* [git2kube load](git2kube_load.md)	 - Loads files from git repository into target


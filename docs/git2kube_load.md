## git2kube load

Loads files from git repository into target

### Synopsis

Loads files from git repository into target

### Options

```
  -b, --branch string         branch name to pull (default "master")
  -c, --cache-folder string   destination on filesystem where cache of repository will be stored (default "/tmp/git2kube/data/")
      --exclude strings       regex that if is a match excludes the file from the upload, example: '*.yaml' or 'folder/*' if you want to match a folder (default [^\..*])
  -g, --git string            git repository address, either http(s) or ssh protocol has to be specified
  -h, --help                  help for load
      --include strings       regex that if is a match includes the file in the upload, example: '*.yaml' or 'folder/*' if you want to match a folder (default [.*])
  -p, --ssh-key string        path to the SSH private key (git repository address should be 'git@<address>', example: git@github.com:wandera/git2kube.git)
```

### Options inherited from parent commands

```
  -l, --log-level string   command log level (options: [panic fatal error warning info debug trace]) (default "info")
```

### SEE ALSO

* [git2kube](git2kube.md)	 - Git to ConfigMap conversion tool
* [git2kube load configmap](git2kube_load_configmap.md)	 - Loads files from git repository into ConfigMap
* [git2kube load folder](git2kube_load_folder.md)	 - Loads files from git repository into Folder
* [git2kube load secret](git2kube_load_secret.md)	 - Loads files from git repository into Secret


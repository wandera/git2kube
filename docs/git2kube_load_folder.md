## git2kube load folder

Loads files from git repository into Folder

### Synopsis

Loads files from git repository into Folder

```
git2kube load folder [flags]
```

### Options

```
  -h, --help                   help for folder
  -t, --target-folder string   path to target folder
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


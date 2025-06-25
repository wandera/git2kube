## git2kube completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(git2kube completion zsh)

To load completions for every new session, execute once:

#### Linux:

	git2kube completion zsh > "${fpath[1]}/_git2kube"

#### macOS:

	git2kube completion zsh > $(brew --prefix)/share/zsh/site-functions/_git2kube

You will need to start a new shell for this setup to take effect.


```
git2kube completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --log-format string   log output format (options: logfmt, json) (default "logfmt")
  -l, --log-level string    command log level (options: [panic fatal error warning info debug trace]) (default "info")
```

### SEE ALSO

* [git2kube completion](git2kube_completion.md)	 - Generate the autocompletion script for the specified shell


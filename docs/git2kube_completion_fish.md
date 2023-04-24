## git2kube completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	git2kube completion fish | source

To load completions for every new session, execute once:

	git2kube completion fish > ~/.config/fish/completions/git2kube.fish

You will need to start a new shell for this setup to take effect.


```
git2kube completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -l, --log-level string   command log level (options: [panic fatal error warning info debug trace]) (default "info")
```

### SEE ALSO

* [git2kube completion](git2kube_completion.md)	 - Generate the autocompletion script for the specified shell


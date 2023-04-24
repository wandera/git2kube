## git2kube completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	git2kube completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
git2kube completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -l, --log-level string   command log level (options: [panic fatal error warning info debug trace]) (default "info")
```

### SEE ALSO

* [git2kube completion](git2kube_completion.md)	 - Generate the autocompletion script for the specified shell


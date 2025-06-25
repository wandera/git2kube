## git2kube completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(git2kube completion bash)

To load completions for every new session, execute once:

#### Linux:

	git2kube completion bash > /etc/bash_completion.d/git2kube

#### macOS:

	git2kube completion bash > $(brew --prefix)/etc/bash_completion.d/git2kube

You will need to start a new shell for this setup to take effect.


```
git2kube completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --log-format string   log output format (options: logfmt, json) (default "logfmt")
  -l, --log-level string    command log level (options: [panic fatal error warning info debug trace]) (default "info")
```

### SEE ALSO

* [git2kube completion](git2kube_completion.md)	 - Generate the autocompletion script for the specified shell


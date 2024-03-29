## Topwatcher Completion ZSH

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(topwatcher completion zsh)

To load completions for every new session, execute once:

#### Linux:

	topwatcher completion zsh > "${fpath[1]}/_topwatcher"

#### macOS:

	topwatcher completion zsh > $(brew --prefix)/share/zsh/site-functions/_topwatcher

You will need to start a new shell for this setup to take effect.


```
topwatcher completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [topwatcher completion](topwatcher_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 18-Jul-2023

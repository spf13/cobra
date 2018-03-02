## Generating Zsh Completion for your cobra.Command

Cobra supports native Zsh completion generated from the root `cobra.Command`.
The generated completion script should be put somewhere in your `$fpath` named
`_<YOUR COMMAND>`.

### What's Supported

* Completion for all non-hidden subcommands using their `.Short` description.
* Completion for all non-hidden flags using the following rules:
  * Filename completion works by marking the flag with `cmd.MarkFlagFilename...`
    family of commands.
  * The requirement for argument to the flag is decided by the `.NoOptDefVal`
    flag value - if it's empty then completion will expect an argument.
  * Flags of one of the various `*Arrary` and `*Slice` types supports multiple
    specifications (with or without argument depending on the specific type).

### What's not yet Supported

* Positional argument completion are not supported yet.
* Custom completion scripts are not supported yet (We should probably create zsh
  specific one, doesn't make sense to re-use the bash one as the functions will
  be different).
* Whatever other feature you're looking for and doesn't exist :)

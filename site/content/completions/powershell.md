---
weight: 13
---

# PowerShell completions

Cobra can generate PowerShell completion scripts. Users need PowerShell version 5.0 or above, which comes with Windows 10 and can be downloaded separately for Windows 7 or 8.1. They can then write the completions to a file and source this file from their PowerShell profile, which is referenced by the `$Profile` environment variable. See `Get-Help about_Profiles` for more info about PowerShell profiles.

*Note*: PowerShell completions have not (yet?) been aligned to Cobra's generic shell completion support.  This implies the PowerShell completions are not as rich as for other shells (see [What's not yet supported](#whats-not-yet-supported)), and may behave slightly differently.  They are still very useful for PowerShell users.

Cobra supports native PowerShell completions generated from the root `cobra.Command`. You can use the `command.GenPowerShellCompletion()` or `command.GenPowerShellCompletionFile()` functions. To include descriptions use `command.GenPowerShellCompletionWithDesc()` and `command.GenPowerShellCompletionFileWithDesc()`. Cobra will provide the description automatically based on usage information. You can choose to make this option configurable by your users.

The script is designed to support all three PowerShell completion modes:

* TabCompleteNext (default windows style - on each key press the next option is displayed)
* Complete (works like bash)
* MenuComplete (works like zsh)

You set the mode with `Set-PSReadLineKeyHandler -Key Tab -Function <mode>`. Descriptions are only displayed when using the `Complete` or `MenuComplete` mode.

Users need PowerShell version 5.0 or above, which comes with Windows 10 and can be downloaded separately for Windows 7 or 8.1. They can then write the completions to a file and source this file from their PowerShell profile, which is referenced by the `$Profile` environment variable. See `Get-Help about_Profiles` for more info about PowerShell profiles.

```
# With descriptions and Mode 'Complete'
$ helm s[tab]
search  (search for a keyword in charts)  show  (show information of a chart)  status  (displays the status of the named release)

# With descriptions and Mode 'MenuComplete' The description of the current selected value will be displayed below the suggestions.
$ helm s[tab]
search    show     status

search for a keyword in charts

# Without descriptions
$ helm s[tab]
search  show  status
```

## Aliases

You can also configure `powershell` aliases for your program and they will also support completions.

```
$ sal aliasname origcommand
$ Register-ArgumentCompleter -CommandName 'aliasname' -ScriptBlock $__origcommandCompleterBlock

# and now when you run `aliasname` completion will make
# suggestions as it did for `origcommand`.

$ aliasname <tab>
completion     firstcommand   secondcommand
```
The name of the completer block variable is of the form `$__<programName>CompleterBlock` where every `-` and `:` in the program name have been replaced with `_`, to respect powershell naming syntax.

## Limitations

* Custom completions implemented in bash scripting (legacy) are not supported and will be ignored for `powershell` (including the use of the `BashCompCustom` flag annotation).
  * You should instead use `ValidArgsFunction` and `RegisterFlagCompletionFunc()` which are portable to the different shells (`bash`, `zsh`, `fish`, `powershell`).
* The function `MarkFlagCustom()` is not supported and will be ignored for `powershell`.
  * You should instead use `RegisterFlagCompletionFunc()`.
* The following flag completion annotations are not supported and will be ignored for `powershell`:
  * `BashCompFilenameExt` (filtering by file extension)
  * `BashCompSubdirsInDir` (filtering by directory)
* The functions corresponding to the above annotations are consequently not supported and will be ignored for `powershell`:
  * `MarkFlagFilename()` and `MarkPersistentFlagFilename()` (filtering by file extension)
  * `MarkFlagDirname()` and `MarkPersistentFlagDirname()` (filtering by directory)
* Similarly, the following completion directives are not supported and will be ignored for `powershell`:
  * `ShellCompDirectiveFilterFileExt` (filtering by file extension)
  * `ShellCompDirectiveFilterDirs` (filtering by directory)

## What's supported

- Completion for subcommands using their `.Short` description
- Completion for non-hidden flags using their `.Name` and `.Shorthand`

## What's not yet supported

- Command aliases
- Required, filename or custom flags (they will work like normal flags)
- Custom completion scripts

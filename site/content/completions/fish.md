---
weight: 12
---

# Fish completions

Cobra supports native fish completions generated from the root `cobra.Command`.  You can use the `command.GenFishCompletion()` or `command.GenFishCompletionFile()` functions. You must provide these functions with a parameter indicating if the completions should be annotated with a description; Cobra will provide the description automatically based on usage information.  You can choose to make this option configurable by your users.
```
# With descriptions
$ helm s[tab]
search  (search for a keyword in charts)  show  (show information of a chart)  status  (displays the status of the named release)

# Without descriptions
$ helm s[tab]
search  show  status
```
*Note*: Because of backward-compatibility requirements, we were forced to have a different API to disable completion descriptions between `zsh` and `fish`.

## Limitations

* Custom completions implemented in bash scripting (legacy) are not supported and will be ignored for `fish` (including the use of the `BashCompCustom` flag annotation).
  * You should instead use `ValidArgsFunction` and `RegisterFlagCompletionFunc()` which are portable to the different shells (`bash`, `zsh`, `fish`, `powershell`).
* The function `MarkFlagCustom()` is not supported and will be ignored for `fish`.
  * You should instead use `RegisterFlagCompletionFunc()`.
* The following flag completion annotations are not supported and will be ignored for `fish`:
  * `BashCompFilenameExt` (filtering by file extension)
  * `BashCompSubdirsInDir` (filtering by directory)
* The functions corresponding to the above annotations are consequently not supported and will be ignored for `fish`:
  * `MarkFlagFilename()` and `MarkPersistentFlagFilename()` (filtering by file extension)
  * `MarkFlagDirname()` and `MarkPersistentFlagDirname()` (filtering by directory)
* Similarly, the following completion directives are not supported and will be ignored for `fish`:
  * `ShellCompDirectiveFilterFileExt` (filtering by file extension)
  * `ShellCompDirectiveFilterDirs` (filtering by directory)

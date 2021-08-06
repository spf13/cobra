# Active Help

Active Help is a framework provided by Cobra which allows a program to define messages (hints, warnings, etc) that will be printed during program usage.  It aims to make it easier for your users to learn how to use your program.  If configured by the program, Active Help is printed when the user triggers shell completion.

For example, 
```
bash-5.1$ helm repo add [tab]
You must choose a name for the repo you are adding.

bash-5.1$ bin/helm package [tab]
Please specify the path to the chart to package

bash-5.1$ bin/helm package [tab][tab]
bin/    internal/    scripts/    pkg/     testdata/
```
## Supported shells

Active Help is currently only supported for the following shells:
- Bash (using [bash completion V2](shell_completions.md#bash-completion-v2) only). Note that bash 4.4 or higher is required for the prompt to appear when an Active Help message is printed.
- Zsh

## Adding Active Help messages

As Active Help uses the shell completion system, the implementation of Active Help messages is done by enhancing custom dynamic completions.  If you are not familiar with dynamic completions, please refer to [Shell Completions](shell_completions.md).

Adding Active Help is done through the use of the `cobra.AppendActiveHelp(...)` function, where the program repeatedly adds Active Help messages to the list of completions.  Keep reading for details.

### Active Help for nouns

Adding Active Help when completing a noun is done within the `ValidArgsFunction(...)` of a command.  Please notice the use of `cobra.AppendActiveHelp(...)` in the following example:

```go
cmd := &cobra.Command{
	Use:   "add [NAME] [URL]",
	Short: "add a chart repository",
	Args:  require.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addRepo(args)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		if len(args) == 0 {
			comps = cobra.AppendActiveHelp(comps, "You must choose a name for the repo you are adding")
		} else if len(args) == 1 {
			comps = cobra.AppendActiveHelp(comps, "You must specify the URL for the repo you are adding")
		} else {
			comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	},
}
```
The example above defines the completions (none, in this specific example) as well as the Active Help messages for the `helm repo add` command.  It yields the following behavior:
```
bash-5.1$ helm repo add [tab]
You must choose a name for the repo you are adding

bash-5.1$ helm repo add grafana [tab]
You must specify the URL for the repo you are adding

bash-5.1$ helm repo add grafana https://grafana.github.io/helm-charts [tab]
This command does not take any more arguments
```

### Active Help for flags

Providing Active Help for flags is done in the same fashion as for nouns, but using the completion function registered for the flag.  For example:
```go
_ = cmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 2 {
			return cobra.AppendActiveHelp(nil, "You must first specify the chart to install before the --version flag can be completed"), cobra.ShellCompDirectiveNoFileComp
		}
		return compVersionFlag(args[1], toComplete)
	})
```
The example above prints an Active Help message when not enough information was given by the user to complete the `--version` flag.
```
bash-5.1$ bin/helm install myrelease --version 2.0.[tab]
You must first specify the chart to install before the --version flag can be completed

bash-5.1$ bin/helm install myrelease bitnami/solr --version 2.0.[tab][tab]
2.0.1  2.0.2  2.0.3
```

## User control of Active Help

You may want to allow your users to disable Active Help or choose between different levels of Active Help.  It is entirely up to the program to define the type of configurability of Active Help that it wants to offer.  

### Configuration using an environment variable

One way to configure Active Help is to use the program's Active Help environment
variable.  That variable is named `<PROGRAM>_ACTIVE_HELP` where `<PROGRAM>` is the name of your 
program in uppercase with any `-` replaced by an `_`.  You can find that variable in the generated
completion scripts of your program.  The variable should be set by the user to whatever Active Help 
configuration values are supported by the program.

For example, say `helm` supports three levels for Active Help: `on`, `off`, `local`.  Then a user
would set the desired behavior to `local` by doing `export HELM_ACTIVE_HELP=local` in their shell.

When in `cmd.ValidArgsFunction(...)` or a flag's completion function, the program should read the
Active Help configuration from the `cmd.ActiveHelpConfig` field and select what Active Help messages
should or should not be added.

For example:
```go
ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	activeHelpLevel := cmd.ActiveHelpConfig
	
	var comps []string
	if len(args) == 0 {
		if activeHelpLevel != "off"  {
			comps = cobra.AppendActiveHelp(comps, "You must choose a name for the repo you are adding")
		}
	} else if len(args) == 1 {
		if activeHelpLevel != "off" {
			comps = cobra.AppendActiveHelp(comps, "You must specify the URL for the repo you are adding")
		}
	} else {
		if activeHelpLevel == "local" {
			comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
},
```
**Note 1**: If the string "0" is used for `cmd.Root().ActiveHelpConfig`, it will automatically be handled by Cobra and will completely disable all Active Help output (even if some output was specified by the program using the `cobra.AppendActiveHelp(...)` function).  Using "0" can simplify your code in situations where you want to blindly disable Active Help.

**Note 2**: Cobra transparently passes the `cmd.ActiveHelpConfig` string you specified back to your program when completion is invoked.  You can therefore define any scheme you choose for your program; you are not limited to using integer levels for the configuration of Active Help.  **However, the reserved "0" value can also be sent to you program and you should be prepared for it.**

**Note 3**: If a user wants to disable Active Help for every single program based on Cobra, the global environment variable `COBRA_ACTIVE_HELP` can be used as follows:
```
export COBRA_ACTIVE_HELP=0
```

### Configuration using a flag

Another approach for a user to configure Active Help is for the program to add a flag to the command that generates 
the completion script.  Using the flag, the user specifies the Active Help configuration that is
desired.  Then the program should specify that configuration by setting the `rootCmd.ActiveHelpConfig` string
before calling the Cobra API that generates the shell completion script.
The ActiveHelp configuration would then be read in `cmd.ValidArgsFunction(...)` or a flag's completion function, in the same 
fashion as explained above, using the same `cmd.ActiveHelpConfig` field.

For example, a program that uses a `completion` command to generate the shell completion script can add a flag `--activehelp-level` to that command.  The user would then use that flag to choose an Active Help level:
```
bash-5.1$ source <(helm completion bash --activehelp-level 1)
```
The code to pass that information to Cobra would look something like:
```go
cmd.Root().ActiveHelpConfig = strconv.Itoa(activeHelpLevel)
return cmd.Root().GenBashCompletionV2(out, true)
```

The advantage of using a flag is that it becomes self-documenting through Cobra's `help` command.  Also, it allows you to
use Cobra's flag parsing to handle the configuration value, instead of having to deal with a environment variable natively.

## Active Help with Cobra's default completion command

Cobra provides a default `completion` command for programs that wish to use it.
When using the default `completion` command, Active Help is configurable using the
environment variable approach described above.  You may wish to document this in more
details for your users.

## Debugging Active Help

Debugging your Active Help code is done in the same way as debugging the dynamic completion code, which is with Cobra's hidden `__complete` command.  Please refer to [debugging shell completion](shell_completions.md#debugging) for details.

When debugging with the `__complete` command, if you want to specify different Active Help configurations, you should use the active help environment variable (as you can find in the generated completion scripts).  That variable is named `<PROGRAM>_ACTIVE_HELP` where any `-` is replaced by an `_`.  For example, we can test deactivating some Active Help as shown below:
```
$ HELM_ACTIVE_HELP=1 bin/helm __complete install wordpress bitnami/h<ENTER>
bitnami/haproxy
bitnami/harbor
_activeHelp_ WARNING: cannot re-use a name that is still in use
:0
Completion ended with directive: ShellCompDirectiveDefault

$ HELM_ACTIVE_HELP=0 bin/helm __complete install wordpress bitnami/h<ENTER>
bitnami/haproxy
bitnami/harbor
:0
Completion ended with directive: ShellCompDirectiveDefault
```

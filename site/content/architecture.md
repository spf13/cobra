# Architecture for contributors

Cobra's core is one Go package in the repository root. The `doc/` directory is
another Go package that generates Markdown, man, reStructuredText, and YAML
documentation. Cobra uses [pflag](https://github.com/spf13/pflag) for flag
parsing; `command_win.go` and `command_notwin.go` provide platform-specific
startup behavior.

## Where to start

| Files | Responsibility |
| --- | --- |
| `command.go` | `Command`, command tree lookup, execution, flags, hooks, help, and usage. |
| `args.go` | Positional argument validators such as `NoArgs`, `ExactArgs`, and `MatchAll`. |
| `flag_groups.go` | Required-together, mutually exclusive, and one-required flag groups. |
| `completions.go` | The hidden `__complete` command, completion request handling, and the default `completion` command. |
| `bash_completions.go`, `bash_completionsV2.go`, `zsh_completions.go`, `fish_completions.go`, `powershell_completions.go` | Shell-specific completion script generation. |
| `active_help.go` | Active Help messages returned during completion. |
| `cobra.go` | Package-level settings, initializer/finalizer hooks, and template helpers. |
| `command_win.go`, `command_notwin.go` | Windows mousetrap handling and its non-Windows counterpart. |
| `doc/` | Documentation generators and their tests. |

Most behavior tests live beside the implementation in `*_test.go` files.
Changes to a shell's generated script should include coverage in that shell's
completion tests.

## How command execution works

`Execute` delegates to `ExecuteC`. `ExecuteC` redirects calls made on a child
command to the root, initializes the default help command and completion
commands, and chooses a command from the argument list with `Find` or, when
`TraverseChildren` is enabled, `Traverse`. It then calls `execute` on the
selected command.

`execute` initializes the help and version flags, parses the selected command's
flags, and handles help and version requests. For a runnable command, it runs
initializers, validates positional arguments, runs the first applicable
persistent pre-run hook (or every applicable hook when
`EnableTraverseRunHooks` is enabled), then the command's pre-run hook. It next
validates required flags and flag groups, invokes `Run` or `RunE`, and runs the
post-run and persistent post-run hooks. Errors are returned to `ExecuteC` for
the caller to handle. Finalizers run when the runnable command path exits,
including when validation or a hook returns an error.

The public `ExecuteContext` and `ExecuteContextC` methods set the context before
entering the same flow. `Find`, `Traverse`, and the completion helpers are
useful starting points when investigating command selection or shell
completion behavior.

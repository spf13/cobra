# Clink completions (Windows cmd.exe)

[Clink](https://github.com/chrisant996/clink) adds completion and line editing to
Windows cmd.exe. Cobra generates a standalone Lua script; no additional completion
bundle or Lua module is needed. Clink 1.3.3 or newer is required.

## Installation

Install Clink and start a cmd.exe session with Clink injected. For an application
named `myapp`, run the following from **cmd.exe**:

```bat
mkdir "%LOCALAPPDATA%\myapp-clink"
myapp completion clink > "%LOCALAPPDATA%\myapp-clink\myapp.lua"
clink installscripts "%LOCALAPPDATA%\myapp-clink"
```

Open a new Clink session and try `myapp <Tab>`. Keep `myapp.exe` on PATH.
The script also recognizes a quoted executable path. Do not put the script in a
lazy-loaded `completions` subdirectory: its generator must be registered before
the first completion request so it can set word boundaries correctly.

Use `myapp completion clink --no-descriptions` to omit descriptions. Applications
with a custom completion command can call:

```go
err := rootCmd.GenClinkCompletion(os.Stdout, true)
// Or write a file:
err = rootCmd.GenClinkCompletionFile("myapp.lua", true)
```

## Behavior

The script calls the executable's hidden `__complete` command (or
`__completeNoDesc`). Subcommands, flags, required flags, aliases and dynamic
completion callbacks use the same Cobra completion engine as other shells.
All `ShellCompDirective` values are supported: Error, NoSpace, NoFileComp,
FilterFileExt, FilterDirs (including a search directory), and KeepOrder.

Results are cached by exact request until the next prompt, so repeating Tab does
not repeatedly launch a process. New prompts refresh dynamic values. Word-break
and hint callbacks do not launch processes. Completion callbacks may still take
time when querying remote resources or devices.

On Clink 1.7 or newer, Active Help is displayed through the input hint row after
completion. Enable it using `clink set comment_row.show_hints true`. Multiple help
messages are joined with ` | `; they are never inserted as completion candidates.
Hints use already-fetched results and disappear when the input changes. Normal
Cobra Active Help environment settings apply. On Clink without hint support,
Active Help is disabled for the completion subprocess.

Arguments are individually quoted for cmd.exe. Literal double quotes, percent
signs, exclamation marks, NUL and line breaks cause a request to be declined,
since passing these through a shell can change its meaning. The script never
executes the edited command line, command-chain neighbors, or redirections.
Quoted paths, trailing backslashes, `--flag=value` and unquoted comma-separated
values are preserved. Completion cannot execute `.cmd`/`.bat` wrappers by name;
the generated matcher is intended for the application's native executable.

## Testing changes

`go test ./...` covers generation APIs, errors and completion command options.
With Lua 5.2+ on PATH it also executes the generated script's protocol tests.
On Windows, set `COBRA_CLINK` to the full path of `clink_x64.exe` to additionally
test native Clink word parsing and a real Cobra fixture program:

```powershell
$env:COBRA_CLINK = 'C:\Tools\Clink\clink_x64.exe'
go test -run Clink -v .
```

`COBRA_CLINK_FIXTURE` can point to an existing build of `./testdata/clink` when
test executables must use stable paths. The CI workflow runs these Windows checks
against a checksum-pinned portable Clink release. Interactive verification should
also check Tab insertion and enable the hint row to check Active Help rendering.

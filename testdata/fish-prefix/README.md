# Fish flag-prefix reproducer

This small Cobra program returns static candidates. It does not need kubectl,
Kubernetes, a server, or a kubeconfig.

From the Cobra repository root, build it and start Fish:

```sh
go build -o /tmp/cobra-fish-bin/fish-prefix ./testdata/fish-prefix
fish --no-config
```

In Fish:

```fish
set -gx PATH /tmp/cobra-fish-bin $PATH
fish-prefix completion fish | source
complete --do-complete 'fish-prefix --selector=app=w'
complete --do-complete 'fish-prefix -s app-name=w'
```

The first command should return `--selector=app=web` and
`--selector=app=worker`. The second should return `app-name=web` and
`app-name=worker`.

Before the fix, the first command returns `--selector=app=app=web` and
`--selector=app=app=worker`; the second returns `-name=app-name=web` and
`-name=app-name=worker`.

The Go backend can be checked directly:

```sh
/tmp/cobra-fish-bin/fish-prefix __complete --selector=app=w
```

It returns `app=web`, `app=worker`, and the `:4` no-file-completion directive
both before and after the fix. The corruption happens in the generated Fish
adapter.

Run the regression test with Fish installed:

```sh
go test -run '^TestFishCompletionFlagPrefix$' -count=1 -v .
```

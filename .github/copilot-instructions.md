# Contributing to Cobra

Follow [CONTRIBUTING.md](../CONTRIBUTING.md) and the current test workflow. Before adding a shell integration, read the existing generators, completion protocol, tests, and documentation under `site/content/completions/`.

- Follow the existing generator and test layout. Reuse Cobra's completion options and name protocol directives where the generated script interprets them.
- For a generated Windows shell script, verify the executable selection and argument parsing together. If the quoting assumes native Windows argv parsing, explicitly select the `.exe` rather than a shell wrapper. Document shell syntax that must remain in place to preserve quoting. Keep raw completion arguments out of diagnostic logs.
- Use the existing Go suite. Keep required tests independent of unrelated installed CLI applications and fixed local paths. Use small fixtures or captured output for protocol cases. Verify interactive behavior in the shell separately, and distinguish Go, runtime, and CI results in the PR.
- Keep review updates focused. Reply to each review point with the change, evidence, or remaining question.

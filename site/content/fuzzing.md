---
title: Fuzz testing
weight: 90
---

This project includes Go fuzz tests (Go 1.18+). You can run a short fuzz session locally with:

```
go test ./... -run=^$ -fuzz=Fuzz -fuzztime=30s
```

Continuous Integration runs a short fuzz pass on each push and pull request.

OSS-Fuzz integration

- Cobra is prepared for OSS-Fuzz via in-repo `Fuzz*` targets (see `fuzz/`).
- To onboard to OSS-Fuzz, open a PR to the upstream `google/oss-fuzz` repo creating `projects/cobra/` with:
  - `project.yaml` referencing Go as the language
  - `Dockerfile` installing dependencies and building the fuzz targets
  - `build.sh` that runs `compile_go_fuzzer` for each fuzz function (e.g., `FuzzLd`, `FuzzConfigEnvVar`)
- Reference: https://google.github.io/oss-fuzz/getting-started/new-project-guide/

Notes

- Keep fuzz targets small and deterministic with clear invariants.
- Prefer focusing on pure functions and parsers (e.g., helpers like `ld`, env var processing).


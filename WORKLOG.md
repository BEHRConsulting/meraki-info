# WORKLOG

Newest first. Each entry says **why**, not just what.

## 2026-10-16 — Documentation pass (overnight, branch `overnight-docs`)

**Why.** CLAUDE.md §11 asks for complete documentation in each language's own standard, and the
repo had one-line comments at best, a README that described behaviour the code no longer has,
and none of the bookkeeping files the bookkeeping profile expects. A reader (human or agent)
could not tell from the repo what actually works. This pass makes the documentation match the
code **as it is**, without changing any behaviour, so the code review that follows starts from
an honest description.

**What was done.**

- **GoDoc on every declaration** in all 10 Go files: package comments, every type, struct
  field, function and method (unexported ones too) and every test function, with
  parameters/returns and the *why* — for example that route sources are best-effort, that
  `calculateBackoff` gives 1s, 1s, 2s, and that consolidated `licenses`/`down`/`alerting`
  ignore `-org`. Existing comments were kept as the first line of each new comment.
- **PowerShell comment-based help** at the top of all four `.ps1` scripts and inside all 16
  functions; `Get-Help <script> -Full` now works. `REM` headers for the two `.bat` files and a
  header for the `Makefile`.
- **README** corrected where it disagreed with the code: Go version (1.24.2, from `go.mod`, not
  1.21), the missing `alerting` command, options-before-command, what `-all` really does, the
  "File Naming" section (automatic timestamped names were removed in `f376ff6`), OAuth2 (not
  reachable from the CLI), and rate-limit behaviour.
- **Created** `AGENTS.md` (claude-config core block + the five sections + "Not done yet"),
  `VERSIONS.md` (1.0.0, the version the code reports), `features.yaml` (FEAT-0001…0011) and this
  file. Statuses are deliberately conservative: nothing was run against the live API, so every
  API-backed command is `in-test`, and `-all`, output formats and OAuth2 are `in-dev` because of
  known gaps written in their notes.

**How it was verified.** A token-stream comparison (Go scanner, comments dropped) of every `.go`
file against tag `pre-docs` shows no non-comment change; the same check with the PowerShell
tokenizer on the `.ps1` files; `git diff` of the `.bat` files adds only `REM` lines. `go build`,
`go vet`, `go test ./...` and `make build`/`make test` pass before and after; `gofmt -l` is
clean; PSScriptAnalyzer findings went from 149 to 147 (the missing-help finding is gone).

**Not done.** No behaviour was changed; the problems found while documenting are left for the
code review on branch `code-review`.

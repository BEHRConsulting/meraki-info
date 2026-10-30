# WORKLOG

Newest first. Each entry says **why**, not just what.

## 2026-10-16 — Code review and safe fixes (overnight, branch `code-review`)

**Why.** The documentation pass surfaced behaviour that contradicts the tool's own help text and
README, and output paths that could fail silently. Kent was away, so the rule was: fix only what
is clearly wrong and cannot change intended behaviour, and turn everything else into a question.
Full write-up: `docs/code-review/2026-10-16.md`.

**What was fixed.**

- **`-org` was ignored by the consolidated `licenses`, `down` and `alerting` reports** (the
  default path when `-network` is not given), so `-org X down` listed every organization's
  devices. `route-tables` already honored it and the `-all` help text promises it. New helper
  `organizationsToReport` + `main_test.go`.
- **CSV flush and file-close errors were dropped**, so a full disk or broken pipe produced an
  empty or truncated file with exit status 0. Test written first (`failingWriter`); it failed
  for all four CSV record types before the fix.
- **`access` printed usage examples that cannot work** (old binary name `meraki-routes-backup`,
  a nonexistent `--access` flag, no command).
- **`make access` / `.\make.ps1 access` always failed** ("only one command is allowed") because
  they put `--apikey` after the command; they now rely on `MERAKI_APIKEY`, which also keeps the
  key out of the process list.
- **`PowerShell-Make-Function.ps1` errored on every dot-source** (`Export-ModuleMember` outside a
  module); the line is removed.

**What was deliberately not fixed**, because each changes behaviour or output format: `-all
-output FILE` overwriting one file per network, JSON emitted for consolidated XML/CSV, no
pagination, `Retry-After` ignored, whether `down` sees live statuses, CSV formula injection,
dead OAuth2 code. Each is a numbered question in the review with a recommendation.

**Verification.** `go build`, `go vet`, `gofmt -l`, `go test ./...` (83 passing, 0 failing,
up from 77), `make build`; golangci-lint default errcheck findings 22 → 19; PSScriptAnalyzer
unchanged at 147; nothing was run against the live API.

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

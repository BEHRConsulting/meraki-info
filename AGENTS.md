<!-- AGENTS-CORE:BEGIN v1 -->
<!-- Shared rules, carried as a replaceable block. Do not edit inside these markers: an
     update replaces everything between them. Repo-specific content goes BELOW the END
     marker, where it is never overwritten. -->

## What this file is for

`CLAUDE.md` holds the rules that do not change between repos. **This file holds everything
that does**: what this project is, where its code lives, how to build and test it, and the
guardrails specific to it. Where the two disagree, this file wins — it is closer to the work.

A session that reads this file should be able to make a correct change without opening
anything else first. Judge it by that, not by length.

## Required sections

Keep these, in this order, and say so plainly when one does not apply yet:

1. **What this is** — one paragraph. What the thing does, for whom, and the single most
   important constraint on it.
2. **Layout** — the files that matter and what each is responsible for. Not a `tree` dump;
   the map someone needs to find the right file on the first try.
3. **Commands** — how to install, run, test, and lint. Copy-pasteable. If a command must be
   run from a particular directory or with particular credentials, say so here.
4. **Conventions** — language level, style, dependency policy, anything a change would
   otherwise get wrong and only discover in review.
5. **Guardrails** — the things that cost money, touch production, or cannot be undone. Each
   one with the reason, because a rule without a reason gets optimised away.

## Bookkeeping

Whatever this repo uses to record what changed and why — a worklog, a changelog, a feature
manifest — **updating it is part of the change, not a follow-up task.** Name those files
here so nobody has to guess.

Write the *why*. The diff already shows what changed; the reasoning — the failure that
motivated a guard, the cheaper approach that was rejected and why — is the part that cannot
be recovered later.

If a change is too small to be worth an entry, say so in one line rather than skipping
silently.

## Keeping this file true

A stale `AGENTS.md` is worse than none, because it is the thing people trust *instead of*
looking. When a change makes a statement here wrong, fix it in the same change.

Most valuable and most often missing: **say what is NOT done.** A known limitation written
down is honest; one left out reads as a promise.
<!-- AGENTS-CORE:END v1 -->

# AGENTS.md — meraki-info

## 1. What this is

`meraki-info` is a Go command-line tool, written for BCI/Kent's own network administration, that
reads the Cisco Meraki Dashboard API v1 with an API key and reports what one key can reach
(`access`), route tables (`route-tables`), licenses (`licenses`), and down or alerting devices
(`down`, `alerting`) as text, JSON, XML or CSV. The most important constraint: **it is read-only
and must stay that way** — every API call is an HTTP GET, and it runs against live production
Meraki organizations.

## 2. Layout

| Path | Responsibility |
| --- | --- |
| `main.go` | Entry point. Command dispatch, the per-network and consolidated (`-all`) loops, and the `access` report. Most of the duplication lives here (four near-identical copies per command). |
| `internal/config/config.go` | Flags + `MERAKI_*` environment variables → validated `Config`. Options must precede the single positional command. |
| `internal/logger/logger.go` | Sets the default `log/slog` logger: JSON on stderr. |
| `internal/meraki/client.go` | API client: `makeRequest` (auth header, retry/backoff), route collection from five sources, licenses, devices, name→ID resolution. Also the record types (`Route`, `License`, `Device`, `*WithNetwork`). |
| `internal/output/writer.go` | `Writer` implementations for text, JSON, XML, CSV, and the XML mirror structs. |
| `*_test.go` | Unit tests. API tests run against `httptest` servers by setting the unexported `baseURL`; nothing calls the real API. `main` has no tests. |
| `Makefile`, `make.ps1`, `make.bat`, `build.ps1`, `build-windows.ps1`, `build-windows.bat`, `PowerShell-Make-Function.ps1` | Build wrappers. `make.ps1` mirrors the Makefile for Windows. |
| `docs/code-review/` | Dated code reviews with open questions for Kent. |
| `README.md` | User documentation. `BUILD_SCRIPTS.md`, `MAKE_SUPPORT.md`, `ARM64_SUPPORT.md`, `MAKEFILE_UPDATES.md`, `RETRY_LOGIC.md` are older topic notes. |
| `REQUIREMENTS.md`, `prompts-initial.md`, `prompts-claude.md`, `CODE_REVIEW.md` | The original AI-generation prompts and requirements (historical). |

## 3. Commands

Run from the repository root. None of these need an API key or touch the network except where
marked.

```sh
go build -o meraki-info .     # or: make build
go test ./...                 # or: make test   (make test-v, make coverage)
go vet ./...
gofmt -l .                    # must print nothing
make build-all                # cross-compile: linux/amd64+arm64, windows/amd64, darwin/amd64+arm64
./meraki-info --help          # usage; exits 0
```

`golangci-lint run ./...` works but the repo has **no** `.golangci.yml`, so it runs the default
linters; on 2026-10-16 that reported 22 `errcheck` findings, all pre-existing (unchecked
`fmt.Fprint*`, `Close`, `w.Write`, and `os.Setenv`/`os.Remove` in tests). Treat it as advisory until a config is chosen.

PowerShell (Windows, or `pwsh` on macOS): `.\make.ps1 help`, `.\make.ps1 test`,
`.\build.ps1 -Help`. `Get-Help .\make.ps1 -Full` shows the comment-based help.

**Live API (needs a real key; do not run unattended):** `make access`, `.\make.ps1 access`, and
any `meraki-info` command other than `--help`.

## 4. Conventions

- Go: the `go` line in `go.mod` is `1.24.2`; this machine builds with Go 1.26.7 (pinned
  machine-wide). Do not change `go.mod`'s version or toolchain casually.
- One dependency: `golang.org/x/oauth2` (only for the unused `NewClientWithOAuth2`). Prefer the
  standard library.
- Logging is `log/slog`, JSON to stderr; command output goes to stdout so it can be piped.
- Errors are wrapped with `fmt.Errorf("...: %w", err)`. Route sources and per-network work are
  deliberately best-effort: failures are logged and skipped, not returned.
- Documentation standard (CLAUDE.md §11): GoDoc on every declaration including unexported ones
  and struct fields; comment-based help (`.SYNOPSIS` …) in every PowerShell script and function;
  `REM` headers in batch files.
- Tests: table-driven, `httptest` servers for API calls, never the real API.

## 5. Guardrails

- **Never call the real Meraki API from tests, CI or an unattended session.** Keys reach
  production organizations; a misused key can be rate-limited or revoked for every user of
  that organization.
- **Never commit an API key or `.env`.** `.env` is git-ignored; `.env.example` holds
  placeholders only. Prefer `MERAKI_APIKEY` over `-apikey` so the key does not show in `ps`
  output or shell history.
- **Keep the client read-only.** Adding a non-GET request would let a reporting tool change
  live network configuration.
- `.gitignore` ignores every `*.json`, `*.txt`, `*.xml` and `*.csv` file (so command output is
  never committed). A test fixture or config in those formats must be force-added or the
  ignore narrowed.

## Bookkeeping

`WORKLOG.md` (newest first, the why), `VERSIONS.md` (what shipped), `features.yaml` (what
exists, `FEAT-NNNN` ids — never renumber). Update them in the same change.

## Not done yet

- `-all` with `-output FILE` writes every network to the same file, so only the last network
  survives (the timestamped per-network names were removed in `f376ff6`).
- Consolidated `licenses`/`down`/`alerting` in `-format xml` or `csv` are emitted as JSON.
- No pagination: list endpoints are read as one page.
- `Retry-After` is not honored on 429.
- `down` relies on the `status` field of `GET /networks/{id}/devices`; `alerting` joins the
  organization device-statuses endpoint instead. Whether `down` sees real statuses has not been
  verified against the live API.
- OAuth2 is not reachable from the CLI.
- See `docs/code-review/2026-10-16.md` for the full list of findings and open questions.

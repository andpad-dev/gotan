This repository contains the Go Conference 2026 workshop materials for Go探無比 (gotan). Apply these rules when creating or editing scenarios.

## File Layout

- Store each scenario at `<category>/<difficulty>/<topic>/README.md`; use one directory per scenario.
- Categories: `01-packages`, `02-features`, `03-cmd-tools`, `04-deep-dive`.
- Difficulties: `01-beginner`, `02-intermediate`, `03-advanced`.
- Use a descriptive English kebab-case topic name, such as `fmt-printf` or `slog-handler`; do not add sequence numbers.
- Do not create separate solution files or directories. Put hints and answers in the scenario README inside `<details>` blocks.
- Each category root must have a README with reverse-search guidance. If the target category lacks one, propose creating it.

## Difficulty and Scenario Format

The directory and the PR label must express the same difficulty: `初級`, `中級`, or `上級`.

| Level | Directory | Story guidance | Questions |
| --- | --- | --- | --- |
| 初級 | `01-beginner` | 目の前のコードや出力から用語・基本仕様を調べる | About 2 |
| 中級 | `02-intermediate` | 実現したいことと制約から具体的な方法を調べる | 3 |
| 上級 | `03-advanced` | 現象の背景にある仕様・設計・歴史を調べる | No fixed limit |

Examples: [beginner #2](https://github.com/andpad-dev/gotan/pull/2), [intermediate #3](https://github.com/andpad-dev/gotan/pull/3), [advanced #4](https://github.com/andpad-dev/gotan/pull/4).

## Scenario README

Use this structure:

1. `# <title>`.
2. One execution-environment badge, so readers can tell before opening whether they need Go installed.
3. An introduction framed as a workplace problem. The observation, purpose, and investigation should connect naturally without requiring a fixed sentence. Link every runnable Go example to Go Playground.
4. `## 設問 N: <heading>` sections, using the question count for the difficulty. Each section contains the problem, an optional `<details><summary>ヒント</summary>` hint, and an `<details><summary>答え</summary>` explanation with **調査ルート** (primary sources and order) and **答え**.
5. An optional `<details><summary>こぼれ話</summary>` section for interesting but nonessential findings.
6. `## 調査の入り口` with primary-source starting links.

Formatting:

- Separate question, trivia, and entry-point sections with `---`.
- Do not use Markdown headings inside `<details>`; use bold labels instead.
- Keep common investigation triggers, such as “press `f` to search first,” visible in the category README rather than hiding them in hints.

Go-version badge:

| Case | Message | Color |
| --- | --- | --- |
| Go version has a minimum requirement | `Go <x.y> 以上` (match the `go` line in the scenario's `go.mod`) | `F39C12` |
| Go version is fixed to one release | `Go <x.y>` | `F39C12` |
| Go version is not specified | `指定なし` | `9E9E9E` |
| No code to run | `不要` | `9E9E9E` |

Write it as `![実行環境: <message>](https://img.shields.io/badge/実行環境-<message>-<color>)` with the label and message percent-encoded, and keep the alt text identical to the badge. Add anything the badge cannot carry as one line below it.

Then add a row to the category table in `workshop/SCENARIOS.md`, pasting the **same badge markdown verbatim**. The two must match character for character, so a mismatch is easy to spot.

## Quality and Research

- Write learner-facing problem statements in Japanese.
- Verify every referenced URL with `curl` or an equivalent HTTP-status check before committing it.
- Run every code snippet with `go run`; copy its actual output into the README. Link all problem code and execution examples to Go Playground:
  `curl -s -X POST --data-binary @main.go https://play.golang.org/share` → `https://go.dev/play/p/<id>`.
- Treat Go 1.27 (`go1.27`) as the current version for this repository. For recent-feature scenarios, prioritize features introduced in Go 1.26 or Go 1.27.
- Never conclude from caches, search indexes, snippets, or memory alone. Trace every technical claim to a primary source.
- Primary sources include `go.dev` (including `/doc`, `/blog/all`, `/ref/spec`, `/talks`, `/pkg`), `pkg.go.dev`, `research.swtch.com`, `golang.design/history`, `dev.golang.org`, `build.golang.org`, `farmer.golang.org`, `github.com/golang/go`, `github.com/golang/proposal`, `github.com/golang/go/wiki`, `go-review.googlesource.com`, `groups.google.com/g/golang-nuts`, `groups.google.com/g/golang-dev`, and `groups.google.com/g/golang-tools`.
- Follow relevant links up to two hops from the starting page and use the most authoritative source reached.
- Start recent-feature research with the [Go 1.26](https://go.dev/doc/go1.26) and [Go 1.27](https://go.dev/doc/go1.27) release notes. If the release-note web source contains an issue number, find the corresponding GitHub issue, read the entire issue—including discussion and linked context—and use it to validate the conclusion. Do not infer behavior from the release-note summary alone.
- Never call a Go 1.25-or-earlier feature “latest.” Before doing so, inspect the [Go 1.21](https://go.dev/doc/go1.21), [Go 1.22](https://go.dev/doc/go1.22), [Go 1.23](https://go.dev/doc/go1.23), [Go 1.24](https://go.dev/doc/go1.24), and [Go 1.25](https://go.dev/doc/go1.25) release notes; a feature documented there is old for this purpose.
- When using `pkg.go.dev`, read the linked `cs.opensource.google/go/go` source code and its surrounding context. Do not rely on pkg.go.dev prose alone.
- Link Go source files with version tags, such as `refs/tags/go1.26.5`.

## Pull Requests

- Create `feature/scenario-<topic>` from `origin/main`.
- Open a Draft PR and apply exactly one difficulty label matching the directory.
- Follow the PR template. Do not include brainstorming notes or process history.
- Use Conventional Commits, for example: `feat: 中級問題「◯◯」を追加`.

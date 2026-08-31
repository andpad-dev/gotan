#!/usr/bin/env python3
"""Check deterministic structure and route preconditions for one scenario."""

from __future__ import annotations

import argparse
import re
import urllib.parse
from pathlib import Path


SCENARIO_PATH_RE = re.compile(
    r"^workshop/(01-packages|02-features|03-cmd-tools|04-deep-dive)/"
    r"(01-beginner|02-intermediate|03-advanced)/([^/]+)/README\.md$"
)
QUESTION_RE = re.compile(r"(?m)^## 設問\s+(\d+)(?::|：|\s)")
URL_RE = re.compile(r"https?://[^\s<>\]})]+")
DETAIL_RE = re.compile(r"<details\b[^>]*>(.*?)</details>", re.IGNORECASE | re.DOTALL)
FENCE_RE = re.compile(r"^\s*(?:>\s*)?(```|~~~)")
LOCAL_PACKAGE_COMMAND_RE = re.compile(
    r"`go\s+(?:run|build|test|vet)\b[^`\n]*(?:\./\.\.\.|(?<!\S)\.)(?:\s[^`\n]*)?`"
)


def emit(level: str, message: str, line: int | None = None) -> bool:
    location = f"line {line}: " if line is not None else ""
    print(f"{level}\t{location}{message}")
    return level == "FAIL"


def line_number(text: str, offset: int) -> int:
    return text.count("\n", 0, offset) + 1


def question_blocks(text: str) -> list[tuple[int, int, str]]:
    matches = list(QUESTION_RE.finditer(text))
    blocks: list[tuple[int, int, str]] = []
    for index, match in enumerate(matches):
        end = matches[index + 1].start() if index + 1 < len(matches) else len(text)
        blocks.append((int(match.group(1)), line_number(text, match.start()), text[match.start():end]))
    return blocks


def first_route_url(block: str) -> str | None:
    route_start = re.search(r"調査ルート", block)
    if not route_start:
        return None
    answer_start = re.search(r"(?:\*\*答え\*\*|<summary>答え</summary>)", block[route_start.end():])
    route = block[route_start.end():]
    if answer_start:
        route = route[:answer_start.start()]
    match = URL_RE.search(route)
    return match.group(0).rstrip(".,;:!?。 、」』）)]}") if match else None


def first_url(text: str) -> str | None:
    match = URL_RE.search(text)
    return match.group(0).rstrip(".,;:!?。 、」』）)]}\"'`") if match else None


def contains_markdown_heading(content: str) -> bool:
    in_fence = False
    for line in content.splitlines():
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if not in_fence and re.match(r"^#{1,6}\s", line):
            return True
    return False


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("scenario", type=Path, help="scenario README under workshop/")
    args = parser.parse_args()

    path = args.scenario.resolve()
    try:
        relative = path.relative_to(Path.cwd().resolve()).as_posix()
    except ValueError:
        relative = args.scenario.as_posix()

    failures = 0
    match = SCENARIO_PATH_RE.match(relative)
    if not match:
        failures += emit("FAIL", "path must be workshop/<category>/<difficulty>/<topic>/README.md")
        return 1

    _category, difficulty, _topic = match.groups()
    category_readme = path.parent.parent.parent / "README.md"
    if not category_readme.is_file():
        failures += emit("FAIL", f"category README is missing: {category_readme}")
    else:
        emit("PASS", f"category README exists: {category_readme}")
        category_text = category_readme.read_text(encoding="utf-8")
        category_url = first_url(category_text)
        if category_url is None:
            failures += emit("FAIL", "category README has no external starting URL")
        elif urllib.parse.urlsplit(category_url).hostname != "go.dev":
            failures += emit("FAIL", f"category README starts at {category_url}; first URL must be go.dev")
        else:
            emit("PASS", f"category README starts at go.dev: {category_url}")
        if "逆引き" not in category_text or not re.search(r"検索|ページ内|\bf\b", category_text):
            failures += emit("FAIL", "category README lacks a usable reverse-search procedure")

    text = path.read_text(encoding="utf-8")
    if not re.search(r"(?m)^#\s+\S", text):
        failures += emit("FAIL", "scenario title heading is missing")
    if "## 調査の入り口" not in text:
        failures += emit("FAIL", "## 調査の入り口 is missing")

    question_matches = list(QUESTION_RE.finditer(text))
    question_numbers = [int(match.group(1)) for match in question_matches]
    expected_numbers = list(range(1, len(question_numbers) + 1))
    if question_numbers != expected_numbers:
        failures += emit("FAIL", f"question numbers are not contiguous from 1: {question_numbers}")
    elif question_numbers:
        emit("PASS", f"question sections found: {question_numbers}")
    else:
        failures += emit("FAIL", "no ## 設問 N section found")

    introduction_end = question_matches[0].start() if question_matches else len(text)
    introduction = re.sub(r"(?m)^#\s+\S.*$", "", text[:introduction_end]).strip()
    if not introduction:
        failures += emit("FAIL", "scenario introduction is missing")
    else:
        emit("PASS", "scenario introduction is present; review its natural context manually")

    for number, line, block in question_blocks(text):
        if not re.search(r"<summary>答え</summary>|\*\*答え\*\*", block):
            failures += emit("FAIL", f"question {number} has no answer section", line)
        if not re.search(r"\*\*調査ルート\*\*", block):
            failures += emit("FAIL", f"question {number} has no 調査ルート", line)
        route_url = first_route_url(block)
        if route_url is None:
            failures += emit("FAIL", f"question {number} has no URL in 調査ルート", line)
        else:
            host = urllib.parse.urlsplit(route_url).hostname
            if host != "go.dev":
                failures += emit("FAIL", f"question {number} route starts at {route_url}; first URL must be go.dev", line)
            else:
                emit("PASS", f"question {number} route starts at go.dev: {route_url}", line)

    entry_start = text.find("## 調査の入り口")
    if entry_start >= 0:
        entry_end = text.find("\n## ", entry_start + 3)
        entry = text[entry_start:] if entry_end < 0 else text[entry_start:entry_end]
        entry_url = first_url(entry)
        if entry_url is None:
            failures += emit("FAIL", "## 調査の入り口 has no external starting URL")
        elif urllib.parse.urlsplit(entry_url).hostname != "go.dev":
            failures += emit("FAIL", f"## 調査の入り口 starts at {entry_url}; first URL must be go.dev")

    for detail in DETAIL_RE.finditer(text):
        if contains_markdown_heading(detail.group(1)):
            failures += emit("FAIL", "Markdown heading found inside <details>", line_number(text, detail.start()))

    if re.search(r"```go\b", text) and "https://go.dev/play/p/" not in text:
        emit("WARN", "Go code block exists without a Go Playground link; verify whether it is runnable")
    elif re.search(r"```go\b", text):
        emit("WARN", "Go code blocks found; compare each block with its individual Playground link")

    local_commands = sorted({match.group(0).strip("`") for match in LOCAL_PACKAGE_COMMAND_RE.finditer(text)})
    if local_commands:
        go_files = sorted(path.parent.glob("*.go"))
        if go_files:
            emit("PASS", f"local package commands have {len(go_files)} sibling .go file(s)")
        else:
            emit(
                "WARN",
                "local package command is shown but the scenario directory has no .go files; "
                "verify complete save/setup steps: " + ", ".join(local_commands),
            )

    if "<summary>ヒント</summary>" in text or "### ヒント" in text:
        emit("WARN", "hint leakage still requires manual comparison against the answer")

    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main())

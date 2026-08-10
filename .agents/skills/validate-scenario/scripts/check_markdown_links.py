#!/usr/bin/env python3
"""Run a first-pass link check for Markdown files used by a scenario.

The script checks relative targets locally and, with --network, follows HTTP
redirects. It deliberately does not decide whether a reachable URL is the
right source; that semantic check belongs to the scenario audit.
"""

from __future__ import annotations

import argparse
import re
import sys
import unicodedata
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path


MARKDOWN_LINK_RE = re.compile(r"(?<!!)\[[^\]]*\]\(\s*<?([^\s>)]*)>?")
AUTOLINK_RE = re.compile(r"<((?:https?://)[^>]+)>")
HTML_HREF_RE = re.compile(r"\bhref\s*=\s*[\"']([^\"']+)[\"']", re.IGNORECASE)
BARE_URL_RE = re.compile(r"https?://[^\s<>\]})]+")
TRAILING_PUNCTUATION = ".,;:!?。 、」』）)]}"
INCONCLUSIVE_STATUSES = {403, 429}


@dataclass(frozen=True)
class Link:
    value: str
    line: int
    kind: str


def strip_trailing_punctuation(url: str) -> str:
    while url and url[-1] in TRAILING_PUNCTUATION:
        url = url[:-1]
    return url


def extract_links(path: Path) -> list[Link]:
    links: list[Link] = []
    for line_number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        candidates: list[tuple[str, str]] = []
        candidates.extend((match.group(1), "markdown") for match in MARKDOWN_LINK_RE.finditer(line))
        candidates.extend((match.group(1), "autolink") for match in AUTOLINK_RE.finditer(line))
        candidates.extend((match.group(1), "html") for match in HTML_HREF_RE.finditer(line))
        candidates.extend((strip_trailing_punctuation(match.group(0)), "bare") for match in BARE_URL_RE.finditer(line))
        links.extend(Link(value, line_number, kind) for value, kind in candidates if value)

    unique: dict[tuple[str, int], Link] = {}
    for link in links:
        unique.setdefault((link.value, link.line), link)
    return list(unique.values())


def github_slug(value: str) -> str:
    value = unicodedata.normalize("NFKC", value).casefold()
    value = "".join(char for char in value if not unicodedata.category(char).startswith("P"))
    value = re.sub(r"\s+", "-", value)
    return value.strip("-")


def local_anchor_exists(path: Path, fragment: str) -> bool | None:
    """Return True/False, or None when the renderer is too ambiguous to know."""
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return None

    escaped = re.escape(fragment)
    if re.search(rf"(?:id|name)\s*=\s*[\"']{escaped}[\"']", text, re.IGNORECASE):
        return True

    headings = re.findall(r"(?m)^#{1,6}\s+(.+?)\s*#*\s*$", text)
    if fragment in {github_slug(heading) for heading in headings}:
        return True
    return None


def check_local(path: Path, link: Link) -> tuple[str, str]:
    parsed = urllib.parse.urlsplit(link.value)
    if parsed.scheme or link.value.startswith("//"):
        return "SKIP", "external"

    target = Path(urllib.parse.unquote(parsed.path)) if parsed.path else path
    if not target.is_absolute():
        target = (path.parent / target).resolve()
    if not target.exists():
        return "FAIL", f"missing target: {target}"
    if not parsed.fragment:
        return "PASS", f"target exists: {target}"

    anchor = local_anchor_exists(target, parsed.fragment)
    if anchor is True:
        return "PASS", f"target and anchor exist: #{parsed.fragment}"
    if anchor is None:
        return "INCONCLUSIVE", f"target exists; inspect anchor manually: #{parsed.fragment}"
    return "FAIL", f"anchor not found: #{parsed.fragment}"


def check_http(value: str, timeout: float) -> tuple[str, str]:
    request = urllib.request.Request(
        value,
        headers={"User-Agent": "gotan-validate-scenario/1.0"},
        method="GET",
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            status = response.status
            final_url = response.geturl()
    except urllib.error.HTTPError as exc:
        status = exc.code
        final_url = exc.geturl() or value
    except ValueError as exc:
        return "FAIL", f"invalid URL: {exc}"
    except (urllib.error.URLError, TimeoutError) as exc:
        return "INCONCLUSIVE", f"network request could not be completed: {exc}"

    if 200 <= status < 300:
        detail = f"HTTP {status}"
        if final_url != value:
            detail += f" -> {final_url}"
        return "PASS", detail
    if status in INCONCLUSIVE_STATUSES:
        return "INCONCLUSIVE", f"HTTP {status}"
    return "FAIL", f"HTTP {status} -> {final_url}"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("paths", nargs="+", type=Path, help="Markdown files to inspect")
    parser.add_argument(
        "--network",
        action="store_true",
        help="follow HTTP redirects and check external URLs",
    )
    parser.add_argument("--timeout", type=float, default=15.0, help="HTTP timeout in seconds")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    failures = 0
    inconclusive = 0
    for raw_path in args.paths:
        path = raw_path.resolve()
        if not path.is_file():
            print(f"FAIL\t{raw_path}: file does not exist", file=sys.stderr)
            failures += 1
            continue
        for link in extract_links(path):
            parsed = urllib.parse.urlsplit(link.value)
            if parsed.scheme in {"http", "https"}:
                if not args.network:
                    result, detail = "SKIP", "external; rerun with --network"
                else:
                    result, detail = check_http(link.value, args.timeout)
            else:
                result, detail = check_local(path, link)
            print(f"{result}\t{path}:{link.line}\t{link.kind}\t{link.value}\t{detail}")
            if result == "FAIL":
                failures += 1
            elif result == "INCONCLUSIVE":
                inconclusive += 1

    if failures:
        return 1
    if inconclusive:
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

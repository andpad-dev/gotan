#!/usr/bin/env python3
"""Verify that Go Playground links expose retrievable shared source code."""

from __future__ import annotations

import argparse
import hashlib
import re
import urllib.error
import urllib.request
from pathlib import Path


PLAYGROUND_RE = re.compile(r"https://go\.dev/play/p/([A-Za-z0-9_-]+)(?:\?[^\s)>]+)?")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--network", action="store_true", help="fetch shared source from go.dev")
    parser.add_argument("paths", nargs="+", type=Path)
    args = parser.parse_args()

    failures = 0
    inconclusive = 0
    seen: set[tuple[Path, int, str]] = set()
    for path in args.paths:
        text = path.read_text(encoding="utf-8")
        for line_number, line in enumerate(text.splitlines(), 1):
            for match in PLAYGROUND_RE.finditer(line):
                url = match.group(0)
                key = (path, line_number, url)
                if key in seen:
                    continue
                seen.add(key)
                source_url = f"https://go.dev/play/p/{match.group(1)}.go"
                if not args.network:
                    print(f"SKIP\t{path}:{line_number}\t{url}\tnetwork check disabled")
                    inconclusive += 1
                    continue
                try:
                    request = urllib.request.Request(source_url, headers={"User-Agent": "gotan-validate-scenario"})
                    with urllib.request.urlopen(request, timeout=15) as response:
                        source = response.read()
                        status = response.status
                except (urllib.error.HTTPError, urllib.error.URLError, TimeoutError) as exc:
                    print(f"FAIL\t{path}:{line_number}\t{url}\t{exc}")
                    failures += 1
                    continue
                if status < 200 or status >= 300:
                    print(f"FAIL\t{path}:{line_number}\t{url}\tHTTP {status}")
                    failures += 1
                    continue
                digest = hashlib.sha256(source).hexdigest()[:12]
                print(f"PASS\t{path}:{line_number}\t{url}\tsource HTTP {status}\tsha256={digest}\tbytes={len(source)}")

    if failures:
        return 1
    return 2 if inconclusive else 0


if __name__ == "__main__":
    raise SystemExit(main())

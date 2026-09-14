#!/usr/bin/env python3
"""Check local inline Markdown links without fetching external resources."""

import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit


ROOT = Path(__file__).resolve().parents[1]
LINK = re.compile(r"\[[^]\n]+\]\(\s*(<[^>]+>|[^\s)]+)(?:\s+[^)]*)?\)")
REFERENCE = re.compile(r"^\s{0,3}\[[^]]+\]:\s*(<[^>]+>|\S+)", re.MULTILINE)
FENCE = re.compile(r"^\s{0,3}(`{3,}|~{3,})")
INLINE_CODE = re.compile(r"(`+)(.+?)\1")


def without_fences(content):
    lines = []
    marker = None
    for line in content.splitlines():
        match = FENCE.match(line)
        if match:
            fence = match.group(1)
            if marker is None:
                marker = fence
            elif fence[0] == marker[0] and len(fence) >= len(marker):
                marker = None
            lines.append("")
        else:
            lines.append("" if marker else line)
    return "\n".join(lines)


def validate(root=ROOT):
    errors = []
    for source in sorted(root.rglob("*.md")):
        if ".git" in source.relative_to(root).parts:
            continue
        relative = source.relative_to(root)
        try:
            content = without_fences(source.read_text(encoding="utf-8"))
        except (OSError, UnicodeError) as exc:
            errors.append(f"{relative}: cannot read UTF-8 text: {exc}")
            continue
        for number, line in enumerate(content.splitlines(), 1):
            line = INLINE_CODE.sub("", line)
            targets = [m.group(1) for m in LINK.finditer(line)]
            targets += [m.group(1) for m in REFERENCE.finditer(line)]
            for target in targets:
                target = target.strip("<>")
                try:
                    parsed = urlsplit(target)
                except ValueError as exc:
                    errors.append(f"{relative}:{number}: invalid link target {target}: {exc}")
                    continue
                if parsed.scheme or parsed.netloc or target.startswith("#"):
                    continue
                path = unquote(parsed.path)
                if not path:
                    continue
                if path.startswith("/"):
                    errors.append(f"{relative}:{number}: absolute local link is unsupported: {target}")
                    continue
                resolved = (source.parent / path).resolve()
                try:
                    resolved.relative_to(root.resolve())
                except ValueError:
                    errors.append(f"{relative}:{number}: local link escapes repository: {target}")
                    continue
                if not resolved.exists():
                    errors.append(f"{relative}:{number}: broken local link: {target}")
    return sorted(errors)


def main():
    errors = validate()
    if errors:
        for error in errors:
            print(f"ERROR: {error}", file=sys.stderr)
        return 1
    print("Repository-local Markdown links passed (external URLs and fragments not fetched)")
    return 0


if __name__ == "__main__":
    sys.exit(main())

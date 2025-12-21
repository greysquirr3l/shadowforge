#!/usr/bin/env python3
"""Generate SVG images from display equations embedded in a Markdown document.

Default behavior:
- Scan a Markdown file for display math blocks delimited by `$$ ... $$`.
- Render each block to a tightly-cropped transparent SVG.
- Write SVGs into an output directory (default: `<md_dir>/equations`).

Optional:
- `--inplace` rewrites the Markdown file, replacing each `$$ ... $$` block with
    a Markdown image reference to the generated SVG.

Notes:
- Uses Matplotlib's built-in mathtext renderer (subset of LaTeX). Some macros
    (e.g., certain AMS commands) may not render; simplify those equations.
- Inline math `$...$` is not processed by default.
"""

from __future__ import annotations

import argparse
import hashlib
import re
import time
from dataclasses import dataclass
from pathlib import Path

import matplotlib

# Use non-interactive backend (must be set before importing pyplot)
matplotlib.use("Agg")

import matplotlib.pyplot as plt


@dataclass(frozen=True)
class MathBlock:
    raw: str
    content: str
    start: int
    end: int


_FENCED_CODE_BLOCK_RE = re.compile(r"^\s*(```|~~~)")


def _is_escaped(markdown_text: str, index: int) -> bool:
    # Treat any odd-length run of backslashes immediately before `index` as escaping.
    backslashes = 0
    j = index - 1
    while j >= 0 and markdown_text[j] == "\\":
        backslashes += 1
        j -= 1
    return (backslashes % 2) == 1


def _is_probably_inline_math(content: str) -> bool:
    """Heuristic to avoid treating currency/normal prose as math."""

    stripped = content.strip()
    if not stripped:
        return False
    if "\n" in stripped or "\r" in stripped:
        return False

    # Common math cues
    if any(
        token in stripped
        for token in (
            "\\",
            "_",
            "^",
            "{",
            "}",
            "=",
            "\\cdot",
            "\\times",
            "\\approx",
            "\\geq",
            "\\leq",
        )
    ):
        return True

    # Unit/math symbols frequently wrapped in `$...$`.
    if re.search(r"[°±≤≥]", stripped):
        return True

    # Chemical-ish / variable-ish: contains at least one letter AND one digit.
    if re.search(r"[A-Za-z]", stripped) and re.search(r"\d", stripped):
        return True

    # Pure numbers (e.g. $100$) are usually not math in prose.
    if re.fullmatch(r"[\d,]+(\.[\d]+)?", stripped):
        return False

    # Single-letter variables are common in math (e.g. $T$), but can also be normal text.
    # Keep them only if they look like a variable name.
    if re.fullmatch(r"[A-Za-z](?:_[A-Za-z0-9]+)?", stripped):
        return True

    return False


def _extract_display_math_blocks(markdown_text: str) -> list[MathBlock]:
    """Extract `$$ ... $$` blocks while ignoring fenced code blocks."""

    blocks: list[MathBlock] = []
    in_fence = False
    fence_marker: str | None = None

    i = 0
    n = len(markdown_text)

    while i < n:
        line_start = markdown_text.rfind("\n", 0, i) + 1
        line_end = markdown_text.find("\n", i)
        if line_end == -1:
            line_end = n
        line = markdown_text[line_start:line_end]

        fence_match = _FENCED_CODE_BLOCK_RE.match(line)
        if fence_match:
            marker = fence_match.group(1)
            if not in_fence:
                in_fence = True
                fence_marker = marker
            elif fence_marker == marker:
                in_fence = False
                fence_marker = None
            i = line_end + 1
            continue

        if in_fence:
            i = line_end + 1
            continue

        if markdown_text.startswith("$$", i):
            start = i
            end = markdown_text.find("$$", i + 2)
            if end == -1:
                break
            end += 2
            raw = markdown_text[start:end]
            content = raw[2:-2].strip()
            blocks.append(MathBlock(raw=raw, content=content, start=start, end=end))
            i = end
            continue

        i += 1

    return blocks


def _extract_inline_math_blocks(markdown_text: str) -> list[MathBlock]:
    """Extract inline `$...$` blocks while ignoring fenced and inline code.

    Notes:
    - Skips display math (`$$...$$`).
    - Skips escaped dollars (`\\$`).
    - Uses a heuristic to avoid treating currency like `$100` as math.
    """

    blocks: list[MathBlock] = []
    in_fence = False
    fence_marker: str | None = None
    in_inline_code = False
    inline_tick_count = 0

    i = 0
    n = len(markdown_text)

    while i < n:
        # Line-level fence detection (reuse the same logic as display extraction).
        line_start = markdown_text.rfind("\n", 0, i) + 1
        line_end = markdown_text.find("\n", i)
        if line_end == -1:
            line_end = n
        line = markdown_text[line_start:line_end]

        fence_match = _FENCED_CODE_BLOCK_RE.match(line)
        if fence_match and not in_inline_code:
            marker = fence_match.group(1)
            if not in_fence:
                in_fence = True
                fence_marker = marker
            elif fence_marker == marker:
                in_fence = False
                fence_marker = None
            i = line_end + 1
            continue

        if in_fence:
            i = line_end + 1
            continue

        ch = markdown_text[i]

        # Inline code spans: toggle on runs of backticks.
        if ch == "`":
            run = 1
            j = i + 1
            while j < n and markdown_text[j] == "`":
                run += 1
                j += 1
            if not in_inline_code:
                in_inline_code = True
                inline_tick_count = run
            elif run == inline_tick_count:
                in_inline_code = False
                inline_tick_count = 0
            i = j
            continue

        if in_inline_code:
            i += 1
            continue

        if ch == "$":
            if _is_escaped(markdown_text, i):
                i += 1
                continue

            # Skip display math.
            if markdown_text.startswith("$$", i):
                i += 2
                continue

            start = i
            # Only treat `$...$` as inline math if it closes on the same line.
            # This avoids swallowing currency like "$5/kg" and table headers like "$\/kWh".
            j = i + 1
            while j < line_end:
                if markdown_text[j] == "$" and not _is_escaped(markdown_text, j):
                    # Ignore $$ here; that's a display marker, not an inline closer.
                    if markdown_text.startswith("$$", j):
                        j += 2
                        continue
                    end = j + 1
                    raw = markdown_text[start:end]
                    content = raw[1:-1]
                    if _is_probably_inline_math(content):
                        blocks.append(MathBlock(raw=raw, content=content.strip(), start=start, end=end))
                    i = end
                    break
                j += 1
            else:
                # No closing `$` on this line: treat as a literal dollar sign.
                i += 1
                continue
            continue

        i += 1

    return blocks


def _stable_equation_id(content: str) -> str:
    normalized = " ".join(content.split())
    return hashlib.sha256(normalized.encode("utf-8")).hexdigest()[:12]


def _wrap_for_mathtext(content: str) -> str:
    return f"${content}$"


def _sanitize_for_mathtext(content: str) -> str:
    """Best-effort sanitization for Matplotlib mathtext.

    Matplotlib supports a subset of LaTeX. This function rewrites a few common
    macros we use in docs into mathtext-friendly equivalents.
    """

    sanitized = content

    # mathtext supports \mathbf{...} but not \textbf{...}
    sanitized = re.sub(r"\\textbf\{([^}]*)\}", r"\\mathbf{\1}", sanitized)

    # Common degree macro used in some markdown renderers.
    sanitized = sanitized.replace(r"\degree", r"^\circ")

    # Mathtext does not support AMS extensible arrows like \xrightarrow{...}.
    # Drop the label and keep the arrow direction.
    _ONE_LEVEL_BRACES = r"\{(?:[^{}]|\{[^{}]*\})*\}"
    sanitized = re.sub(
        rf"\\xrightarrow(\[[^\]]*\])?{_ONE_LEVEL_BRACES}",
        r"\\rightarrow",
        sanitized,
    )
    sanitized = re.sub(
        rf"\\xleftarrow(\[[^\]]*\])?{_ONE_LEVEL_BRACES}",
        r"\\leftarrow",
        sanitized,
    )

    return sanitized


def _make_backup(path: Path) -> Path:
    timestamp = time.strftime("%Y%m%d-%H%M%S")
    backup_path = path.with_name(f"{path.name}.bak-{timestamp}")
    backup_path.write_text(path.read_text(encoding="utf-8"), encoding="utf-8")
    return backup_path


def _inline_img_tag(rel_path: str, alt_text: str = "equation") -> str:
    # Use HTML so images stay inline.
    # height/vertical-align keeps the SVG aligned with surrounding text in most renderers.
    safe_alt = alt_text.replace('"', "'")
    return (
        f"<img src=\"{rel_path}\" alt=\"{safe_alt}\" "
        f"style=\"height:1em; vertical-align:middle;\" />"
    )


def _escape_currency_dollars(markdown_text: str) -> str:
    """Escape literal currency dollars (e.g. $250,000 or $5/kg) as `\$`.

    This is intended to prevent KaTeX/MathJax-style parsers from treating currency
    as math. Best-effort: skips fenced code blocks and inline code spans.
    """

    out_chars: list[str] = []
    in_fence = False
    fence_marker: str | None = None
    in_inline_code = False
    inline_tick_count = 0

    i = 0
    n = len(markdown_text)

    while i < n:
        line_start = markdown_text.rfind("\n", 0, i) + 1
        line_end = markdown_text.find("\n", i)
        if line_end == -1:
            line_end = n
        line = markdown_text[line_start:line_end]

        fence_match = _FENCED_CODE_BLOCK_RE.match(line)
        if fence_match and not in_inline_code:
            marker = fence_match.group(1)
            if not in_fence:
                in_fence = True
                fence_marker = marker
            elif fence_marker == marker:
                in_fence = False
                fence_marker = None

            # Copy the whole line as-is.
            out_chars.append(markdown_text[i:line_end])
            if line_end < n:
                out_chars.append("\n")
            i = line_end + 1
            continue

        if in_fence:
            out_chars.append(markdown_text[i:line_end])
            if line_end < n:
                out_chars.append("\n")
            i = line_end + 1
            continue

        ch = markdown_text[i]

        # Inline code spans: toggle on runs of backticks.
        if ch == "`":
            run = 1
            j = i + 1
            while j < n and markdown_text[j] == "`":
                run += 1
                j += 1
            if not in_inline_code:
                in_inline_code = True
                inline_tick_count = run
            elif run == inline_tick_count:
                in_inline_code = False
                inline_tick_count = 0
            out_chars.append(markdown_text[i:j])
            i = j
            continue

        if in_inline_code:
            out_chars.append(ch)
            i += 1
            continue

        if ch == "$" and not _is_escaped(markdown_text, i):
            # Treat `$` as currency if it's followed by optional whitespace then a digit.
            j = i + 1
            while j < n and markdown_text[j] in (" ", "\t"):
                j += 1
            if j < n and markdown_text[j].isdigit():
                out_chars.append("\\$")
                i += 1
                continue

        out_chars.append(ch)
        i += 1

    return "".join(out_chars)

def generate_svg(equation_latex: str, output_path: Path, fontsize: int = 22) -> None:
    """Generate a tightly-cropped transparent SVG from a math expression."""

    fig = plt.figure(figsize=(0.01, 0.01))
    fig.patch.set_alpha(0.0)
    ax = fig.add_axes([0, 0, 1, 1])
    ax.set_axis_off()

    # Render equation. Anchor to the lower-left to make bbox math simpler.
    text_artist = ax.text(
        0.0,
        0.0,
        equation_latex,
        ha="left",
        va="bottom",
        fontsize=fontsize,
        transform=ax.transAxes,
    )

    # Force a draw so we can get an accurate bounding box.
    fig.canvas.draw()
    renderer = fig.canvas.get_renderer()
    bbox_px = text_artist.get_window_extent(renderer=renderer).expanded(1.03, 1.25)

    # Convert bbox from pixels to inches for bbox_inches.
    bbox_in = bbox_px.transformed(fig.dpi_scale_trans.inverted())

    fig.savefig(
        output_path,
        format="svg",
        bbox_inches=bbox_in,
        pad_inches=0.0,
        transparent=True,
    )
    plt.close(fig)


def _render_equations_from_markdown(
    markdown_path: Path,
    out_dir: Path,
    fontsize: int,
    out_dir_inline: Path,
    fontsize_inline: int,
    inplace: bool,
    backup: bool,
    escape_currency: bool,
    validate: bool,
    validate_only: bool,
    max_failures: int,
) -> None:
    markdown_text = markdown_path.read_text(encoding="utf-8")
    display_blocks = _extract_display_math_blocks(markdown_text)
    inline_blocks = _extract_inline_math_blocks(markdown_text)
    if not display_blocks and not inline_blocks:
        if inplace and escape_currency:
            if backup:
                backup_path = _make_backup(markdown_path)
                print(f"Backup created: {backup_path}")
            rewritten = _escape_currency_dollars(markdown_text)
            markdown_path.write_text(rewritten, encoding="utf-8")
            print(f"Rewrote Markdown in-place: {markdown_path}")
            return

        print("No math blocks found.")
        return

    out_dir.mkdir(parents=True, exist_ok=True)
    out_dir_inline.mkdir(parents=True, exist_ok=True)

    def _dedupe(blocks: list[MathBlock]) -> list[str]:
        seen: set[str] = set()
        ordered: list[str] = []
        for block in blocks:
            eq_id = _stable_equation_id(block.content)
            if eq_id not in seen:
                ordered.append(eq_id)
                seen.add(eq_id)
        return ordered

    display_ids = _dedupe(display_blocks)
    inline_ids = _dedupe(inline_blocks)

    display_filenames: dict[str, str] = {
        eq_id: f"eq_{i+1:03d}_{eq_id}.svg" for i, eq_id in enumerate(display_ids)
    }
    inline_filenames: dict[str, str] = {
        eq_id: f"inl_{i+1:03d}_{eq_id}.svg" for i, eq_id in enumerate(inline_ids)
    }

    failures: list[tuple[str, str, str]] = []
    display_successes: set[str] = set()
    inline_successes: set[str] = set()

    def _first_content(eq_id: str, blocks: list[MathBlock]) -> str:
        return next(b.content for b in blocks if _stable_equation_id(b.content) == eq_id)

    # Validation pass (attempt render to temp SVGs and delete).
    if validate or validate_only:
        tmp_dir = out_dir / ".validate_tmp"
        tmp_dir.mkdir(parents=True, exist_ok=True)

        ok_ids: set[str] = set()
        fail_ids: set[str] = set()
        local_failures: list[tuple[str, str, str]] = []

        for eq_id in display_ids:
            content = _first_content(eq_id, display_blocks)
            sanitized = _sanitize_for_mathtext(content)
            latex = _wrap_for_mathtext(sanitized)
            try:
                generate_svg(latex, tmp_dir / f"tmp_display_{eq_id}.svg", fontsize=fontsize)
                ok_ids.add(eq_id)
            except Exception as e:
                fail_ids.add(eq_id)
                local_failures.append((eq_id, f"{type(e).__name__}: {e}", content[:140].replace("\n", " ")))

        for eq_id in inline_ids:
            content = _first_content(eq_id, inline_blocks)
            sanitized = _sanitize_for_mathtext(content)
            latex = _wrap_for_mathtext(sanitized)
            try:
                generate_svg(latex, tmp_dir / f"tmp_inline_{eq_id}.svg", fontsize=fontsize_inline)
                ok_ids.add(eq_id)
            except Exception as e:
                fail_ids.add(eq_id)
                local_failures.append((eq_id, f"{type(e).__name__}: {e}", content[:140].replace("\n", " ")))

        # Cleanup temp files.
        for p in tmp_dir.glob("*.svg"):
            p.unlink(missing_ok=True)
        tmp_dir.rmdir()

        print("Validation:")
        print("ok", len(ok_ids), "fail", len(fail_ids))
        for eq_id, err, snippet in local_failures[:max_failures]:
            print("FAIL", eq_id, err)
            print("  eq:", snippet)

        if validate_only:
            return

    # Render display equations.
    for eq_id in display_ids:
        content = _first_content(eq_id, display_blocks)
        sanitized = _sanitize_for_mathtext(content)
        latex = _wrap_for_mathtext(sanitized)
        try:
            generate_svg(latex, out_dir / display_filenames[eq_id], fontsize=fontsize)
            display_successes.add(eq_id)
        except Exception as e:
            failures.append((eq_id, f"{type(e).__name__}: {e}", content[:140].replace("\n", " ")))

    # Render inline equations.
    for eq_id in inline_ids:
        content = _first_content(eq_id, inline_blocks)
        sanitized = _sanitize_for_mathtext(content)
        latex = _wrap_for_mathtext(sanitized)
        try:
            generate_svg(latex, out_dir_inline / inline_filenames[eq_id], fontsize=fontsize_inline)
            inline_successes.add(eq_id)
        except Exception as e:
            failures.append((eq_id, f"{type(e).__name__}: {e}", content[:140].replace("\n", " ")))

    if display_ids:
        print(f"Generated {len(display_successes)} display SVG(s) in: {out_dir}")
    if inline_ids:
        print(f"Generated {len(inline_successes)} inline SVG(s) in: {out_dir_inline}")
    if failures:
        print(f"Skipped {len(failures)} equation(s) due to render errors.")

    if not inplace:
        return

    if backup:
        backup_path = _make_backup(markdown_path)
        print(f"Backup created: {backup_path}")

    rewritten = markdown_text

    # Replace from end -> start so indices remain valid.
    all_blocks = sorted(display_blocks + inline_blocks, key=lambda b: b.start, reverse=True)
    for block in all_blocks:
        eq_id = _stable_equation_id(block.content)

        # Display math replacement
        if block.raw.startswith("$$"):
            if eq_id not in display_successes:
                continue
            rel_path = f"{out_dir.name}/{display_filenames[eq_id]}"
            replacement = f"\n\n![equation]({rel_path})\n\n"
            rewritten = rewritten[: block.start] + replacement + rewritten[block.end :]
            continue

        # Inline math replacement
        if eq_id not in inline_successes:
            continue
        rel_path = f"{out_dir_inline.name}/{inline_filenames[eq_id]}"
        replacement = _inline_img_tag(rel_path)
        rewritten = rewritten[: block.start] + replacement + rewritten[block.end :]

    markdown_path.write_text(rewritten, encoding="utf-8")
    print(f"Rewrote Markdown in-place: {markdown_path}")

    if escape_currency:
        escaped = _escape_currency_dollars(markdown_path.read_text(encoding="utf-8"))
        markdown_path.write_text(escaped, encoding="utf-8")


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Render display math blocks (`$$...$$`) from Markdown to SVG.",
    )
    parser.add_argument(
        "markdown",
        type=Path,
        help="Path to a Markdown file containing display equations delimited by `$$...$$`.",
    )
    parser.add_argument(
        "--outdir",
        type=Path,
        default=None,
        help="Output directory for SVGs (default: <md_dir>/equations).",
    )
    parser.add_argument(
        "--fontsize",
        type=int,
        default=22,
        help="Font size for rendered equations.",
    )
    parser.add_argument(
        "--fontsize-inline",
        type=int,
        default=16,
        help="Font size for rendered inline equations.",
    )
    parser.add_argument(
        "--inplace",
        action="store_true",
        help="Rewrite the Markdown file to replace `$$...$$` blocks with SVG image links.",
    )
    parser.add_argument(
        "--no-backup",
        action="store_true",
        help="Disable automatic .bak-* backup creation when using --inplace.",
    )
    parser.add_argument(
        "--escape-currency",
        action="store_true",
        help="Escape currency dollars like $250,000 or $5/kg into \\$ to avoid KaTeX-style parsing.",
    )
    parser.add_argument(
        "--outdir-inline",
        type=Path,
        default=None,
        help="Output directory for inline SVGs (default: <outdir>_inline).",
    )
    parser.add_argument(
        "--validate",
        action="store_true",
        help="Validate renderability (prints ok/fail counts + snippets) before rendering and rewriting.",
    )
    parser.add_argument(
        "--validate-only",
        action="store_true",
        help="Only validate renderability; do not generate final SVGs or rewrite files.",
    )
    parser.add_argument(
        "--max-failures",
        type=int,
        default=10,
        help="Max failure details to print during validation.",
    )
    return parser.parse_args()

def main():
    args = _parse_args()
    markdown_path: Path = args.markdown
    if not markdown_path.exists():
        raise SystemExit(f"Markdown file not found: {markdown_path}")

    out_dir = args.outdir
    if out_dir is None:
        out_dir = markdown_path.parent / "equations"

    out_dir_inline = args.outdir_inline
    if out_dir_inline is None:
        out_dir_inline = out_dir.with_name(f"{out_dir.name}_inline")

    _render_equations_from_markdown(
        markdown_path=markdown_path,
        out_dir=out_dir,
        fontsize=args.fontsize,
        out_dir_inline=out_dir_inline,
        fontsize_inline=args.fontsize_inline,
        inplace=args.inplace,
        backup=not args.no_backup,
        escape_currency=args.escape_currency,
        validate=args.validate,
        validate_only=args.validate_only,
        max_failures=args.max_failures,
    )

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Generate PNG images from display equations embedded in a Markdown document.

Modified from generate_equation_svgs.py to output PNG format suitable for watermarking.
"""

from __future__ import annotations
import argparse
import hashlib
import re
import time
from dataclasses import dataclass
from pathlib import Path

import matplotlib
import matplotlib.pyplot as plt

# Use non-interactive backend
matplotlib.use("Agg")


@dataclass(frozen=True)
class MathBlock:
    raw: str
    content: str
    start: int
    end: int


_FENCED_CODE_BLOCK_RE = re.compile(r"^\s*(```|~~~)")


def _is_escaped(markdown_text: str, index: int) -> bool:
    backslashes = 0
    j = index - 1
    while j >= 0 and markdown_text[j] == "\\":
        backslashes += 1
        j -= 1
    return (backslashes % 2) == 1


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
            elif marker == fence_marker:
                in_fence = False
                fence_marker = None
            i = line_end + 1
            continue

        if in_fence:
            i = line_end + 1
            continue

        # Look for $$ (display math)
        if i + 1 < n and markdown_text[i:i+2] == "$$" and not _is_escaped(markdown_text, i):
            start = i
            i += 2
            end_index = markdown_text.find("$$", i)
            while end_index != -1 and _is_escaped(markdown_text, end_index):
                end_index = markdown_text.find("$$", end_index + 2)
            if end_index == -1:
                break

            raw = markdown_text[start:end_index+2]
            content = markdown_text[i:end_index].strip()
            blocks.append(MathBlock(raw=raw, content=content, start=start, end=end_index+2))
            i = end_index + 2
        else:
            i += 1

    return blocks


def _stable_equation_id(content: str) -> str:
    normalized = " ".join(content.split())
    return hashlib.sha256(normalized.encode("utf-8")).hexdigest()[:12]


def _wrap_for_mathtext(content: str) -> str:
    return f"${content}$"


def _sanitize_for_mathtext(content: str) -> str:
    """Best-effort sanitization for Matplotlib mathtext."""
    sanitized = content
    sanitized = re.sub(r"\\textbf\{([^}]*)\}", r"\\mathbf{\1}", sanitized)
    sanitized = sanitized.replace(r"\degree", r"^\circ")

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


def generate_png(equation_latex: str, output_path: Path, fontsize: int = 22, dpi: int = 150) -> None:
    """Generate a tightly-cropped PNG from a math expression."""
    fig = plt.figure(figsize=(0.01, 0.01))
    fig.patch.set_facecolor('white')
    ax = fig.add_axes([0, 0, 1, 1])
    ax.set_axis_off()
    ax.patch.set_facecolor('white')

    # Render equation
    text_artist = ax.text(
        0.0,
        0.0,
        equation_latex,
        ha="left",
        va="bottom",
        fontsize=fontsize,
        transform=ax.transAxes,
    )

    # Force draw to get accurate bounding box
    fig.canvas.draw()
    renderer = fig.canvas.get_renderer()
    bbox_px = text_artist.get_window_extent(renderer=renderer).expanded(1.05, 1.3)

    # Convert to inches
    bbox_in = bbox_px.transformed(fig.dpi_scale_trans.inverted())

    fig.savefig(
        output_path,
        format="png",
        bbox_inches=bbox_in,
        pad_inches=0.05,
        dpi=dpi,
        facecolor='white',
    )
    plt.close(fig)


def _render_equations_from_markdown(
    markdown_path: Path,
    out_dir: Path,
    fontsize: int,
    dpi: int,
    inplace: bool,
) -> None:
    markdown_text = markdown_path.read_text(encoding="utf-8")
    display_blocks = _extract_display_math_blocks(markdown_text)

    if not display_blocks:
        print(f"No display math blocks found in {markdown_path}")
        return

    out_dir.mkdir(parents=True, exist_ok=True)

    # Deduplicate equations
    seen_ids: set[str] = set()
    unique_blocks: list[tuple[str, MathBlock]] = []
    for block in display_blocks:
        eq_id = _stable_equation_id(block.content)
        if eq_id not in seen_ids:
            seen_ids.add(eq_id)
            unique_blocks.append((eq_id, block))

    print(f"Found {len(display_blocks)} display equations ({len(unique_blocks)} unique)")

    # Generate PNGs
    filenames: dict[str, str] = {}
    for i, (eq_id, block) in enumerate(unique_blocks, 1):
        filename = f"eq_{i:03d}_{eq_id}.png"
        filenames[eq_id] = filename
        output_path = out_dir / filename

        try:
            sanitized = _sanitize_for_mathtext(block.content)
            wrapped = _wrap_for_mathtext(sanitized)
            generate_png(wrapped, output_path, fontsize=fontsize, dpi=dpi)
            print(f"  [{i}/{len(unique_blocks)}] {filename}")
        except Exception as e:
            print(f"  [FAILED] {filename}: {e}")
            continue

    print(f"\nGenerated {len(filenames)} PNG files in {out_dir}")

    if not inplace:
        return

    # Replace equations with image references
    rewritten = markdown_text
    for block in sorted(display_blocks, key=lambda b: b.start, reverse=True):
        eq_id = _stable_equation_id(block.content)
        if eq_id in filenames:
            rel_path = out_dir.name + "/" + filenames[eq_id]
            img_tag = f"![equation]({rel_path})"
            rewritten = rewritten[:block.start] + img_tag + rewritten[block.end:]

    # Backup original
    timestamp = time.strftime("%Y%m%d-%H%M%S")
    backup_path = markdown_path.with_name(f"{markdown_path.name}.bak-{timestamp}")
    backup_path.write_text(markdown_text, encoding="utf-8")
    print(f"\nBacked up original to: {backup_path}")

    # Write modified markdown
    markdown_path.write_text(rewritten, encoding="utf-8")
    print(f"Rewrote Markdown in-place: {markdown_path}")


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Render display math blocks (`$$...$$`) from Markdown to PNG.",
    )
    parser.add_argument(
        "markdown",
        type=Path,
        help="Path to Markdown file containing equations delimited by `$$...$$`.",
    )
    parser.add_argument(
        "--outdir",
        type=Path,
        default=None,
        help="Output directory for PNGs (default: <md_dir>/equations).",
    )
    parser.add_argument(
        "--fontsize",
        type=int,
        default=22,
        help="Font size for rendered equations (default: 22).",
    )
    parser.add_argument(
        "--dpi",
        type=int,
        default=150,
        help="DPI for PNG output (default: 150).",
    )
    parser.add_argument(
        "--inplace",
        action="store_true",
        help="Rewrite Markdown file to replace `$$...$$` with PNG image links.",
    )
    return parser.parse_args()


def main():
    args = _parse_args()
    markdown_path: Path = args.markdown

    if not markdown_path.exists():
        print(f"Error: {markdown_path} not found")
        raise SystemExit(1)

    out_dir = args.outdir
    if out_dir is None:
        out_dir = markdown_path.parent / "equations"

    _render_equations_from_markdown(
        markdown_path=markdown_path,
        out_dir=out_dir,
        fontsize=args.fontsize,
        dpi=args.dpi,
        inplace=args.inplace,
    )


if __name__ == "__main__":
    main()

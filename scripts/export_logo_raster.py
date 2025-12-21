#!/usr/bin/env python3

"""
Export rasterized logo assets (PNG squares, PDF at specified DPI, and raster EPS)
from a source image without any hardcoded paths.

Usage examples:

    1) Basic export (default sizes 256,512,1024,2048 and PDF at 300 DPI):

         python3 scripts/export_logo_raster.py \
             --src /path/to/logo.png \
             --out /path/to/exports

    2) Custom base name, sizes, and PDF DPI:

         python3 scripts/export_logo_raster.py \
             --src ./logo.png \
             --out ./out \
             --name brandmark \
             --sizes 128,256,512 \
             --pdf-dpi 300

Notes:
- No hardcoded paths; you must provide --src and --out.
- Requires Pillow: pip install pillow
- PNG exports preserve aspect ratio and are padded to square with transparent background.
- EPS export is raster-only (RGB, no alpha channel).
"""

import argparse
from PIL import Image
from pathlib import Path
from typing import List


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Export rasterized logo assets at multiple sizes, plus PDF (300 DPI) and EPS."
    )
    parser.add_argument(
        "--src",
        required=True,
        type=Path,
        help="Path to source logo image (PNG recommended).",
    )
    parser.add_argument(
        "--out",
        required=True,
        type=Path,
        help="Output directory to write exported assets.",
    )
    parser.add_argument(
        "--name",
        required=False,
        help="Base name to use for output files (default: derived from source filename).",
    )
    parser.add_argument(
        "--sizes",
        required=False,
        default="256,512,1024,2048",
        help="Comma-separated square sizes for PNG exports (default: 256,512,1024,2048).",
    )
    parser.add_argument(
        "--pdf-dpi",
        type=int,
        default=300,
        help="DPI for PDF export (default: 300).",
    )
    return parser.parse_args()


def parse_sizes(s: str) -> List[int]:
    vals = []
    for part in s.split(","):
        part = part.strip()
        if not part:
            continue
        try:
            vals.append(int(part))
        except ValueError:
            raise SystemExit(f"Invalid size value: {part}")
    if not vals:
        raise SystemExit("No valid sizes provided.")
    return vals


def main() -> int:
    args = parse_args()

    src_png: Path = args.src
    out_dir: Path = args.out
    out_dir.mkdir(parents=True, exist_ok=True)

    if not src_png.exists():
        print("Source image not found:", src_png)
        return 1

    base_name = args.name if args.name else src_png.stem
    sizes = parse_sizes(args.sizes)

    # Load source image (prefer RGBA for transparent canvas handling)
    img = Image.open(src_png).convert("RGBA")

    # Generate square PNGs with transparent padding
    for size in sizes:
        w, h = img.size
        if w >= h:
            new_w = size
            new_h = int(size * h / w)
        else:
            new_h = size
            new_w = int(size * w / h)

        resized = img.resize((new_w, new_h), Image.LANCZOS)
        out_png = out_dir / f"{base_name}_{size}x{size}.png"

        # create square canvas with transparent background
        canvas = Image.new("RGBA", (size, size), (255, 255, 255, 0))
        canvas.paste(resized, ((size - new_w) // 2, (size - new_h) // 2), resized)
        canvas.save(out_png)
        print("Saved", out_png)

    # High-resolution PDF export (embed raster at specified DPI)
    pdf_img = img.copy()
    pdf_rgb = pdf_img.convert("RGB")
    pdf_out = out_dir / f"{base_name}_print_{args.pdf_dpi}dpi.pdf"
    pdf_rgb.save(pdf_out, format="PDF", resolution=args.pdf_dpi)
    print("Saved", pdf_out)

    # Raster EPS export (note: not a vector EPS)
    eps_out = out_dir / f"{base_name}_raster.eps"
    eps_rgb = img.convert("RGB")
    eps_rgb.save(eps_out, format="EPS")
    print("Saved", eps_out)

    print("All exports complete in", out_dir)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

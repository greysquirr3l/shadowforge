# !/usr/bin/env python3

from PIL import Image
from pathlib import Path

# Configuration

ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / 'strangecarbon.co' / 'assets'
SRC_PNG = ASSETS / 'strange_carbon.png'
OUT_DIR = ASSETS / 'exports'
OUT_DIR.mkdir(exist_ok=True)

sizes = [256, 512, 1024, 2048]

if not SRC_PNG.exists():
    print('Source PNG not found:', SRC_PNG)
    raise SystemExit(1)

img = Image.open(SRC_PNG).convert('RGBA')

for size in sizes:
    # preserve aspect ratio
    w, h = img.size
    if w >= h:
        new_w = size
        new_h = int(size *h / w)
    else:
        new_h = size
        new_w = int(size* w / h)

    resized = img.resize((new_w, new_h), Image.LANCZOS)
    out_png = OUT_DIR / f'strange_carbon_{size}x{size}.png'
    # create square canvas with transparent background
    canvas = Image.new('RGBA', (size, size), (255,255,255,0))
    canvas.paste(resized, ((size-new_w)//2, (size-new_h)//2), resized)
    canvas.save(out_png)
    print('Saved', out_png)

# Create a high-resolution PDF (embed a 300 DPI raster)

pdf_img = img.copy()

# convert to RGB for PDF

pdf_rgb = pdf_img.convert('RGB')
pdf_out = OUT_DIR / 'strange_carbon_print_300dpi.pdf'

# Save at 300 DPI by specifying quality via Pillow

pdf_rgb.save(pdf_out, format='PDF', resolution=300)
print('Saved', pdf_out)

# Raster EPS export (note: not a vector EPS)

eps_out = OUT_DIR / 'strange_carbon_raster.eps'

# EPS requires RGB and no alpha

eps_rgb = img.convert('RGB')
eps_rgb.save(eps_out, format='EPS')
print('Saved', eps_out)

print('All exports complete in', OUT_DIR)

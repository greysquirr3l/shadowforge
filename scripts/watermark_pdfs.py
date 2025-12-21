#!/usr/bin/env python3

from __future__ import annotations

import argparse
import hashlib
import json
import os
import platform
import shutil
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from pathlib import Path

import fitz  # PyMuPDF


@dataclass(frozen=True)
class WatermarkOptions:
    recipient: str
    gpg_key_path: Path
    shadowforge_bin: Path
    technique: str
    redundancy: float
    strategy: str
    document_id: str | None
    no_encrypt: bool
    no_sign: bool
    receipt_dir: Path | None


def _run(cmd: list[str]) -> None:
    result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    if result.returncode != 0:
        raise RuntimeError(f"command failed (exit={result.returncode}): {' '.join(cmd)}\n{result.stdout}")


def _run_output(cmd: list[str]) -> str:
    result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    if result.returncode != 0:
        raise RuntimeError(f"command failed (exit={result.returncode}): {' '.join(cmd)}\n{result.stdout}")
    return result.stdout


def _sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        while True:
            chunk = f.read(1024 * 1024)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()


def _try_shadowforge_version(shadowforge_bin: Path) -> str | None:
    for args in (["version"], ["--version"]):
        try:
            out = _run_output([str(shadowforge_bin), *args]).strip()
        except Exception:
            continue
        if out:
            return out
    return None


def _upsert_markdown_json_section(markdown_path: Path, payload: dict) -> None:
    begin = "<!-- SHADOWFORGE_JSON_RECEIPT_BEGIN -->"
    end = "<!-- SHADOWFORGE_JSON_RECEIPT_END -->"
    section = (
        "\n\n## JSON Receipt\n\n"
        f"{begin}\n\n"
        "```json\n"
        f"{json.dumps(payload, indent=2, sort_keys=True)}\n"
        "```\n\n"
        f"{end}\n"
    )

    if markdown_path.exists():
        existing = markdown_path.read_text(encoding="utf-8")
        start = existing.find(begin)
        stop = existing.find(end)
        if start != -1 and stop != -1 and stop > start:
            stop = stop + len(end)
            updated = existing[:start] + section.strip("\n") + existing[stop:]
            markdown_path.write_text(updated, encoding="utf-8")
            return

        markdown_path.write_text(existing + section, encoding="utf-8")
        return

    markdown_path.parent.mkdir(parents=True, exist_ok=True)
    markdown_path.write_text(
        "# Shadowforge Watermark Receipt\n\n"
        "This receipt records watermarking inputs/outputs and parameters.\n"
        + section.lstrip("\n"),
        encoding="utf-8",
    )


def _safe_receipt_stem(doc_id: str) -> str:
    stem = doc_id.strip().replace("/", "_").replace("\\", "_")
    if not stem:
        return "document"
    return stem


def _rasterize_pdf_to_pngs(pdf_path: Path, out_dir: Path, dpi: int) -> list[Path]:
    out_dir.mkdir(parents=True, exist_ok=True)

    doc = fitz.open(str(pdf_path))
    scale = dpi / 72.0
    matrix = fitz.Matrix(scale, scale)

    page_paths: list[Path] = []
    for page_index in range(doc.page_count):
        page = doc.load_page(page_index)
        pix = page.get_pixmap(matrix=matrix, alpha=False)
        out_path = out_dir / f"page_{page_index + 1:04d}.png"
        pix.save(str(out_path))
        page_paths.append(out_path)

    doc.close()
    return page_paths


def _build_pdf_from_pngs(png_paths: list[Path], out_pdf: Path, dpi: int) -> None:
    if not png_paths:
        raise ValueError("no PNG pages to build PDF")

    out_pdf.parent.mkdir(parents=True, exist_ok=True)

    pdf = fitz.open()
    try:
        for png_path in png_paths:
            pix = fitz.Pixmap(str(png_path))

            width_pt = pix.width * 72.0 / dpi
            height_pt = pix.height * 72.0 / dpi

            page = pdf.new_page(width=width_pt, height=height_pt)
            page.insert_image(page.rect, filename=str(png_path))

        pdf.save(str(out_pdf))
    finally:
        pdf.close()


def _watermark_png_dir(input_dir: Path, output_dir: Path, opts: WatermarkOptions) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)

    cmd = [
        str(opts.shadowforge_bin),
        "watermark",
        "embed",
        "--recipient",
        opts.recipient,
        "--gpg-key",
        str(opts.gpg_key_path),
        "--input-dir",
        str(input_dir),
        "--output-dir",
        str(output_dir),
        "--technique",
        opts.technique,
        "--redundancy",
        str(opts.redundancy),
        "--strategy",
        opts.strategy,
    ]

    if opts.document_id:
        cmd.extend(["--document-id", opts.document_id])
    if opts.no_encrypt:
        cmd.append("--no-encrypt")
    if opts.no_sign:
        cmd.append("--no-sign")

    if not opts.no_encrypt and opts.receipt_dir:
        opts.receipt_dir.mkdir(parents=True, exist_ok=True)
        receipt_path = opts.receipt_dir / f"watermark_{_safe_receipt_stem(opts.document_id or 'document')}.md"
        cmd.extend(["--receipt", str(receipt_path)])

    _run(cmd)


def _extract_watermark_json(
    input_dir: Path,
    out_json: Path,
    shadowforge_bin: Path,
    threshold: int,
    no_decrypt: bool,
    receipt_path: Path | None,
) -> None:
    out_json.parent.mkdir(parents=True, exist_ok=True)

    if threshold <= 0:
        raise ValueError(f"invalid threshold: {threshold}")

    cmd = [
        str(shadowforge_bin),
        "watermark",
        "extract",
        "--input-dir",
        str(input_dir),
        "--output",
        str(out_json),
        "--threshold",
        str(threshold),
    ]

    if no_decrypt:
        cmd.append("--no-decrypt")

    if not no_decrypt and receipt_path:
        cmd.extend(["--receipt", str(receipt_path)])

    _run(cmd)


def _output_with_suffix(pdf_path: Path, suffix: str) -> Path:
    if pdf_path.suffix.lower() != ".pdf":
        raise ValueError(f"expected .pdf input, got: {pdf_path}")
    return pdf_path.with_name(f"{pdf_path.stem}{suffix}.pdf")


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Rasterize PDFs -> watermark pages -> rebuild PDFs (Shadowforge watermark pipeline)."
    )
    parser.add_argument("pdfs", nargs="+", type=Path, help="Input PDF path(s)")

    parser.add_argument(
        "--recipient",
        required=True,
        help='Recipient string, e.g. "Carrie Wheeler <carrie.wheeler@outlook.com>"',
    )
    parser.add_argument("--gpg-key", required=True, type=Path, help="ASCII-armored public key file")
    parser.add_argument(
        "--shadowforge",
        default=Path("./bin/shadowforge"),
        type=Path,
        help="Path to Shadowforge CLI binary",
    )

    parser.add_argument("--dpi", default=200, type=int, help="Rasterization DPI (default: 200)")
    parser.add_argument("--suffix", default="_cwheeler", help="Filename suffix for outputs")

    parser.add_argument("--technique", default="lsb", choices=["lsb", "dct"], help="Watermark technique")
    parser.add_argument("--redundancy", default=0.3, type=float, help="Reed-Solomon redundancy")
    parser.add_argument(
        "--strategy",
        default="distributed",
        choices=["distributed", "single", "redundant"],
        help="Embedding strategy",
    )

    parser.add_argument(
        "--document-id",
        default=None,
        help="Optional document-id to embed (defaults to input PDF filename)",
    )
    parser.add_argument("--no-encrypt", action="store_true", help="Disable encryption (not recommended)")
    parser.add_argument("--no-sign", action="store_true", help="Disable signing (not recommended)")

    parser.add_argument(
        "--receipt-dir",
        default=None,
        type=Path,
        help=(
            "Directory to write per-document watermark encryption receipts (Markdown). "
            "Required when embedding encrypted watermarks if you want to later decrypt/verify them."
        ),
    )

    args = parser.parse_args()

    opts = WatermarkOptions(
        recipient=args.recipient,
        gpg_key_path=args.gpg_key,
        shadowforge_bin=args.shadowforge,
        technique=args.technique,
        redundancy=args.redundancy,
        strategy=args.strategy,
        document_id=args.document_id,
        no_encrypt=args.no_encrypt,
        no_sign=args.no_sign,
        receipt_dir=args.receipt_dir,
    )

    if not opts.shadowforge_bin.exists():
        raise SystemExit(f"shadowforge binary not found: {opts.shadowforge_bin}")
    if not os.access(opts.shadowforge_bin, os.X_OK):
        raise SystemExit(f"shadowforge binary not executable: {opts.shadowforge_bin}")
    if not opts.gpg_key_path.exists():
        raise SystemExit(f"gpg key not found: {opts.gpg_key_path}")

    for pdf_path in args.pdfs:
        pdf_path = pdf_path.expanduser().resolve()
        if not pdf_path.exists():
            raise SystemExit(f"input PDF not found: {pdf_path}")

        out_pdf = _output_with_suffix(pdf_path, args.suffix)
        out_json = out_pdf.with_suffix(out_pdf.suffix + ".watermark.json")

        effective_doc_id = opts.document_id or pdf_path.name

        receipt_dir = None
        receipt_path = None
        if not opts.no_encrypt:
            receipt_dir = (opts.receipt_dir.expanduser().resolve() if opts.receipt_dir else out_pdf.parent)
            receipt_path = receipt_dir / f"watermark_{_safe_receipt_stem(effective_doc_id)}.md"

        per_doc_opts = WatermarkOptions(
            recipient=opts.recipient,
            gpg_key_path=opts.gpg_key_path,
            shadowforge_bin=opts.shadowforge_bin,
            technique=opts.technique,
            redundancy=opts.redundancy,
            strategy=opts.strategy,
            document_id=effective_doc_id,
            no_encrypt=opts.no_encrypt,
            no_sign=opts.no_sign,
            receipt_dir=receipt_dir,
        )

        json_receipt_path = out_pdf.with_suffix(out_pdf.suffix + ".receipt.json")
        md_receipt_path = receipt_path or out_pdf.with_suffix(out_pdf.suffix + ".receipt.md")

        with tempfile.TemporaryDirectory(prefix="shadowforge_pdfwm_") as tmp:
            tmp_dir = Path(tmp)
            pages_in = tmp_dir / "pages_in"
            pages_out = tmp_dir / "pages_out"

            # Rasterize
            _rasterize_pdf_to_pngs(pdf_path, pages_in, dpi=args.dpi)

            # Watermark
            _watermark_png_dir(pages_in, pages_out, per_doc_opts)

            # Rebuild PDF
            watermarked_pages = sorted(pages_out.glob("*.png"))
            if not watermarked_pages:
                raise RuntimeError(f"no watermarked PNGs produced in {pages_out}")

            # Create output in a temp file first, then atomically replace
            tmp_out_pdf = out_pdf.with_suffix(out_pdf.suffix + ".tmp")
            if tmp_out_pdf.exists():
                tmp_out_pdf.unlink()

            _build_pdf_from_pngs(watermarked_pages, tmp_out_pdf, dpi=args.dpi)
            shutil.move(str(tmp_out_pdf), str(out_pdf))

            # Extract (verification)
            _extract_watermark_json(
                pages_out,
                out_json,
                opts.shadowforge_bin,
                threshold=len(watermarked_pages),
                no_decrypt=per_doc_opts.no_encrypt,
                receipt_path=receipt_path,
            )

        # Produce JSON receipt + embed it into Markdown receipt
        shadowforge_version = _try_shadowforge_version(opts.shadowforge_bin)

        receipt = {
            "schema": "shadowforge.watermark.receipt.v1",
            "created_at": __import__("datetime").datetime.now(__import__("datetime").timezone.utc).isoformat(),
            "host": {
                "platform": platform.platform(),
                "python": sys.version.split()[0],
            },
            "tool": {
                "shadowforge_bin": str(opts.shadowforge_bin),
                "shadowforge_version": shadowforge_version,
                "script": "scripts/watermark_pdfs.py",
            },
            "inputs": {
                "pdf": {
                    "path": str(pdf_path),
                    "sha256": _sha256_file(pdf_path),
                }
            },
            "outputs": {
                "pdf": {
                    "path": str(out_pdf),
                    "sha256": _sha256_file(out_pdf),
                },
                "extracted_watermark_json": {
                    "path": str(out_json),
                    "sha256": _sha256_file(out_json),
                },
                "json_receipt": {
                    "path": str(json_receipt_path),
                },
                "markdown_receipt": {
                    "path": str(md_receipt_path),
                },
            },
            "watermark": {
                "recipient": per_doc_opts.recipient,
                "gpg_key_path": str(per_doc_opts.gpg_key_path),
                "document_id": per_doc_opts.document_id,
                "technique": per_doc_opts.technique,
                "redundancy": per_doc_opts.redundancy,
                "strategy": per_doc_opts.strategy,
                "encrypt": not per_doc_opts.no_encrypt,
                "sign": not per_doc_opts.no_sign,
                "dpi": args.dpi,
            },
        }

        json_receipt_path.parent.mkdir(parents=True, exist_ok=True)
        json_receipt_path.write_text(json.dumps(receipt, indent=2, sort_keys=True) + "\n", encoding="utf-8")

        _upsert_markdown_json_section(md_receipt_path, receipt)

        print(f"OK: {out_pdf}")
        print(f"OK: {out_json}")
        if receipt_path:
            print(f"OK: {receipt_path}")
        print(f"OK: {json_receipt_path}")
        print(f"OK: {md_receipt_path}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())

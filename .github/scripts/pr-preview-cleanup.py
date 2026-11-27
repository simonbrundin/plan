#!/usr/bin/env python3
"""Tar bort kustomize-overlays för en PR-preview.

Skriptet körs av GitHub Actions när en PR stängs. Miljövariabeln PR_NUM
kommer från workflow-steget.

När katalogen försvinner ur environments/pr/ prunar Flux bort resurserna
i klustret - inklusive preview-ns:et och dess databas.
"""

import os
import shutil
from pathlib import Path

PR_NUM = os.environ["PR_NUM"]
OUT_DIR = Path(f"environments/pr/{PR_NUM}")

if OUT_DIR.exists():
    shutil.rmtree(OUT_DIR)
    print(f"✓ Tog bort {OUT_DIR}")
else:
    print(f"- {OUT_DIR} fanns inte")

# README-filen ligger kvar så att katalogen aldrig blir tom. En tom
# katalog gör att Flux bygger "no resources" och ingenting prunas -
# previewen skulle bli kvar i klustret.
README = Path("environments/pr/README.md")
if not README.exists():
    README.write_text(
        "# PR-previews\n\n"
        "Overlayer per PR, genererade av .github/scripts/pr-preview-resources.py.\n",
        encoding="utf-8",
    )
    print(f"✓ Skapade {README}")

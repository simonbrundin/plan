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

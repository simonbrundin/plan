#!/usr/bin/env python3
"""Skapar kustomize-overlays för en PR-preview.

Skriptet körs av GitHub Actions. Miljövariablerna PR_NUM, NUXT_IMAGE och
API_IMAGE kommer från workflow-steget.

Overlayen hamnar i environments/pr/<pr-nummer>/ på branchen "previews".
Flux applicerar hela environments/pr-katalogen, så varje PR får en egen
katalog - och inget index behövs (två PR:er som rör samma fil skulle
annars slåss om filåtkomsten).
"""

import os
import shutil
from pathlib import Path

PR_NUM = os.environ["PR_NUM"]
NUXT_IMAGE = os.environ["NUXT_IMAGE"]
API_IMAGE = os.environ["API_IMAGE"]

TEMPLATE_DIR = Path("environments/pr-template")
OUT_DIR = Path(f"environments/pr/{PR_NUM}")

# Registry-uppgifterna för GHCR ligger krypterade i prod. Kustomize får
# inte referera filer ovanför katalogen, så filen kopieras in i overlayen
# i stället - krypterad, precis som i prod. Flux dekrypterar den.
GHCR_SECRET_SOURCE = Path("environments/prod/secrets/ghcr-secret.enc.yaml")
GHCR_SECRET_NAME = "ghcr-secret.enc.yaml"

# Schemat (ConfigMap) kopieras in från basen - samma schema som dev och
# prod. Overlayen får inte referera till basen, eftersom den publiceras
# på branchen "previews" som bara innehåller preview-overlays.
SCHEMA_SOURCE = Path("environments/base/db/schema.yaml")
SCHEMA_NAME = "schema-configmap.yaml"

NAMESPACE = f"plan-pr-{PR_NUM}"
HOSTNAME = f"plan-pr-{PR_NUM}.simonbrundin.com"

# Platshållarna i mallen byts ut här. Sökvägarna i mallens
# kustomization.yaml är relativa till mallens egen plats - i den
# genererade katalogen ligger de ett djupare, så de skrivs om.
REPLACEMENTS = {
    "NAMESPACE_PLACEHOLDER": NAMESPACE,
    "PR_NUMBER_PLACEHOLDER": PR_NUM,
    "HOSTNAME_PLACEHOLDER": HOSTNAME,
    "NUXT_IMAGE_PLACEHOLDER": NUXT_IMAGE,
    "API_IMAGE_PLACEHOLDER": API_IMAGE,
    "../base/": "../../base/",
    "../prod/": "../../prod/",
}

HEADER = (
    "# DO NOT EDIT MANUALLY\n"
    "# Genereras av .github/scripts/pr-preview-resources.py\n"
    f"# PR {PR_NUM} - https://{HOSTNAME}\n\n"
)


def render(text: str) -> str:
    """Byter ut platshållarna i en mallfil."""
    for placeholder, value in REPLACEMENTS.items():
        text = text.replace(placeholder, value)
    return text


def write_overlay() -> None:
    """Skriver overlayen för preview-miljön."""
    if not TEMPLATE_DIR.is_dir():
        raise SystemExit(f"FEL: mallen {TEMPLATE_DIR} saknas")

    if OUT_DIR.exists():
        shutil.rmtree(OUT_DIR)
    OUT_DIR.mkdir(parents=True)

    for source in sorted(TEMPLATE_DIR.glob("*.yaml")):
        target = OUT_DIR / source.name
        target.write_text(HEADER + render(source.read_text()), encoding="utf-8")
        print(f"✓ Skapade {target}")

    ghcr_target = OUT_DIR / GHCR_SECRET_NAME
    ghcr_target.write_text(GHCR_SECRET_SOURCE.read_text(encoding="utf-8"), encoding="utf-8")
    print(f"✓ Kopierade {GHCR_SECRET_SOURCE} -> {ghcr_target}")

    schema_target = OUT_DIR / SCHEMA_NAME
    schema_target.write_text(SCHEMA_SOURCE.read_text(encoding="utf-8"), encoding="utf-8")
    print(f"✓ Kopierade {SCHEMA_SOURCE} -> {schema_target}")


if __name__ == "__main__":
    write_overlay()
    print(f"Previewen deployas till namespace {NAMESPACE} med URL https://{HOSTNAME}")

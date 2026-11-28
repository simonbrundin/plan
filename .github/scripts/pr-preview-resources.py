#!/usr/bin/env python3
"""Skapar kustomize-overlays för en PR-preview.

Skriptet körs av GitHub Actions. Miljövariablerna PR_NUM (eller
PR_NUMBER), NUXT_IMAGE och
API_IMAGE kommer från workflow-steget.

Overlayen hamnar i environments/pr/<pr-nummer>/ på branchen "previews".
Flux applicerar hela environments/pr-katalogen, så varje PR får en egen
katalog - och inget index behövs (två PR:er som rör samma fil skulle
annars slåss om filåtkomsten).
"""

import os
import shutil
from pathlib import Path

# PR_NUM och PR_NUMBER är samma sak - olika steg i kedjan har använt
# olika namn. Acceptera båda, så att skriptet inte är känsligt för det.
PR_NUM = os.environ.get("PR_NUM") or os.environ["PR_NUMBER"]
NUXT_IMAGE = os.environ["NUXT_IMAGE"]
API_IMAGE = os.environ["API_IMAGE"]
AGENT_IMAGE = os.environ["AGENT_IMAGE"]

# Agentens jobb klonar repot som PR:en gäller, så att agenten ser
# samma kod som granskaren. Utan värden faller vi tillbaka på
# huvudrepoet.
REPO_URL = os.environ.get("PR_HEAD_REPO_URL") or "https://github.com/simonbrundin/plan.git"
REPO_BRANCH = os.environ.get("PR_HEAD_BRANCH") or "main"

TEMPLATE_DIR = Path("environments/pr-template")
OUT_DIR = Path(f"environments/pr/{PR_NUM}")

# Registry-uppgifterna för GHCR ligger krypterade i prod. Kustomize får
# inte referera filer ovanför katalogen, så filen kopieras in i overlayen
# i stället - krypterad, precis som i prod. Flux dekrypterar den.
GHCR_SECRET_SOURCE = Path("environments/prod/secrets/ghcr-secret.enc.yaml")
GHCR_SECRET_NAME = "ghcr-secret.enc.yaml"

# Modellnyckeln till agenten (MINIMAX_API_KEY) ligger krypterad i prod.
# Previewen kopierar in den krypterat - Flux dekrypterar den i
# preview-ns:et. Utan nyckeln kan agenten inte prata med modellen, så
# en preview-agent skulle sakna hela sin poäng.
AGENT_LLM_SECRET_SOURCE = Path("environments/prod/secrets/agent-llm.enc.yaml")
AGENT_LLM_SECRET_NAME = "agent-llm.enc.yaml"

# Schemat (ConfigMap) kopieras in från basen - samma schema som dev och
# prod. Overlayen får inte referera till basen, eftersom den publiceras
# på branchen "previews" som bara innehåller preview-overlays.
SCHEMA_SOURCE = Path("environments/base/db/schema.yaml")
SCHEMA_NAME = "schema-configmap.yaml"

# README i environments/pr/ - se ensure_readme().
README = Path("environments/pr/README.md")

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
    "AGENT_IMAGE_PLACEHOLDER": AGENT_IMAGE,
    "REPO_URL_PLACEHOLDER": REPO_URL,
    "REPO_BRANCH_PLACEHOLDER": REPO_BRANCH,
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

    llm_target = OUT_DIR / AGENT_LLM_SECRET_NAME
    llm_target.write_text(AGENT_LLM_SECRET_SOURCE.read_text(encoding="utf-8"), encoding="utf-8")
    print(f"✓ Kopierade {AGENT_LLM_SECRET_SOURCE} -> {llm_target}")

    schema_target = OUT_DIR / SCHEMA_NAME
    schema_target.write_text(SCHEMA_SOURCE.read_text(encoding="utf-8"), encoding="utf-8")
    print(f"✓ Kopierade {SCHEMA_SOURCE} -> {schema_target}")


def ensure_readme() -> None:
    """Skapar environments/pr/README.md om den saknas.

    Filen håller katalogen ifylld när den sista previewen tas bort. Utan
    den skulle kustomize bygga "no resources" och Flux inte prunade
    något - previewen skulle bli kvar i klustret.
    """
    if not README.exists():
        README.write_text(
            "# PR-previews\n\n"
            "Overlayer per PR, genererade av .github/scripts/pr-preview-resources.py.\n",
            encoding="utf-8",
        )
        print(f"✓ Skapade {README}")


if __name__ == "__main__":
    write_overlay()
    ensure_readme()
    print(f"Previewen deployas till namespace {NAMESPACE} med URL https://{HOSTNAME}")

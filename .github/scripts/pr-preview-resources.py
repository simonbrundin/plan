#!/usr/bin/env python3
"""Skapar kustomize-overlays för en PR-preview.

Skriptet körs av GitHub Actions. Miljövariablerna PR_NUM och
IMAGE_DIGEST kommer från workflow-steget.

Overlayen pekar på basen och applicerar de patchar som gör
preview-miljön unik: eget namespace, egen image och egen hostname.
"""

import os

import yaml

PR_NUM = os.environ["PR_NUM"]
IMAGE_DIGEST = os.environ["IMAGE_DIGEST"]

# Image-digesten kommer som "namn@digest". Kustomize behöver hela
# strängen, så den används som den är.
image = IMAGE_DIGEST

PR_DIR = f"environments/pr/{PR_NUM}"
NAMESPACE = f"plan-pr-{PR_NUM}"
HOSTNAME = f"plan-pr-{PR_NUM}.simonbrundin.com"

NAMESPACE_PATCH = {
    "apiVersion": "v1",
    "kind": "Namespace",
    "metadata": {
        "name": NAMESPACE,
        "labels": {
            "app.kubernetes.io/managed-by": "flux",
            "preview": "true",
            "pr-number": PR_NUM,
        },
    },
}

DEPLOYMENT_PATCH = {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "name": "nuxt-app",
        "labels": {"app": "plan", "preview": "true"},
    },
    "spec": {
        "replicas": 1,
        "template": {
            "spec": {
                "containers": [
                    {
                        "name": "plan",
                        "image": image,
                        "resources": {
                            "requests": {"memory": "64Mi", "cpu": "50m"},
                            "limits": {"memory": "128Mi", "cpu": "200m"},
                        },
                    }
                ]
            }
        },
    },
}

ROUTE_PATCH = {
    "apiVersion": "gateway.networking.k8s.io/v1",
    "kind": "HTTPRoute",
    "metadata": {
        "name": "plan-route",
        "annotations": {
            "external-dns.alpha.kubernetes.io/hostname": f"{HOSTNAME}.",
        },
    },
    "spec": {"hostnames": [HOSTNAME]},
}


def overlay() -> dict:
    """Bygger kustomization-filen för preview-miljön."""
    return {
        "apiVersion": "kustomize.config.k8s.io/v1beta1",
        "kind": "Kustomization",
        "resources": ["../../base"],
        "namespace": NAMESPACE,
        # Kustomize vill ha patchen som text - ett objekt går inte att
        # tolka i fältet "patch".
        "patches": [
            {
                "patch": yaml.dump(NAMESPACE_PATCH, default_flow_style=False, sort_keys=False),
                "target": {"kind": "Namespace", "name": "plan"},
            },
            {
                "patch": yaml.dump(DEPLOYMENT_PATCH, default_flow_style=False, sort_keys=False),
                "target": {"kind": "Deployment", "name": "nuxt-app"},
            },
            {
                "patch": yaml.dump(ROUTE_PATCH, default_flow_style=False, sort_keys=False),
                "target": {"kind": "HTTPRoute", "name": "plan-route"},
            },
        ],
    }


def write_overlay() -> None:
    """Skriver overlayen för PR:n."""
    os.makedirs(PR_DIR, exist_ok=True)
    path = f"{PR_DIR}/kustomization.yaml"

    with open(path, "w") as handle:
        handle.write("# DO NOT EDIT MANUALLY\n")
        handle.write("# Genereras av .github/scripts/pr-preview-resources.py\n\n")
        yaml.dump(overlay(), handle, default_flow_style=False, sort_keys=False)

    print(f"✓ Skapade {path}")


def update_pr_index() -> list:
    """Lägger in PR:n i environments/pr/kustomization.yaml."""
    path = "environments/pr/kustomization.yaml"
    resources = []

    if os.path.exists(path):
        with open(path) as handle:
            existing = yaml.safe_load(handle)
        if existing and "resources" in existing:
            resources = list(existing.get("resources", []))

    # Ta bort eventuella varianter av denna PR.
    resources = [
        resource
        for resource in resources
        if not (
            str(resource) == PR_NUM
            or str(resource) == f"./{PR_NUM}"
            or str(resource).endswith(f"/{PR_NUM}")
        )
    ]

    resources.append(f"./{PR_NUM}")
    resources.sort(key=lambda resource: (not str(resource).startswith("./"), str(resource)))

    with open(path, "w") as handle:
        handle.write("# DO NOT EDIT MANUALLY\n")
        handle.write("# Denna fil uppdateras automatiskt av GitHub Actions\n\n")
        yaml.dump(
            {
                "apiVersion": "kustomize.config.k8s.io/v1beta1",
                "kind": "Kustomization",
                "resources": resources,
            },
            handle,
            default_flow_style=False,
            sort_keys=False,
        )

    print(f"✓ Uppdaterade {path}")
    print(f"  Resurser: {resources}")
    return resources


def main() -> None:
    write_overlay()
    update_pr_index()
    print(f"✓ Image: {image}")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Tar bort kustomize-overlays för en PR-preview.

Skriptet körs av GitHub Actions. Miljövariablerna PR_NUM och
IMAGE_DIGEST kommer från workflow-steget.
"""

import os
import yaml

pr_num = os.environ['PR_NUM']

# Ta bort environments/pr/<PR_NUM>
pr_dir = f"environments/pr/{pr_num}"
if os.path.exists(pr_dir):
    import shutil
    shutil.rmtree(pr_dir)
    print(f"✓ Tog bort {pr_dir}")
else:
    print(f"- {pr_dir} fanns inte")

# Uppdatera environments/pr/kustomization.yaml
kustomization_file = "environments/pr/kustomization.yaml"
if os.path.exists(kustomization_file):
    with open(kustomization_file, 'r') as f:
        kustomization = yaml.safe_load(f)
    
    if kustomization and 'resources' in kustomization:
        original_count = len(kustomization['resources'])
        kustomization['resources'] = [
            r for r in kustomization['resources']
            if not (str(r) == pr_num or str(r) == f"./{pr_num}" or str(r).endswith(f"/{pr_num}"))
        ]
        
        if len(kustomization['resources']) < original_count:
            with open(kustomization_file, 'w') as f:
                f.write("# DO NOT EDIT MANUALLY\n")
                f.write("# Denna fil uppdateras automatiskt av GitHub Actions\n\n")
                yaml.dump(kustomization, f, default_flow_style=False, sort_keys=False)
            print(f"✓ Uppdaterade {kustomization_file}")
            print(f"  Kvarvarande resurser: {kustomization['resources']}")
        else:
            print(f"- {pr_num} fanns inte i kustomization.yaml")
    
    # Ta bort kustomization.yaml om inga resurser finns kvar
    if not kustomization.get('resources'):
        os.remove(kustomization_file)
        print(f"✓ Tog bort tom kustomization.yaml")
else:
    print(f"- {kustomization_file} fanns inte")

# Städa upp tomma mappar
try:
    os.rmdir(f"environments/pr")
    print(f"✓ Tog bort tom environments/pr/-mapp")
except OSError:
    pass  # Mappen finns kvar om den inte är tom

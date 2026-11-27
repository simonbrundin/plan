# PR-previews

Katalogen innehåller en overlay per PR-preview. Overlayerna genereras av
`.github/scripts/pr-preview-resources.py` och publiceras på branchen
`previews`, som Flux läser via Kustomization `plan-preview`.

Varje underkatalog är en egen resurs - kustomize byggs utan
rot-kustomization, så två PR:er behöver inte röra samma fil.

Filerna i den här katalogen genereras. Ändra dem inte manuellt - ändra
mallen i `environments/pr-template/` eller skriptet.

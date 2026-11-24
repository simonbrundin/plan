---
title: LightCraft + PhotoCraft installerade på Omarchy-workstation (utan sudo)
type: note
tags: [install, omarchy, photocraft, lightcraft, mcp, projekt:simon]
---

Båda ArtCraft-appar installerades 2026-11-16 från GitHub Releases (storytold/lightcraft v0.21.1, storytold/photocraft v0.23.1) som .pkg.tar.zst-paket.

Varför inte pacman: sudo kräver lösenord och ingen NOPASSWD-regel finns för pacman/pkexec-agent på maskinen. Lösning: paketen extraherades med bsdtar till ~/.local (bin/, share/applications/, share/icons/, bash- och fish-completion). Desktop-filer patchades med absolut Exec-sökväg eftersom fönsterhanteraren inte alltid hittar ~/.local/bin i PATH. Checksummor verifierades mot SHA256SUMS.txt.

PhotoCraft CLI (photocraft-cli) installerades globalt som MCP-server 'photocraft' i ~/.pi/agent/mcp.json: . 26 MCP-tools (doc_new, doc_open, command_run, doc_render_preview, doc_save, ui_* bridge-verktyg). Verifierat: pi mcp list -> connected, samt end-to-end-test av cli run med gaussianBlur -> PNG.

Fotnot: Arch-paketet för LightCraft innehåller INTE lightcraft-cli (finns bara för macOS). PhotoCraft-paketet innehåller både photocraft och photocraft-cli.

Uppdatering: hämta ny .pkg.tar.zst och ersätt filerna i ~/.local (eller kör sudo pacman -U om man vill ha systeminstallation).

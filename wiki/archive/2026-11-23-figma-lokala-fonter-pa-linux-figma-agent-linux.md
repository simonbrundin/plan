---
title: Figma lokala fonter på Linux: figma-agent-linux
type: note
tags: [figma, fonts, linux, omarchy, projekt:simon]
---

Figma har ingen officiell Font Helper på Linux (FigmaAgent finns bara för macOS/Windows). Lösning: figma-agent-linux (neetly) - en öppen reimplementering av Figmas lokala font-protokoll på 127.0.0.1:44950.

Installerad 2026-11-23 på maskinen:
- Binär: ~/.local/share/figma-agent/figma-agent (statisk binary, ~5MB)
- systemd user units: ~/.local/share/systemd/user/figma-agent.{service,socket} (socket-aktiverat, lyssnar på 127.0.0.1:44950)
- Config: ~/.config/figma-agent/config.json (valfritt; t.ex. bind, font_directories)
- Chromium-policy: ~/.config/chromium/policies/managed/figma-local-fonts.json -> LocalNetworkAccessAllowedForUrls for figma.com (Chromium 142+ frågar annars om 'Access other apps and services on this device')

Varför detta behövs: Figma-webbläsarversionen kan bara se lokala fonter via den lokala font-hjälpen. Figma-agent listar fonter via fontconfig (fc-list) - allt under ~/.local/share/fonts syns automatiskt, inkl. de tre Robin Springett-fonterna (4 Privet Drive, DearMrPotter, Hellboy).

Verifierat: agenten svarar på /figma/version och /figma/font-files inkl. de tre fonterna. Figma gör faktiskt en probe mot /figma/version även med Linux user-agent (testat headless via chromium --dump-dom), så UA-spoofning behövde inte för att nå agenten.

Kvar att göra i UI (en gång per browser): Figma-logon -> Figma-logotypen -> Preferences -> Permissions and helpers -> 'Apps on device' -> Show me how to connect -> I'm ready to connect -> tillåt local network access. Sedan syns fonterna i fontpickern under 'Installed by you'.

Fallback om fonterna inte listas: ~/.local/bin/figma-fonts startar Chromium med Windows-UA i en separat profil (~/.config/chromium-figma) - kräver separat Figma-login.

Viktigt: om ~/.config/chromium/policies/managed används MÅSTE webbläsaren startas om för att policyn ska läsas.

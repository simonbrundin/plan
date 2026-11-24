---
title: Pi desktop-notis när agent är klar
type: note
summary: Ny extension ~/.pi/agent/extensions/done-notify.ts: på agent_settled körs notify-send med titel 'Pi är klar' och en kort sammanfattning a...
tags: [pi, extension, notis, projekt:simon]
created: 2026-10-22
updated: 2026-10-22
status: active
---

Ny extension ~/.pi/agent/extensions/done-notify.ts: på agent_settled körs notify-send med titel 'Pi är klar' och en kort sammanfattning av sista assistenttexten (max 300 tecken) samt projektnamn. Miljövariabler: PI_NO_DONE_NOTIFY=1 stänger av, PI_DONE_NOTIFY_CMD byter kommando. Motivering: agenten kör ofta i bakgrunden i tmux/workmux så att man behöver en extern signal, inte bara TUI-toast. Använd ctx.sessionManager.getBranch() för att hämta sista assistant-meddelandet. Notifieringen är kosmetisk - fel fångas och ignoreras.

## Relaterade

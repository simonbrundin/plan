---
title: PhotoCraft: kamerans look-varning vid RAW-öppning och hur den hanteras
type: note
tags: [photocraft, raw, workflow, varning, projekt:simon]
---

Varningen 'colour fitted to the camera's embedded JPEG (mean ΔE 0.8 against it; 11.0 with the neutral fallback)' är INTE ett fel utan en develop-not (crates/raw/src/look.rs). Sony ARW saknar inbäddad ICC-kalibrering, så PhotoCraft sätter en 'camera look' (3x3-matris + tonkurva) som anpassas till kamerans egen JPEG-preview. ΔE 0.8 = nästan exakt kamerans rendering; neutral-fallback ger ΔE 11.0 (platt/grönaktig). Det finns ingen inställning för att stänga av den (crates/io/src/raw.rs develop_fitted: den passar alltid in när filen saknar kalibrering och inte är mono).

Hur man 'lösar' den i praktiken: utveckla rawen en gång till .pcraft (får inbäddad sRGB-profil) och jobba vidare på .pcraft - då kommer varningen inte tillbaka. Dessutom: kontrollera alltid vitbalansen explicit (mätt ljusaste pixlarnas kromaticitet innan du låter kamerans look stå).

Fälla som kostade tid: rawDefaults.openInCameraRaw=true (i preferences.json: openInCameraRaw) öppnar RAW i ett modal Camera Raw-fönster ('RAW Develop (DSC03172.ARW)') - medan det är öppet svarar alla engine-kommandon 'command is unavailable while a modal dialog is open'. Satt openInCameraRaw=false för att kunna jobba via automation.

Fälla 2: automationens kontrollkanal (--control) behöver att fönstret presenterar frames, annars time:a ut (Wayland: inga frame-callbacks för fönster på annan workspace). Fokusera fönstret med hyprctl eval 'hl.dsp.global("workspace N")' - gamla dispatch-syntaxen 'hyprctl dispatch workspace N' är trasig i den här Hyprland-versionen.

Fälla 3: occh användaren i taget kan ändra dokumentet (t.ex. dra Canvas Size) medan automationen kör - kontrollera dimensioner efteråt.

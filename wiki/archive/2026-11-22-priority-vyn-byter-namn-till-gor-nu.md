---
title: Priority-vyn byter namn till Gör nu
type: note
tags: [feature, frontend, constraints, projekt:plan]
---

Gör nu introducerades som ny route /now. Den gamla /priority-routen redirectar till /now. Vyns nav- och hjälptexter använder Gör nu och leader-kommandot är G+N. Backendens prioriterade mål filtrerar nu bort mål vars direkta beroenden inte är slutförda via NOT EXISTS. Hårda enhetskrav (dator/mobil) och en separat framstegspoäng är ännu inte implementerade; nuvarande vikt och manuella ordning finns kvar.

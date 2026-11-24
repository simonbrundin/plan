---
title: Brevlådelistan måste hämtas innan granskningsläget startar
type: note
tags: [nuxt, mail, ui, bug, projekt:pappa]
---

2026-11-19: I pappa/appens granskningsvy försvann brevlådelistan helt när granskningen startade, och den kom inte tillbaka när man tryckte "Stäng".

Varför: listvyn (EmailList) bygger på emails-arrayen, men analyzeEmails fyllde bara queue/analyses - aldrig emails. Efter en lyckad Acceptera anropades fetchEmails, så då fungerade det, men direkt efter "Processa mejl" + "Stäng" var listan tom.

Fix: analyzeEmails anropar nu fetchEmails när granskningsläget startas. Då är listan alltid färdig när granskningen stängs.

Generell regel: en vy som växlar mellan listläge och detaljläge måste ha samma underlag i båda lägena - annars ser det ut som att data försvann.

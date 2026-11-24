---
title: Pappas grundare är Michael Brundin - inte git-användaren Simon Brundin
type: note
tags: [pappa, grundare, identitet, auth, miss, projekt:pappa]
---

2026-11-10: När redigerarlistan för relationer skulle seedas tog jag namnet från git config (user.name = 'Simon Brundin') och lade in Simon Brundin som chef. Det var FEL: grundaren av pappa/Kameo är Michael Brundin (Michael Brundin AB, Kameo - Michael Brundin AB). Simon Brundin är git-användaren som kör agenterna - inte appens ägare.

VARFÖR DET SPARAR TID: git config user.name är ett HANTVERK - det säger vem som committar, inte vem som är principal. I pappa ska 'chef' i redigerarlistan (relation_editor) vara Michael Brundin. Relationer som Michael läger in blir ägda av Michael.

Korrigering gjort: relation_editor.replace('Simon Brundin' -> 'Michael Brundin', role lead). Inget i repot behövde ändras - namnet fanns bara i databasen.

FALLGROP: 1Password-objektet 'Kameo - Michael Brundin AB' och mailkontot michael@brundins.se hör till samma principal - när man ska lista 'teamet' i pappa ska Michael Brundin vara med, inte Simon.

---
title: bash-fallgropar i pxe-flash: tab som IFS-whitespace och read med for få variabler
type: note
summary: Två bash-fallgropar som nästan släppte igenom i `simon pxe flash`
tags: [bash, bug, pxe, cli, projekt:simon]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [entities/pxe-server]
---

Två bash-fallgropar som nästan släppte igenom i `simon pxe flash`
(scripts/pxe-flash.sh i simon-cli):

1. Tab är IFS-whitespace. `IFS=$'\t' read -r a b c` kollapsar angränsande
   tabb och tar bort tomma fält i början/slutet. En enhetsrad med tom
   markör ("/dev/sdb|0||<etikett>") lästes därför som om markören vore
   etiketten - och systemdisk-spärren jämförde fel värde. Lösning:
   använd "|" som delimiter (icke-whitespace) mellan fälten.

2. `read -r f1 f2 f3` ger f3 resten av raden. Med fyra fält måste man
   alltid läsa fyra variabler, annars hamnar etiketten i markörfältet.

Båda fångades av att guard-spärren testades med riktiga serverns
enhetslista (systemdisk + tom kortläsare) - inte bara unit tests.

## Relaterade

- [[entities/pxe-server]]

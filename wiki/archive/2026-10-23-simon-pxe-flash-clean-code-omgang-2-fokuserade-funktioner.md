---
title: simon pxe flash: clean code-omgång 2 - fokuserade funktioner
type: note
tags: [clean-code, refactor, pxe, cli, projekt:simon]
---

Andra refaktoreringen av simon-cli/scripts/pxe-flash.sh (efter första
omgången som delade upp main_pxe_flash i 18 funktioner).

Vad som fortfarande luktade:
- select_device gjorde tre saker (bygga val-lista, välja, validera)
  -> choose_device + assert_device_is_safe + select_device
- fzf-mönstret duplicerades -> select_with_fzf
- enhetsetiketten byggdes i ett svep -> device_flags + device_label
- verify_written gjorde läsning + jämförelse + rapport -> sectors_for_file,
  remote_sha, compare_sha
- explain_no_image hårdkodade "Raspberry Pi 4" -> board_manual_hint
- magiska tal 512/511/1073741824 -> konstanter
- pipe-delimiter -> PXE_FLASH_FIELD_SEP, med field/join_fields

Lärdom: guard-funktioner som returnerar ett "record" (path|size|marker)
måste läsas med rätt antal variabler, annars får sista variabeln resten
av raden. Bash-fällan är samma som med IFS-whitespace: tomma fält
försvinner om delimiter är whitespace.

Verifiering: 43 enhetstester + de fyra kodvägarna (flash, systemdisk,
ingen media, Pi4) körda mot riktiga PXE-servern.

---
title: Ny RS232-hårdvara: AK-CO-02 + DB9-breakout, och kopplingen till KVM:ns 3-pin
type: note
tags: [kvm, rs232, hardware, bug, projekt:simon]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Inköp 2026-11-20: AKYGA AK-CO-02 (B07TXJXWJ3) = USB-A hane -> DB9 **hane** RS-232-kabel, 1 m, chip oklart (Akyga-länken pekar på FT232-drivrutiner, en amazon-recension säger CH340 - verifiera med lsusb på Pi:n). YIOVVOM DB9-breakout (B07DL13B32) = DB9 **hona** -> 9 skruvterminaler, silkscreen 1-9.

Kopplingskedja: Pi USB -> AK-CO-02 (DB9 hane) -> DB9-hona-breakout -> 3 skruvar -> 3 kablar -> KVM:ns 3-pin terminal.

Verifierat med loopback att AK-CO-02 kan både sända och ta emot. KVM:ns terminalmärkning är omvänd: med adaptern frånkopplad och KVM påslagen gav R till G −8,8 V (TX i RS-232-vila) medan T till G gav 0 V (RX). Fungerande koppling är därför adapter TX/DB9 pin 3 → KVM T, KVM R → adapter RX/DB9 pin 2, G → DB9 pin 5. Vid omvänd signaldragning fick vi bara 0xFF-brus.

Signalfornamn i DB9 (DTE/PC-sida, DB9 hane): pin 2 = RXD, pin 3 = TXD, pin 5 = GND. Korskoppling: KVM TX -> pin 2, KVM RX -> pin 3, GND -> pin 5. **Verifiera silkscreensen** pa breakout-bordet med multimeter (kontinuitet till pin 5) - billiga Female-breakouts har ibland spegelvand numrering.

CH340 gotcha pa Ubuntu (Pi): brltty snappar CH340 - sudo apt remove brltty, sedan dyker /dev/ttyUSB0 upp. Verifierat fungerande protokoll för MLEEDA KC-KVM1601A (Amazon ASIN B0GCD38QN9): 115200 baud, 8N1, ASCII `X3,1$` utan radslut byter till port 3. KVM:en svarar med `R1:[0]:3,[3]:1\r\n`. TESmart-binärramar och ASCII-varianter med punkt/radslut fungerade inte. Config: /etc/kvm-switch/config.json, device /dev/ttyUSB0.

---
title: Talos v1.14 utan UnattendedInstallConfig: oinstallerad metal-nod bootar om i en loop
type: note
tags: [pxe, talos, reboot-loop, config, projekt:infrastructure]
---

Symtom: noden startar om var ~2:a minut. Matchbox-loggen får GET
/assets/talos-config/nodes/<ip>.yaml varje gång, men noden hinner aldrig stanna.
Inga felmeddelanden i API:t (noden svarar inte alls mellan boots).

ROTORSAK: Talos v1.14:s install-sekvens (runtime.SequenceInstall) gor:
- finns UnattendedInstallConfig i configen -> sekvensen HOPPAS OVER, och
  installationen sker istallet av UnattendedInstallController (utan reboot)
- finns det inte -> Talos installerar med machine.install. Finns inte heller
  det ar disken tom -> filepath.EvalSymlinks("") misslyckas -> sekvensen
  returnerar fel -> i internal/app/machined/main.go:
      if err := c.Run(ctx, runtime.SequenceInstall, nil); err != nil { return err }
  -> machined avslutar -> noden bootar om. Loop.
Loggbild: "installing Talos to disk ." (tomt disk-namn) + "pulling image".
Det var detta som gjorde worker-2 (10.10.10.22) till en reboot-loop, inte
efivarfs-felet.

FIX: varje nodconfig MÅSTE ha ett UnattendedInstallConfig-dokument.
- maintenance-config (worker-N.yaml): diskSelector matchar ingen disk
  (disk.serial == "0xFFFFFFFF") + reboot: false -> ingen disk skrivs,
  ingen reboot, men sekvensen hoppas over.
- install-config (worker-N-install.yaml): diskSelector = nodens egen disk,
  reboot: false pa ROCK (noden ska aldrig lämna PXE).

Detta genereras nu av pxe/generate-node-configs.py (konstanterna NO_DISK och
install_doc). Se även noten om partitionerna som MÅSTE skapas.

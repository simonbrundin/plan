---
title: Mac mini (worker-9) kraschar vid tyngre USB-I/O på Talos-stickan
type: note
tags: [bug, mac-mini, worker-9, usb, projekt:simon]
---

# Symptom
Tre incidenter 2026-11-24: Mac minin (worker-9, 10.10.10.29) hänger/kraschar när
USB-stickan (Kingston DataTraveler 3.0, Talos-systemdisk) utsätts för tyngre I/O
från macOS. Typiska fall:

1. Montering av FAT32-ESP:en + kopiering av 116 MB UKI -> crash.
2. dd-läsning av 100 MB STATE + 64 MB ESP + scp av 164 MB -> crash.
3. Samma som (2) -> crash igen.

Symptom: TCP-anslutningar till 22 accepteras men ingen SSH-banner, ping ger
inget svar, maskinen svarar inte pa ARP. Maskinen maste strömcyklas fysiskt.
Load average var 62 direkt efter boot (nagot far hogt).

# Varfor det spelar roll
Det gor varje macOS-fonster dyrt: vi har bara nagra minuter per gang innan
maskinen kan hanga igen. Därfor arbetar vi med sa lite USB-I/O som mojligt fran
macOS.

# Regel
- Montera ALDRIG FAT32-partitionen pa USB-stickan fran macOS igen.
- Anvand raw dd for att lasa/skriva partitionerna.
- Skriv bara en gang per macOS-fonster (t.ex. en 100 MB dd pa STATE).

# Teknisk kunskap: STATE-partitionens layout
USB-stickans partitioner: s1 = FAT32 ESP (ca 2,2 GB), s2 = liten linux,
s3 = XFS "STATE" (100 MiB), s4 = XFS "EPHEMERAL" (59,7 GB).

Talos lagger maskinconfigen i roten av STATE-filsystemet som config.yaml
(v1alpha1 MachineConfig + extra dokument, t.ex. UserVolumeConfig for ssd-1).
Dessa filer finns i STATE: config.yaml, node-identity.yaml,
encryption-salt.yaml, platform-network.yaml.

# Gammal PKI pa noden (viktigt for att fa API:t att svara)
Nodens nuvarande (april 2026) maskin-CA har fingerprint B583226CD35A7F8A -
det ar EXAKT samma CA som infrastructure/talos/talosconfig-new och
talosconfig.expired. Alltsa: klientcertet i talosconfig-new kan autentisera
mot nodens Talos-API om apid (port 50000) startar. Den nuvarande
~/.talos/config (CA 5B9A553882B81313) ar for den nuvarande klustergenerationen
och kan INTE anvandas mot den gamla configen pa stickan.

# Gjort den 24 nov
- Ny STATE-avbildning med aktuell PKI (maskin-CA 5B9A55, cluster.id
  c+KQTosq..., cluster.secret AJpWNf..., kubelet v1.37.0, installer v1.14.1)
  byggd lokalt: patchad kopia av gamla config.yaml, validerad med
  "talosctl validate --mode metal" och "xfs_repair -n".
- Profilen node-worker-9 i pxe/matchbox serverar fortfarande worker-9.yaml
  med PI5-diskSelector (disk.serial == "0xFFFFFFFF") - maste fixas.
- Matchbox pa 10.10.10.2:8080 svarade inte (tomt svar) - PXE-vagen var
  dodsdod just nu.

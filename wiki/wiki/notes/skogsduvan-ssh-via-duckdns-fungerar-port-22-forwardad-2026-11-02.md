---
title: "Skogsduvan: SSH via duckdns fungerar (port 22 forwardad 2026-11-02)"
type: note
tags: [ssh, homeassistant, duckdns, drift, projekt:infrastructure]
created: 2026-11-02
updated: 2026-11-02
status: active
related: [notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen]
---

**Status 2026-11-02: löst.** Simons har vidarebefordrat port 22 i stugans router. `ssh skogsduvan` fungerar nu direkt från hemmet med samma nyckel som `homeassistant.local`.

Verifierat:
```
$ ssh skogsduvan 'hostname; uptime; whoami'
core-ssh
 22:29:57 up 34 days,  3:24,  0 users,  load average: 0.06, 0.08, 0.06
root
```
- Host key for hosten matchar homeassistant.local/192.168.189.2: `SHA256:O+N+hEkKXbgrWOQP9rROr9i7h7aHQu8hIcmRisse83U` -> ingen MITM, det ar samma maskin.
- Inget nyckelbyte behovdes: `~/.ssh/stuganha` (RSA 4096, `SHA256:/Hk40UVuiRokemoj7sGNFaQhTuPuXKTVQN5Qz1hgjlw`), anvandare `root`.

Viktig detalj: `hostname` svarar **`core-ssh`** - det ar HA:s *Terminal & SSH*-addon (containern), inte varden `192.168.189.2` som HA Supervised kor pa. Alltsa: `docker ps`, `ha` och dylikt finns inte direkt i den shellen; for att na varden gaar man via addonens supervisor-API/socket.

Fakta om maskinen (undersokt samma dag, fore forwarden):
- stuganskogsduvan.duckdns.org -> 78.69.91.181 (Telia/CGNAT).
- Home Assistant Supervised pa Debian-vard 192.168.189.2/24 (end0), gateway 192.168.189.1.
- Offentliga portar: 80 (tomt svar), 443 (HA), 8123, och nu 22.
- LAN-vagen 192.168.189.2 gar inte att na via Twingate: publicerade resurser ar bara "Hemnatverk" 10.1.1.0/24 och "Datacenter datorer" 10.10.10.0/24, ingen route till 192.168.189.0/24.

Varfor det var ett problem: tidigare foljetjanter anvande `ssh -i ~/.ssh/stuganha root@192.168.189.2`, vilket bara fungerade nar datorn var i stugans LAN. Fran hemmet fanns ingen LAN-vag och routeren forwardade inte 22.

Alternativ som funkar utan routerandring: publicera stugans natverk via en Twingate-connector, eller anvand HA:s Terminal & SSH-add-on via ingress over 443 (kräver HA-inloggning).

Arbetsnotering: `~/.ssh/config` har Host `skogsduvan` -> stuganskogsduvan.duckdns.org med stuganha-nyckeln och `StrictHostKeyChecking accept-new`.

## Relaterade

- [[notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen]]


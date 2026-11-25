---
title: Kubeconfig-städning: en kontext mot cluster1 (admin@cluster1), 6 dubbletter borttagna
type: note
tags: [kubeconfig, cluster1, pki, städning, projekt:infrastructure]
---

2026-11-25: ~/.kube/config innehöll 8 kontexter mot samma Talos-kluster. -1-1 samt -2..-6 pekade alla på https://10.10.10.10:6443 med identisk CA och samma admin-roll (O=system:masters) - de var funktionella dubbletter, skapade av att `talosctl kubeconfig --merge` döper om en krockande "cluster1" till cluster1-N. admin@cluster1 och admin@cluster1-1 pekade på 10.10.10.12 med en gammal CA (rotation 2026-10-22) -> x509 unknown authority.

Åtgärd: admin@cluster1 reparerades (ny CA + admin-cert från -2, server = VIP 10.10.10.10), och -1-1, -1, -2..-6 togs bort tillsammans med sina överblivna cluster/user-entry. Backup: ~/.kube/config.pre-cleanup.20261125_222345.

Varför: sex fungerande admin-kuber gjorde det oklart vilken kontext som var rätt, och risken är att man väljer den gamla CA:n och får x509-fel. Nu finns exakt en giltig admin-kontext.

FÄLLA (viktig): `talosctl kubeconfig` med --merge (default) lägger in en NY kontext varje gång den körs mot en befintig "cluster1" - dubbletterna kommer tillbaka. Kör alltid mot en tempfil istället: `talosctl --nodes 10.10.10.12 kubeconfig /tmp/kc.yaml --merge=false`, och kopiera in CA+cert manuellt om man vill byta ut admin@cluster1.

oidc-cluster1 (exec-pluginkontext mot Zitadel) fungerar INTE och kunde inte rättas med CA-byte: API-serverns faktiska authentication-config har `jwt: []` (verifierat via talosctl read /system/config/kubernetes/kube-apiserver/authentication-config.yaml), så OIDC-jwt stöds inte av apiservern just nu. Patchen i repot (talos/patches/controlplane.yaml) har jwt-konfigurationen men den är inte applicerad på noderna. Tidigare version av patchen använde prefix "zitadel-" på username-claimet; aktuell patch saknar prefix och använder username-claim=sub med audiences [391307619412774131] - kontrollera att det stämmer mot ClusterRoleBindings i infrastructure-flux/components/auth (där namnet är "zitadel-346062090475757945", vilket INTE matchar prefix + sub = "zitadel-34606209047557945").

Referenser uppdaterade: plan/AGENTS.md och infrastructure/pxe/Bootloaders/make-talos-sd-card.sh pekar nu på admin@cluster1.

---
title: Query-parametrar i Actor Smartbook-endpointen: '+' URL-avkodas till mellanslag
type: note
tags: [actorsmartbook, stk, bug, http, projekt:stk]
created: 2026-11-06
updated: 2026-11-06
status: active
---

Fallet: Scalar UI prefillade defaultvärdena 'now' och 'now + 7 days' och skickade dem som text -> 400 'invalid from'.

Två delfaller:
1) Parametervärdena 'now'/'now + 7 days' var rena strängar. Nu accepteras RFC3339, YYYY-MM-DD, 'now', 'today' och relativa uttryck (now+7d, now-1d, 'now + 2 hours').
2) VIKTIGT: '+' i en query-sträng URL-avkodas till MELLANSLAG. 'now+7d' anländer alltså som 'now 7d' - tecknet '+' finns kvar i URL:en men försvinner i avkodningen. Därfor är signumtecknet optionalt i parsern och ett saknat tecken räknas som plus. Testa alltid query-parametrar mot en riktig server, inte bara enhetstester.

Fallet är generellt: alla framtida tidsparametrar med plustecken i docs-defaults behöver denna hantering.

---
title: Priorityfilter ska behandla saknat started som ej påbörjat
type: note
tags: [bug, priority, filter, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/mal-kan-ha-due-datum-deadline]
---

På /priority matchade filtret Påbörjade mål utan startdatum när Go-modellen serialiserade started med omitempty. Då försvann null-fältet och frontendens started !== null blev sant för undefined. Fixen är att Go alltid serialiserar started och finished, även som null, och att frontendens matchesStartedFilter normaliserar undefined till null. Standardfiltret på /priority är Ej påbörjade.

## Relaterade

- [[notes/mal-kan-ha-due-datum-deadline]]


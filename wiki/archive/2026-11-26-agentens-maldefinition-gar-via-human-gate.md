---
title: Agentens måldefinition går via human gate
type: note
tags: [agent, human-gate, session, projekt:plan]
---

Agenten kan nu föreslå hur ett mål ska se ut - titel, beskrivning, ikon, now/eternal/agent, datum, uppskattad tid, status, taggar, platser och upprepning. Förslaget lämnas som ett godkännande i Plan (typ goal_definition) och appliceras exakt som det godkändes - agenten genererar aldrig om efter beslut.

Varför: agenten ska fylla luckor, inte skriva över användarens värden. Ett befintligt värde lämnas orört, taggar och platser slås ihop med målets egna, ogiltiga värden hoppas över med en notering. Kodförslag och definitionsförslag skiljs åt via godkännandets typ (goal_definition / code_change) så att samma körning inte blandar ihop dem.

Konsekvenser: kontraktet är version 3, agent-scopet tillåter platser och målets tagg-/platsresurser, och Plan avvisar en okänd godkännandetyp. Kodern körs först när definitionen är godkänd och applicerad - ett avslag stoppar körningen.

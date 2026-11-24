---
name: orchestrator
description: |
  Koordinerar checkpoint-gate arbetsflödet. Används för att köra stora features
  genom checkpoints med gates.
  
  Utlösare: "orchestrator", "checkpoint-gate", "kör workflow", "bygg med checkpoints"
---

# Orchestrator Skill

Koordinerar Shopifys Helix-inspirerade checkpoint-gate arbetsflöde.

## Övergripande flöde

```
Feature Request
      ↓
Checkpoint Planner (sub-agent, fresh context)
      ↓
Human Approval (du)
      ↓
┌─────────────────────────────────────────┐
│   Per Checkpoint:                       │
│                                         │
│   1. Test Writer (sub-agent)            │
│   2. Implementation (sub-agent)         │
│   3. Behavior Gate (automated)          │
│   4. UI Gate (sub-agent, om relevant)   │
│   5. Adversarial Gate (2 sub-agents)   │
│   6. Human Gate (du)                    │
└─────────────────────────────────────────┘
      ↓
     PR
```

## Användning

Starta workflowet genom att säga:

> "orchestrator: bygg [feature-namn]"
> "kör checkpoint-gate workflow för [beskrivning]"
> "Jag vill implementera [feature] med checkpoints"

## Exempel-prompts

### Start a feature

```
orchestrator: Implementera ett notifikationssystem för användare

Features:
- Skicka notifikationer via email och push
- Användare kan välja kanaler
- Notifikationer loggas för historik
```

### Review and approve plan

Workflowet visar checkpoint-plan och väntar på ditt godkännande.

Svara:
- `ja` / `approve` / `y` - Godkänn och börja köra
- `nej` / `reject` / `n` - Avvisa och avsluta
- `ändra [checkpoint/tasks]` - Begär ändringar

### During execution

Om agenten frågar om hjälp vid en checkpoint:

Använd `/checkpoint status` för att se aktuell status.

### Request changes

Efter human gate kan du begära ändringar:

> "ändra X i checkpoint 2"

Ändringen blir en ny checkpoint och går genom samma gates.

## Kommandon

- `/checkpoint status` - Visa nuvarande checkpoint och gate-status
- `/checkpoint plan` - Visa hela planen
- `/learnings` - Visa tidigare lärdomar
- `/abort` - Avbryt workflow

## Gates

### Behavior Gate
Kör automatiska tester, lint, och vet. Blockerar om något failar.

### UI Gate
Jämför implementation med design/prototype. Kräver agent med fresh context.

### Adversarial Gate
Critic + Fixer loop. två agenter som criticar och fixar tills inga issues återstår.

### Human Gate
Du testar manuellt och godkänner.

## Principer

1. **Checkpoints bygger på varandra** - Complexity ökar
2. **Gates blockerar** - Agenten kan inte gå vidare förrän gate passerar
3. **Human in the loop** - Planen och slutresultatet granskas av människa
4. **Learnings sparas** - Feedback dokumenteras för framtida agenter

## Konfiguration

```yaml
orchestrator:
  max_adversarial_iterations: 3
  gate_timeout: 30m
  viewer: console  # eller html
```

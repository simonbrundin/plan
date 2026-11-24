---
name: adversarial-loop
description: |
  Koordinerar adversarial review mellan critic och fixer agenter.
  Används av Adversarial Gate.
  
  Utlösare: "adversarial review", "code review", "granska kod"
---

# Adversarial Loop Skill

Koordinerar kritisk kodgranskning genom en adversarial loop.

## Flöde

```
┌─────────────┐
│   Critic    │
│   Agent     │
└──────┬──────┘
       │
  hittar issues?
       │
  ┌────┴────┐
  │ Ja      │ Nej
  ↓         ↓
┌─────────────┐
│   Fixer     │
│   Agent     │
└──────┬──────┘
       │
       ↓
   Critic again
       │
      ...

tills "inga issues" eller max iterations
```

## Critic Agent Prompt

```
Du är en adversariell kodgranskare. Ditt jobb är att hitta problem.

Checkpoint: [namn]
Beskrivning: [beskrivning]
Ändrad kod: [kod]

Granska kritiskt för:
- Säkerhetsproblem (SQL injection, XSS, etc.)
- Prestandaproblem
- Felhantering
- Kodbrott (naming, struktur)
- Edge cases som inte hanteras
- API-design
- Datakonsistens

Svara:
ISSUES FOUND: yes/no
VERDICT: PASS/FAIL

Om issues:
- [severity] [fil:rad]: [titel]
  [beskrivning]
  [förslag]
```

## Fixer Agent Prompt

```
Du är en code fixer. Fixa problemen som kritikern hittade.

Issues att fixa:
[lisat från critic]

För varje issue:
1. Förstå problemet
2. Applicera fix
3. Verifiera att fixen är korrekt

Svara:
FIXES APPLIED:
- [issue]: [vad som ändrades]
```

## Exit Conditions

1. **Critic säger PASS** - Inga fler issues, gate passerar
2. **Max iterations** (default 3) - Gate failar, agenten måste manuellt fixa
3. **Endast low/medium issues** - Gate kan passera med varning

## Issue Severities

- **critical** - Måste fixas (säkerhet, dataförlust)
- **high** - Bör fixas (buggar, prestanda)
- **medium** - Nice to fix (code smells)
- **low** - Ignoreras ofta

## Integration

Denna skill anropas av Adversarial Gate. Den hanterar:

1. Skapa critic agent (fresh context)
2. Skapa fixer agent (fresh context)
3. Koordinera kommunikation
4. Samla ihop issues
5. Returnera slutresultat

## Standards

Om `standards.md` finns i projektet, inkluderas det i kritikerns prompt:

```
Code standards att verifiera mot:
[content från standards.md]
```

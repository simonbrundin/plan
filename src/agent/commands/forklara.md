---
allowed-tools: Read, Glob, Grep, Bash, WebFetch, FetchContent
description: Skapa en lättläst HTML-sida som förklarar ett ämne
argument-hint: [ämne att förklara]
---

# Förklara

## Syfte

Skapa en väldesignad och lättförståelig HTML-sida i
`/home/simon/repos/docs/html` som förklarar det ämne användaren frågar om.
HTML-sidan ska vara informativ, visuellt tilltalande och enkel att läsa.

## Variabler

ÄMANE: baserat <INPUT>: baserat på det användaren frågar om, skapa en HTML-sida i /home/simon/repos/docs/html som förklarar ämnet på ett lättförståeligt sätt

## Workflow

1. Analysera `ÄMANE` för att förstå vad användaren vill lära sig
2. Sök efter relevant dokumentation eller information om ämnet om det behövs
3. Skapa en väldesignad, responsiv och lättläst HTML-sida i
   `/home/simon/repos/docs/html/` med följande struktur:
   - använd ett tydligt sidnamn baserat på ämnet
   - inkludera tydliga rubriker och underrubriker
   - dela upp informationen i lättsmält sektioner
   - använd visuella element som listor, tabeller eller kodexempel vid behov
   - lägg till navigation eller innehållsförteckning om sidan är lång
   - inkludera metadata som titel och beskrivning
4. Spara HTML-filen i `/home/simon/repos/docs/html/` med ett beskrivande filnamn
5. Verifiera att filen skapades korrekt

## Relevanta filer

- `/home/simon/repos/docs/html/` - Mappen där HTML-dokumentation ska sparas
- Eventuella befintliga HTML-filer i mappen för att matcha stil och konventioner

## Exempel

Exempel på kommando:
`/forklara hur Kubernetes pods fungerar`

Genererad fil: `/home/simon/repos/docs/html/kubernetes-pods.html`

Exempel på HTML-struktur som ska skapas:

```html
<!DOCTYPE html>
<html lang="sv">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Ämne - Dokumentation</title>
  <style>
    /* Tydlig och ren styling */
  </style>
</head>
<body>
  <header>
    <h1>Ämne</h1>
    <nav>Innehållsförteckning</nav>
  </header>
  <main>
    <section>
      <h2>Översikt</h2>
      <!-- Enkelt att förstå förklaring -->
    </section>
    <!-- Fler sektioner vid behov -->
  </main>
</body>
</html>
```

## Rapport

- Returnera sökvägen till den skapade HTML-filen
- Ge en kort sammanfattning av vad sidan täcker
- Ange eventuella källor som användes för att skapa dokumentationen

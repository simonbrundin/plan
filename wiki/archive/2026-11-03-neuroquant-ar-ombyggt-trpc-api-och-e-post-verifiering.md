---
title: NeuroQuant är ombyggt - tRPC-API och e-post-verifiering
type: note
tags: [neuro, scraper, neuroquant, projekt:neuro]
---

## Vad som ändrats (2026-11-03)

NeuroQuant har byggts om. Det gamla DOM-skrapping-läget fungerar inte längre:

- sajten är lokaliserad (/se/login, /se/home, /se/portfolios)
- den gamla URL:en /portfolios ger 404
- portföljtabellen är inte längre nåbar med de gamla selectorerna
- inloggning kräver ett 6-siffrigt verifieringskode som mejlas till kontot

## Lösning

- scraper.py använder tRPC-API:t (app.neuroquant.ai/api/trpc):
  watchlist.get + strategy.getPortfolioData. Portföljen väljs på titel
  ("NQ Värde & Momentum") och datumet kommer från updatedAt.
- Inloggning: cookies först (cookies.json), annars lösenord + verifieringskod.
  Koden hämtas automatiskt från inkorgen via IMAP (EMAIL_USER/EMAIL_PASSWORD) -
  därför ska NeuroQuant-kontot vara samma konto som inkorgen.
- portfolio_data.py har all formatering/tolkning (tabellrader, datum,
  verifieringskod, låst konto/felaktigt lösenord) och testas utan browser.
- run_guard.py (max körningar/dygn, pause vid fel) ligger kvar oförändrat.

## Viktigt

NeuroQuant-kontot som fungerar är simonbrundin@gmail.com (samma konto som
inkorgen). michael.brundinab@gmail.com går inte att läsa brev från, så där
skulle verifieringskoden inte kunna hämtas automatiskt.

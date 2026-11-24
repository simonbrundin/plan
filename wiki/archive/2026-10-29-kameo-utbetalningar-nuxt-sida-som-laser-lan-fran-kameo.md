---
title: Kameo utbetalningar: Nuxt-sida som läser lån från Kameo
type: note
tags: [kameo, nuxt, 1password, api, projekt:mailmaster]
---

Skapade sidan /kameo-utbetalningar i src/nuxt (Nuxt/Nitro). Den loggar in på Kameo server-side via api.kameo.se/v1 och klassificerar varje lån som kortsiktig (<12 mån kvar) eller långsiktig.

Varför: Michael Brundin AB har investeringar i Kameo och behöver snabbt veta om en tillgång är kort- eller långsiktig.

Lösning: Inloggning sker med POST /v1/authenticate/2fa {email,password,otp} där otp är TOTP (Google Authenticator-format). Uppgifterna läses från 1Password-objektet 'Kameo - Michael Brundin AB' via op-cli (op item get --otp). Data: GET /v1/account/transactions-by-loan?home_currency=SEK&per_page=50 (loan_number, loan_duration_months, status), GET /v1/account/risk-distribution (slutförsdatum i next_amortization_payment.due_date per LoanApplication.<nr>), GET /v1/account/upcoming-repayments (kommande återbetalningar).

Viktigt: ap.kameo.se är React Native Web; inloggningen sker via json-api, inte webbskrapning. 1Password-cli:t kräver att op-anropet sker i samma session som 1Password-appen (annars 'connection reset') - i Docker måste man därför använda KAMEO_EMAIL/KAMEO_PASSWORD/KAMEO_TOTP_SECRET som reserv.

Klassificering: kortsiktig om återstående löptid (månader till slutbetalning) < 12, annars långsiktig; för återbetald lån används lånets ursprungliga löptid.

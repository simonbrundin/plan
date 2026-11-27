---
title: Plan-API: webbläsaren anropar API:t direkt - Nuxt-proxyn är borttagen
type: note
tags: [api, nuxt, cors, auth, session, refaktor, projekt:plan]
---

Borttog Nuxt:s proxy mot Plan API:t (2026-11-27).

Varför: proxyn (/api/proxy + forwardToGo i server/utils/goApi.ts) fanns för att föra fram sessioncookien när authen flyttades till Go API:t (2026-11-10). Men API:t är redan publikt på samma host (plan.simonbrundin.com/api/v1) och validerar sessionen själv - proxyn gav ingen säkerhetsvinst, bara en extra hopp och buffring (forwardToGo läste response.text(), så ingen streaming).

Vad som ändrades:
- /api/proxy/* och /api/auth/* (Nuxt:s server routes) är borttagna. Klienten anropar nu NUXT_PUBLIC_GO_API_URL direkt med credentials: include (app/utils/api.ts).
- Kvar i Nuxt: bara agent-API:t (/api/agent/conversations, /api/agent/workflow) - AGENT_API_KEY får inte nå webbläsaren. Dit går även sessionkontrollen server-till-server.
- CORS: API:t (både Go och Rust) svarar nu med CORS_ALLOWED_ORIGINS (kommaseparerad lista). Origin echas tillbaka, aldrig wildcard - ett credentialed anrop med sessioncookie blockeras av webbläsaren om Access-Control-Allow-Origin är *. Tidigare var det just * + Allow-Credentials: true, vilket aldrig fungerade för cookies.
- Dev: appen på :3100 och API:t på :4100 = cross-origin, därför CORS_ALLOWED_ORIGINS=http://localhost:3100. I prod ligger appen och API:t på samma host = same-origin, CORS behövs inte.
- Inloggning: dev-login är ett formulär som postar till API:t (top-level navigation - ingen CORS). Zitadel-länken går direkt till /api/v1/auth/login. Callbacken låg redan hos API:t, så inget ändrades i OAuth-flödet.

Lärdom: en proxy är inte automatiskt svaret på 'cookien ska följa med'. Fråga först om API:t kan exponeras direkt - i prod låg appen och API:t redan på samma host, så proxyn var ren omkostnad. CORS-headrar med * och Allow-Credentials: true är ett vanligt fynd: det ser konfigurerat ut men blockeras i webbläsaren.

Verifierat: headless chromium mot en canary av API:t i dev-klustret - 0 anrop via /api/proxy, dev-login fungerade, skrivande POST (mål) med cookie och preflight gick igenom.

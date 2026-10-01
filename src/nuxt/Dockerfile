# ==============================
# Build Stage
# ==============================
FROM node:lts-alpine AS build
WORKDIR /app

RUN apk add --no-cache python3 make g++ gcc

# Installera Bun för att bygga och installera dependencies
RUN npm install -g bun

# Kopiera bara det som behövs för install
COPY package.json bun.lock* ./

# Installera beroenden (cache-vänligt)
RUN bun install --frozen-lockfile

# Kopiera källkod (efter install för bättre cache)
COPY . .

# Bygg applikationen
RUN bun run build

# ==============================
# Runtime Stage (Bun)
# ==============================
FROM oven/bun:1-slim AS production

WORKDIR /app

# Kör med Bun - mer stabil HTTP-hantering än Node.js vid långa upptider
USER bun

# Kopiera bara det som behövs för runtime
COPY --from=build --chown=bun:bun /app/.output .

EXPOSE 3000

ENV HOST=0.0.0.0
ENV PORT=3000
ENV NODE_ENV=production

# Bun hanterar HTTP-trafiken stabilare än Node.js vid långa upptider
# och har bättre minneshantering vid hög belastning
CMD ["bun", "run", "server/index.mjs"]

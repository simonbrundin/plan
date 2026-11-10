# ==============================
# Build Stage
# ==============================
# Node-versionen är pinnad: "lts" flyttade sig till Node 26, vilket gjorde
# att better-sqlite3 (ett auto-installerat peer-dependency) inte gick att
# kompilera - byggen föll i CI.
FROM node:24-alpine AS build
WORKDIR /app

RUN apk add --no-cache python3 make g++ gcc
# RUN apk add openssh g++ make python3 git


# Kopiera bara det som behövs för install
COPY package.json bun.lock* ./

# Installera Bun (globalt)
RUN npm install -g bun

# Installera beroenden (cache-vänligt)
# --omit=peer: peer-dependencies (t.ex. better-sqlite3) auto-installeras av
# bun och kräver en nativ kompilering som är skör över Node-versioner.
# Appen använder dem inte.
RUN bun install --frozen-lockfile --omit=peer

# Kopiera källkod (efter install för bättre cache)
COPY . .

# Bygg applikationen
RUN bun run build

# ==============================
# Runtime Stage (minimal, säker)
# ==============================
FROM node:24-alpine AS production

WORKDIR /app

# Kör Nitro med Node.js. Nitro använder preset node-server och Bun kan starta
# processen men inte alltid hantera HTTP-trafiken korrekt i produktion.
USER node

# Kopiera bara det som behövs för runtime
COPY --from=build --chown=node:node /app/.output .

EXPOSE 3000

ENV HOST=0.0.0.0
ENV PORT=3000

CMD ["node", "server/index.mjs"]

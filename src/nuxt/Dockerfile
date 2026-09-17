# ==============================
# Build Stage
# ==============================
FROM node:lts-alpine AS build
WORKDIR /app

RUN apk add --no-cache python3 make g++ gcc
# RUN apk add openssh g++ make python3 git


# Kopiera bara det som behövs för install
COPY package.json bun.lock* ./

# Installera Bun (globalt)
RUN npm install -g bun

# Installera beroenden (cache-vänligt)
RUN bun install --frozen-lockfile

# Kopiera källkod (efter install för bättre cache)
COPY . .

# Bygg applikationen
RUN bun run build

# ==============================
# Runtime Stage (minimal, säker)
# ==============================
FROM node:lts-alpine AS production

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

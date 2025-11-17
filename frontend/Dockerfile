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
FROM oven/bun:1-slim AS production

WORKDIR /app

# Använd inbyggd non-root 'bun'-användare
USER bun

# Kopiera bara det som behövs för runtime
COPY --from=build --chown=bun:bun /app/.output .
# Kopiera fullständiga node_modules från build (mer tillförlitligt än .output/server/node_modules)
# COPY --from=build --chown=bun:bun /app/node_modules ./node_modules

# Exponera port
EXPOSE 3000

# Sätt NODE_PATH så att Bun hittar node_modules från root
ENV NODE_PATH=/app/node_modules

# Startkommando
CMD ["bun", "server/index.mjs"]

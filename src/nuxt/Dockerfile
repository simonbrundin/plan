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

# Skapa icke-root användare
RUN addgroup -g 1001 nodejs && adduser -S nodejs -u 1001

# Kopiera bara det som behövs för runtime
COPY --from=build --chown=nodejs:nodejs /app/.output .

# Exponera port
EXPOSE 3000

# Byt till icke-root användare
USER nodejs

# Startkommando
CMD ["node", "server/index.mjs"]

# ==============================
# Build Stage
# ==============================
FROM --platform=$BUILDPLATFORM oven/bun:1-debian AS build
WORKDIR /app

COPY package.json bun.lockb* ./
RUN bun install --frozen-lockfile --production

COPY . .
RUN bun run build

# ==============================
# Production Stage
# ==============================
FROM oven/bun:1-debian AS production
WORKDIR /app

# Använd inbyggd non-root 'bun'-användare
USER bun

COPY --from=build --chown=bun:bun /app/node_modules /app/node_modules
COPY --from=build --chown=bun:bun /app/.output /app

EXPOSE 3000
CMD ["bun", "run", "server/index.mjs"]

# use the official Bun image
# see all versions at https://hub.docker.com/r/oven/bun/tags
FROM oven/bun:1 AS build
WORKDIR /app

COPY package.json bun.lockb ./

# use ignore-scripts and no-optional to avoid building native modules like better-sqlite3
RUN bun install --frozen-lockfile --ignore-scripts --no-optional

# Copy the entire project
COPY . .

RUN bun --bun run build

# copy production dependencies and source code into final image
FROM oven/bun:1 AS production
WORKDIR /app

# Copy node_modules for runtime dependencies
COPY --from=build /app/node_modules /app/node_modules

# Only `.output` folder is needed from the build stage
COPY --from=build /app/.output /app

# run the app
EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "--bun", "run", "/app/server/index.mjs" ]


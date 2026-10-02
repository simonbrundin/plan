FROM caddy:2

# Caddy admin API listens on :2019 and is forwarded by docker-compose.
# Caddyfile mounts dynamic routes from /etc/caddy/routes/ via `import`.
COPY Caddyfile /etc/caddy/Caddyfile
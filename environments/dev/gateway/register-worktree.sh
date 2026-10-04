#!/usr/bin/env bash
#
# Register or remove a worktree's dev server in the shared Caddy dev-gateway.
#
# Usage:
#   ./register-worktree.sh add <name> <nuxt_port> <api_port>
#   ./register-worktree.sh remove <name>
#   ./register-worktree.sh list
#
# The gateway runs on http://localhost:2019 (admin API).
# All worktrees share a single OAuth redirect URL:
#   https://localhost:3000/oauth/callback
#
# Routes:
#   /<wt>/*            -> :nuxt_port  (Nuxt frontend)
#   /oauth/callback    -> :api_port   (Go API, matched by w cookie)
#
# Requirements: curl, jq

set -euo pipefail

CADDY_API="${CADDY_API:-http://localhost:2019}"
GATEWAY_PORT="${GATEWAY_PORT:-3000}"
BASE_URL="https://localhost:${GATEWAY_PORT}"

action="${1:-}"
name="${2:-}"
nuxt_port="${3:-}"
api_port="${4:-}"

if [[ -z "$action" ]]; then
  echo "Usage: $0 <add|remove|list> [worktree-name] [nuxt_port] [api_port]" >&2
  exit 1
fi

add_route() {
  local wt="$1" nuxt="$2" api="$3"
  local cookie_value="wt=${wt}"

  # Remove any existing routes for this worktree first so re-registration is idempotent.
  curl -fsS -X DELETE "${CADDY_API}/id/wt-${wt}-oauth-callback" > /dev/null 2>&1 || true
  curl -fsS -X DELETE "${CADDY_API}/id/wt-${wt}-path-route"     > /dev/null 2>&1 || true
  curl -fsS -X DELETE "${CADDY_API}/id/wt-${wt}-path-root"      > /dev/null 2>&1 || true

  # Route 1a: /<wt>/* -> Nuxt (with path)
  local path_route
  path_route=$(cat <<EOF
{
  "@id": "wt-${wt}-path-route",
  "match": [{"path": ["/${wt}/*"]}],
  "handle": [{
    "handler": "reverse_proxy",
    "upstreams": [{"dial": "127.0.0.1:${nuxt}"}]
  }],
  "terminal": true
}
EOF
)
  curl -fsS -H "Content-Type: application/json" \
    -X POST "${CADDY_API}/config/apps/http/servers/srv0/routes" \
    --data "$path_route" > /dev/null

  # Route 1b: /<wt> (root without trailing slash) -> Nuxt
  local path_root
  path_root=$(cat <<EOF
{
  "@id": "wt-${wt}-path-root",
  "match": [{"path": ["/${wt}"]}],
  "handle": [{
    "handler": "reverse_proxy",
    "upstreams": [{"dial": "127.0.0.1:${nuxt}"}]
  }],
  "terminal": true
}
EOF
)
  curl -fsS -H "Content-Type: application/json" \
    -X POST "${CADDY_API}/config/apps/http/servers/srv0/routes" \
    --data "$path_root" > /dev/null

  # Route 2: /oauth/callback with Cookie: wt=<wt> -> Go API
  local oauth_route
  oauth_route=$(cat <<EOF
{
  "@id": "wt-${wt}-oauth-callback",
  "match": [
    {
      "path": ["/oauth/callback"],
      "header": {"Cookie": ["${cookie_value}"]}
    }
  ],
  "handle": [{
    "handler": "reverse_proxy",
    "upstreams": [{"dial": "127.0.0.1:${api}"}]
  }],
  "terminal": true
}
EOF
)
  curl -fsS -H "Content-Type: application/json" \
    -X POST "${CADDY_API}/config/apps/http/servers/srv0/routes" \
    --data "$oauth_route" > /dev/null

  echo "Registered worktree '${wt}': Nuxt=${nuxt} API=${api}"
  echo "  Browse:    http://localhost:3000/${wt}/"
  echo "  OAuth URL: http://localhost:3000/oauth/callback (requires cookie wt=${wt})"
}

remove_route() {
  local wt="$1"

  curl -fsS -X DELETE "${CADDY_API}/id/wt-${wt}-oauth-callback" > /dev/null || true
  curl -fsS -X DELETE "${CADDY_API}/id/wt-${wt}-path-route"     > /dev/null || true

  echo "Removed worktree '${wt}'."
}

list_routes() {
  curl -fsS "${CADDY_API}/config/apps/http/servers/srv0/routes" \
    | python3 -m json.tool 2>/dev/null \
    || curl -fsS "${CADDY_API}/config/apps/http/servers/srv0/routes"
}

case "$action" in
  add)
    [[ -z "$name" || -z "$nuxt_port" || -z "$api_port" ]] && {
      echo "Name, nuxt_port, and api_port required for 'add'" >&2
      exit 1
    }
    add_route "$name" "$nuxt_port" "$api_port"
    ;;
  remove)
    [[ -z "$name" ]] && { echo "Name required for 'remove'" >&2; exit 1; }
    remove_route "$name"
    ;;
  list)
    list_routes
    ;;
  *)
    echo "Unknown action: $action" >&2
    exit 1
    ;;
esac
#!/usr/bin/env bash
# Skapar och krypterar de hemligheter agenten behöver i prod.
#
# Agenten körs inte i prod ännu - deploymenten och databasen är borttagna
# i prod/patches. Det här skriptet skapar hemligheterna så att agenten kan
# sättas i drift utan att värdena hamnar i git.
#
# Värdena tas från miljövariabler, aldrig från argument - annars hamnar de
# i shell-historiken.
#
#   SOPS_AGE_KEY_FILE   fil med agentens age-nyckel
#   SOPS_RECIPIENT      age-mottagare (standard: repoets publika nyckel)
#   AGENT_PLAN_API_KEY  nyckel Plan använder mot agent-API:t
#   AGENT_WORKER_KEY    nyckel jobbet använder mot Plan (agent-scope)
#   AGENT_LLM_API_KEY   nyckel till modellleverantören
#   AGENT_GIT_HEADER    git-http-header, t.ex. "Authorization: Basic ..."
#   AGENT_DB_PASSWORD   lösenord för agentens databasroll i prod
#
# Om AGENT_WORKER_KEY saknas genereras en nyckel, och hashen skrivs till
# plan-api-key.sql. Den filen körs i Plan-databasen när Plan är igång.
#
# Kör:
#   SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt \
#   AGENT_PLAN_API_KEY=... AGENT_WORKER_KEY=... \
#   AGENT_LLM_API_KEY=... AGENT_GIT_HEADER=... AGENT_DB_PASSWORD=... \
#   ./create-agent-secrets.sh

set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Mottagaren kan överstyras vid nyckelrotation och vid test av skriptet.
recipient="${SOPS_RECIPIENT:-$(yq -r '.creation_rules[0].age' "$here/../../base/sops-public-key.yaml")}"

require() {
  if [[ -z "${!1:-}" ]]; then
    echo "FEL: miljövariabeln $1 saknas" >&2
    exit 1
  fi
}

require AGENT_PLAN_API_KEY
require AGENT_LLM_API_KEY
require AGENT_GIT_HEADER
require AGENT_DB_PASSWORD

if [[ -z "${SOPS_AGE_KEY_FILE:-}" && -z "${SOPS_AGE_KEY:-}" ]]; then
  echo "FEL: sätt SOPS_AGE_KEY_FILE (eller SOPS_AGE_KEY) med agentens age-nyckel" >&2
  exit 1
fi

# Värden kan innehålla kolon och mellanslag - sätt dem i enkla citattecken
# så att YAML-parseraren inte misstar dem för nyckel-värde-par.
yaml_quote() {
  printf "'%s'" "${1//\'/\'\'}"
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Plan sparar bara hashen av en API-nyckel. Nyckeln måste därför skapas
# här och hashas på samma sätt som Plan gör, så att nyckeln kan skrivas
# in i databasen senare.
#
# Nyckeln består av 32 alfanumeriska tecken, precis som Plan genererar.
# cut läser hela strömmen - en pipe som avbryts i förväg ger ett
# SIGPIPE-fel när skriptet körs med pipefail.
generate_api_key() {
  openssl rand -base64 48 | tr -dc 'a-zA-Z0-9' | cut -c1-32
}

api_key_hash() {
  printf '%s' "$1" | sha256sum | cut -d' ' -f1
}

worker_key_generated=false
if [[ -z "${AGENT_WORKER_KEY:-}" ]]; then
  AGENT_WORKER_KEY="$(generate_api_key)"
  worker_key_generated=true
fi

# Namnen måste stämma med dem i environments/base/agent-deployment.yaml:
# jobbcontrollern läser hemligheterna därifrån.
cat > "$tmp/agent-api-key.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: plan-agent-api-key
  namespace: plan
  labels:
    app: agent-worker
type: Opaque
stringData:
  PLAN_API_KEY: $(yaml_quote "$AGENT_WORKER_KEY")
  AGENT_API_KEY: $(yaml_quote "$AGENT_PLAN_API_KEY")
EOF

cat > "$tmp/agent-llm.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: plan-agent-llm
  namespace: plan
type: Opaque
stringData:
  MINIMAX_API_KEY: $(yaml_quote "$AGENT_LLM_API_KEY")
EOF

cat > "$tmp/agent-git.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: plan-agent-git
  namespace: plan
type: Opaque
stringData:
  GIT_AUTH_HEADER: $(yaml_quote "$AGENT_GIT_HEADER")
EOF

cat > "$tmp/agent-db.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: plan-agent-db-app
  namespace: plan
type: kubernetes.io/basic-auth
stringData:
  username: agent
  password: $(yaml_quote "$AGENT_DB_PASSWORD")
  uri: $(yaml_quote "postgres://agent:${AGENT_DB_PASSWORD}@plan-agent-db-rw.plan.svc.cluster.local:5432/agent?sslmode=require")
EOF

for name in agent-api-key agent-llm agent-git agent-db; do
  sops --encrypt \
    --age "$recipient" \
    --encrypted-regex '^(data|stringData)$' \
    "$tmp/$name.yaml" > "$here/$name.enc.yaml"
  echo "krypterade $here/$name.enc.yaml"
done

if [[ "$worker_key_generated" == true ]]; then
  cat > "$here/plan-api-key.sql" <<EOF
-- Körs i Plan-databasen (prod) när Plan är igång.
-- Nyckeln finns krypterad i agent-api-key.enc.yaml - den här filen
-- innehåller bara hashen, som är allt Plan behöver.
INSERT INTO api_keys (name, key_hash, user_id, scope)
VALUES ('agent-worker', '$(api_key_hash "$AGENT_WORKER_KEY")',
        (SELECT id FROM users ORDER BY id LIMIT 1), 'agent')
ON CONFLICT (key_hash) DO NOTHING;
EOF
  echo "skrev $here/plan-api-key.sql - kör den i Plan-databasen när Plan är igång"
fi

cat <<'EOF'

Klart. Lägg till filerna i prod/secrets/kustomization.yaml:

  - agent-api-key.enc.yaml
  - agent-llm.enc.yaml
  - agent-git.enc.yaml
  - agent-db.enc.yaml

och ta bort patcharna

  - path: patches/agent-deployment.yaml
  - path: patches/agent-db-patch.yaml

ur prod/kustomization.yaml när agenten ska sättas i drift.
EOF

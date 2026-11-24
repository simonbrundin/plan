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
require AGENT_WORKER_KEY
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

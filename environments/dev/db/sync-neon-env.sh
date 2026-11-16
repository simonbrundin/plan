#!/usr/bin/env bash
# Genererar environments/dev/pgadmin/pgadmin-values.yaml ur neon.env och
# pgadmin-values.base.yaml.
#
#   ./sync-neon-env.sh
#
# Kör detta när Neon-credentials eller endpoint ändras. Resultatet gitignoras.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
env_file="${here}/neon.env"
base_file="${here}/../pgadmin/pgadmin-values.base.yaml"
out_file="${here}/../pgadmin/pgadmin-values.yaml"

if [[ ! -f "${env_file}" ]]; then
  echo "FEL: ${env_file} saknas - kopiera neon.env.example till neon.env" >&2
  exit 1
fi

read_env() {
  # Läser en nyckel ur neon.env (sista förekomsten vinner).
  local key="$1"
  grep -E "^${key}=" "${env_file}" | tail -n 1 | cut -d= -f2-
}

uri="$(read_env uri)"
if [[ -z "${uri}" ]]; then
  echo "FEL: 'uri' saknas i ${env_file}" >&2
  exit 1
fi

# Bryt ner URI:n: scheme://user:password@host:port/db?params
rest="${uri#*://}"
userinfo="${rest%@*}"
hostpart="${rest#*@}"
credentials="${userinfo%%:*}"
password="${userinfo#*:}"
hostport="${hostpart%%/*}"
db_and_params="${hostpart#*/}"
dbname="${db_and_params%%\?*}"
host="${hostport%%:*}"
port="${hostport#*:}"
[[ "${port}" == "${host}" ]] && port=5432

# YAML-skalär - dubbla citattecken måste escaped.
esc() { printf '%s' "$1" | sed 's/"/\\"/g'; }

cat > "${out_file}" <<EOF
# Genererad av environments/dev/db/sync-neon-env.sh - gitignoras.
$(cat "${base_file}")
# Serverdefinition för Neon i pgAdmin.
serverDefinitions:
  enabled: true
  resourceType: ConfigMap
  servers:
    "neon":
      Name: "neon"
      Group: "Neon"
      Host: "$(esc "${host}")"
      Port: ${port}
      Username: "$(esc "${credentials}")"
      Password: "$(esc "${password}")"
      MaintenanceDB: "$(esc "${dbname}")"
      ConnectionParameters:
        sslmode: "require"
EOF

echo "Skrev ${out_file} (${host}:${port}/${dbname})"

---
title: Lösning: 1Password-backed env-variabel för MCP-tokens
type: note
tags: [säkerhet, dotfiles, mcp, 1password, projekt:simon]
---

Efter att Neon-token hamnat i git-historiken i dotfiles (privat repo) roterades nyckeln: gammal nyckel (id 3491077) revokerad via console.neon.tech/api/v2/api_keys, nyckel 'Neon API key' (op://Private/Neon API key/credential) i 1Password. Mönster som används nu: config-filer (pi mcp.json, opencode.json, goose, gemini) refererar till $NEON_API_KEY / {env:NEON_API_KEY} / ${NEON_API_KEY} i stället för klartext. Shell-init laddar hemligheten från 1Password och cachar den i ~/.config/secrets/<namn> (600) så att den fungerar även när 1Password är låst. Nushell: nushell/.config/nushell/vendor/autoload/secrets.nu (laddas från config.nu, kommandot 'secrets refresh'). Bash: bash/.config/secrets/load.sh (laddas från .bashrc FÖRE interaktivitetskontrollen så icke-interaktiva shells också får variabeln). Viktigt: Neon Console-API flyttade från api.neon.tech (A-record finns inte längre) till console.neon.tech/api/v2. Övriga tokens i repot är kvar i klartext (MiniMax-nyckel i nushell env.nu, GitHub PAT i .mac/mcphub/servers.json) - samma sanering borde göras.

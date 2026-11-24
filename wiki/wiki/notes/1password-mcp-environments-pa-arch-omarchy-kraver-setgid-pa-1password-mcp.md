---
title: 1Password MCP (Environments) på Arch/Omarchy kräver setgid på 1password-mcp
type: note
tags: [1password, mcp, arch, omarchy, security, projekt:tmp]
created: 2026-11-23
updated: 2026-11-23
status: active
related: [notes/losning-1password-backed-env-variabel-for-mcp-tokens]
---

1Password:s desktop-app exponerar Environments-MCP via en IPC-relay ( abstrakt unix-socket @1PASSWORD_MCP ) som bara startar när 'Settings > Labs > Enable local MCP server' är på (setting-nyckeln i ~/.config/1Password/settings/settings.json är 'developers.mcpIntegration.enabled'; värdet MÅSTE signeras av appen - direkt redigering av filen återställs: 'setting ... had an untrusted value'). Relay-klienten /opt/1Password/1password-mcp verifierar med SO_PEERCRED att peer-processens effektiva GID är gruppen 'onepassword-mcp'. AUR-paketet 1password (Arch) kör INTE vendor-skriptet after-install.sh och skippar därför setupen av gruppen + setgid-bit => appen loggar 'Rejecting MCP connection: Linux peer effective GID check failed' och MCP-serven startar aldrig. Fix (root): groupadd onepassword-mcp; chgrp onepassword-mcp /opt/1Password/1password-mcp; chmod g+s /opt/1Password/1password-mcp; (valfritt) ln -sf /opt/1Password/1password-mcp /usr/bin/1password-mcp och /opt/1Password/onepassword-mcp. VIKTIGT: pacman-uppgradering ersätter binären och tappar setgid - applicera om efter varje uppgradering (t.ex. pacman-hook i /etc/pacman.d/hooks). Felsökning: loggen för MCP-klienten ligger i ~/.config/1Password/logs/MCP/1Password_rCURRENT.log; appens IPC-start loggas i ~/.config/1Password/logs/1Password_rCURRENT.log ('MCP integration is enabled, starting IPC server').

**LÖST 2026-11-23 (maskinen):** gruppen 'onepassword-mcp' (gid 1004) skapades, /opt/1Password/1password-mcp fick chgrp + setgid (g+s), /usr/bin/1password-mcp symlinkades. Verifierat: `authenticate` -> {"account_id":"CAKJ6FY2DBE5JFEP6RIKD63WTE"} och `list_environments` -> {"environments": []}. VIKTIGT: `authenticate` MÅSTE köras först i varje session - annars svarar övriga tools med 'Missing required scope: environments:list. Call the authenticate tool first.' (scoperna sätts på authenticate-svaret). 1Password-accounts userId (RNKLPRX6NFH47EEMXD6JITAZFU) är INTE accountId - accountId = CAKJ6FY2DBE5JFEP6RIKD63WTE.

**pacman-hook (INSTALLERAD 2026-11-23):** /etc/pacman.d/hooks/1password-mcp.hook - PostTransaction på paketet '1password' som kör groupadd/chgrp/chmod g+s + symlänkar, så att setgid inte försvinner vid uppgradering.

**pi-konfig:** ~/.pi/agent/mcp.json har servern '1password' -> command '/opt/1Password/1password-mcp', exposure 'direct' (8 tools). Symlink ~/.local/bin/1password-mcp -> /opt/1Password/1password-mcp för andra MCP-klienter.

## Relaterade

- [[notes/losning-1password-backed-env-variabel-for-mcp-tokens]]


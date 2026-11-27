---
title: Agentens kodändringsflöde portat till Rust
type: note
tags: [rust, agent, workspace, parity, projekt:plan]
---

Agentens kodändringsflöde är nu portat till Rust: workspace, executor,
proposer och coder. Flödet verifierades live mot Go med samma Plan-API
och samma databas - händelsesekvensen i körningens historik var
identisk.

**Varför detta läggs upp:** kodändringar är agentens farligaste del. Den
ska aldrig skriva direkt i det repo som utvecklas, så Rust-arbetet gör
ändringen i en isolerad kopia (git clone eller filkopiering), diffar
den, committar på en egen branch och skickar en Pull Request. Huvudgrenen
rörs aldrig - en människa mergear.

**Lärdom (defineraren):** modellens svar kan innehålla `null` för fält
den inte har något till. Go lämnar sådana fält oförändrade när
`json.Unmarshal` körs, men Serde avvisar dem. Rust lärde sig därför att
läsa `null` som tomma värden i `Definition` (null_som_tom_str,
null_som_false, null_som_noll, null_som_tom_lista). Utan det föll
definitionen över på textformatet och agenten kunde inte definiera något
mål alls - felet var "svaret innehåller ingen titel".

**Lärdom (felmeddelanden):** `gh` och `git` anropas olika i Go - `gh`
körs direkt och ger "failed to create pull request: ...", medan git går
genom workspace-hjälpen och ger "command \"git\" failed: ...". Rust
portade båda formateringen så att händelsehistoriken blir jämförbar
mellan Go och Rust.

**Lärdom (workern återupptar en misslyckad körning):** workern återanvänder
den senaste körningen för ett mål, även om den redan är misslyckad. Då
går livscykeln i Plan att neka övergången failed -> waiting_approval, och
workern loggar felet i varje poll. Det är Go-paritet (Go gör likadant),
men det gör målet till en oändlig retry - ett nytt mål behövs för att
testa vidare.

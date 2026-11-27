---
title: Kunskaps-wikin och dess verktyg har portats till Rust
type: note
tags: [wiki, rust, migrering, agent, verktyg, projekt:plan]
---

---
title: Kunskaps-wikin och dess verktyg har portats till Rust
tags: [wiki, rust, migrering, agent, verktyg, projekt:plan]
---

Wikin är fortfarande filbaserad, men verktyget som skriver den är nu
implementerat i Rust i stället för i Go. Paketet `internal/wiki` i
agenten ersatts av crate:t `plan-wiki` i `src/rust/crates/plan-wiki`, med
samma CLI: `wikictl`.

## Varför

Migreringen av agenten från Go till Rust gick igenom paketen ett för ett.
Wiken låg kvar i Go som det sista verktyget, vilket gjorde att
"migreringen klar" inte kunde sägas förrän även den flyttat. Att ha två
implementationer av samma filformat är dessutom en felkälla: den ena
skriver en sida och den andra läser den.

## Vad som gjordes

- `plan-wiki` blev ett eget crate med biblioteket och binären `wikictl`.
- Alla Go-testerna i `internal/wiki` portades till `tests/wiki.rs` - 37
  tester som går samma fall: slugging, frontmatter, sökning, backlinks,
  ingestion, lint, råkällor och `AGENTS.md`-blocket.
- `scripts/compare-wiki-parity.sh` jämför Go- och Rust-versionen
  kommando för kommando: läsande kommandon mot den riktiga wikin,
  skrivande kommandon mot kopior av en testwiki där båda resultaten
  jämförs fil för fil.

## Vad pariteten krävde

- Go formaterar filfel som `open filnamn: no such file or directory`.
  Rust ger `No such file or directory (os error 2)`. Skillnaden syns i
  felmeddelandena, så errno-texterna mappas explicit.
- `filepath.Join` städar bort `./` i sökvägar - `wikictl init .` ska
  därför skriva `AGENTS.md`, inte `./AGENTS.md`.
- `index.md` har en egen form i Go: en typ-tabell först och sedan en
  avdelning per sidtyp, med sidor sorterade på titel.
- Skrivning till en stängd pipe får inte panika - `wikictl list | head`
  annars avbryter med en panik i Rust, medan Go bara ignorerar felet.

## En fälla i `wikictl init`

`init` ersätter **hela** wikinblocket mellan `<!-- wiki:start -->` och
`<!-- wiki:end -->`. Plan-repot har ett eget block ("Ägande av
kunskaps-wikin") som beskriver vem som äger wikin - det blocket skrivs
över av det generiska. Kör därför `init` i en kopia, eller läs tillbaka
`AGENTS.md` från git efteråt. Det hände en gång under porten och
röktes av `git checkout -- AGENTS.md`.

## Relaterade

- [[notes/kunskaps-wiki]]
- [[notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det]]
- [[notes/agentens-kodandringsflode-portat-till-rust]]

---
title: Målens id:n är INTEGER - och tysta SQLx-fel är farliga
type: note
tags: [rust, bug, sqlx, postgres, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Under Rust-migreringen av Plan-API:t upptäcktes en felaktig cykelkontroll som
inte hittade cykler. Orsaken: `goal_dependencies.depends_on_id` är INTEGER i
schemat, men frågan läste den som `i64`. SQLx kunde inte avkoda värdet, och
felet hanterades med `let Ok(barn) = barn else { continue }` - vilket gjorde att
grenen tyst hoppades över. Cykelkontrollen svarade alltså alltid "ingen cykel",
precis som om logiken vore rätt.

**Lärdom:** i kod som bär på ett beslut (till exempel "detta beroende bildar en
cykel") får ett databasfel aldrig översättas till "nej, allt är bra". Ett fel
ska antingen bubbla upp som 500 eller loggas som varning - aldrig tyst
hoppas över. Rusts `let ... else` gör detta lätt att göra fel; därför är
mönstret nu: `else { tracing::warn!(...); continue }` i stället för en tyst
`continue`.

Den andra lärdomen gäller SQLx: `QueryBuilder` sätter ihop argumenten utan
mellanrum, så varje fält i en dynamisk UPDATE måste börja med ett blanksteg och
komman mellan fälten måste skrivas manuellt. En fråga som
"UPDATE goals SET title = $1 due = $2" är ett syntaxfel i Postgres, och den
syntesen fångades först när frågan kördes mot en riktig databas.

Båda fynden är av samma sort: ett fel som inte ger ett fel, utan ett
saknat resultat. Paritetsskriptet (scripts/compare-api-parity.sh) upptäckte
inte cykelfelet - det krävdes ett anrop som jämförde Go och Rust på samma
cykelseende för att skillnaden syntes.

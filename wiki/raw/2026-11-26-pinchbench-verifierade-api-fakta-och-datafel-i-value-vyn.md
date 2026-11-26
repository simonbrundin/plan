---
title: PinchBench: verifierade API-fakta och datafel i Value-vyn
type: source
added: 2026-11-26
origin: https://pinchbench.com/?view=value
immutable: true
---

# PinchBench - verifierade fakta fran det publika API:t

Kalla: https://pinchbench.com/ , https://api.pinchbench.com , https://github.com/pinchbench/api
Verifierat: 2026-11-26

## Vad det ar
Kilo Codes benchmark for LLM:er som agenter i OpenClaw. v2.0.0 = 148 tasks, 11 kategorier.
Grading: automated / llm_judge (Haiku default i v2) / hybrid. Semver-versionering med
"Current"-badge, historik bevarad. Official-korningar vs community-korningar markeras.
Tasks + grading-kriterier ar open source (github.com/pinchbench/skill).

## API-endpoints (offentliga)
GET /api/leaderboard?official=true&version=2.0.0&limit=200
  -> per modell: best_score_percentage, average_score_percentage, average/best_cost_usd,
     average/best_execution_time_seconds, submission_count, best_submission_id, weights, hf_link
GET /api/models
GET /api/stats
GET /api/benchmark_versions  och  /api/benchmark_versions/latest
GET /api/submissions?official=true&limit=N&offset=N  -> lista
GET /api/submissions/:id  -> enskild korning med per-uppgift-data:
   tasks[]: task_id, score, max_score, grading_type, timed_out, execution_time_seconds, breakdown, frontmatter
   usage_summary: total_input_tokens, total_output_tokens, total_requests, total_cost_usd
   metadata.suite ("all" = fullsuite | "task_xxx" = delkorning)
   metadata.system, openclaw_version, benchmark_version, official, run_id

## Verifierade datafel i Value-vyn
Value Score = Success % / Best Cost.
- "best"-vardena kommer från två olika korningar.
- Kostnaden ar inte normaliserad per uppgift, sa delkoringar (1 uppgift = ~1/147 av kostnaden)
  ser extremt billiga ut.
- CPST-kolumnen (Cost Per Successful Task) ar "N/A" for alla rader.

Exempel (verifierat via API):
google/gemma-4-26b-a4b-it
  best_score 100 % kommer fra submission 049af669-... som har 1 task (community, ej official).
  Officiell fullkorning (147 tasks): score 77.06 %, kostnad $2.0304, losta 113.28 -> 55.8 losta/$.
  PinchBench Value Score: 173.1  => ~3x for hojt (best fran olika korningar).
deepseek/deepseek-v4-flash
  official 147 tasks: 91.46 %, $2.2400, losta 134.45 -> 60.0 losta/$.
stepfun/step-3.5-flash
  official 147 tasks: 84.73 %, $1.1604, losta 124.55 -> 107.4 losta/$.
openai/gpt-oss-120b
  official 147 tasks: 47.45 %, $0.6531, losta 69.75 -> 106.8 losta/$.

## Poang är inte pass rate
Graderingen ger delpoang (t.ex. 0.93 for en llm_judge-uppgift). Leaderboardprocenten ar
medel av task-poang, inte andel helt lyckade uppgifter. BenchClaw (benchclaw.io/pinchbench,
2026-11-18) dokumenterar detta och att README (53 tasks), About-sidan (147) och
v2.0.0-release notes (148) inte overensstammer.

## Att notera
- Vendor-benchmark (Kilo Code) med incitament att lyfta sina 500+ modeller.
- Harness-bound (OpenClaw) - resultat ar inte jamforbara med t.ex. mini-swe-agent/Codex.
- Judge byttes over tid (Opus -> Haiku default i v2), alltsa domar-drift mellan versioner.

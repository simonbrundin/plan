---
title: Dashboard: budget-finder och prenumerationsjamforelse
type: note
tags: [dashboard, feature, pricing, projekt:dashboard]
---

Användarens egentliga behov: 'billigaste modellen som är tillräckligt smart för kod' + jämföra prenumerationsmodeller. Två nya paneler på models-sidan: (1) ModelsBudgetFinder - välj källa (AA Coding Agent/TB 4.1/TB 4.0/DeepSWE/uppskattad Intelligence) och referensmodell; listan visar modeller som poängsätts >= referensen på SAMMA källa, billigast först; mätta kostnader och token-estimat (≈) hålls i separata grupper och savings beräknas bara inom samma kostnadstyp. (2) ModelsSubscriptionComparison - SUBSCRIPTION_PLANS i app/data/subscriptions.ts (Claude Pro/Max/Ultra, ChatGPT Plus, Copilot Pro/Pro+/Max verifierade 2026-11-26 mot leverantörernas prissidor; Google AI Pro/Ultra markerade overifierade) jämförs mot AA:s mätta agentkostnad/task för planens modeller: break-even = pris/mån / $/task, effektiv $/task = pris/(uppg/dag * 30) med slider 1-60 uppg/dag. Användningsgränser modelleras inte - står i metodiken (ny sektion 7/8). Verifierat utfall: Sonnet-nivån (1.40/task) slås av Gemini 4 Neon (-40%, .86) och Fable 5.2 High (-34%); Copilot Pro är billigast per task vid 5 uppg/dag (/bin/bash.067 eff). 33/33 tester, lint, typecheck (utom gamla talos-fel), build och Tilt Ready.

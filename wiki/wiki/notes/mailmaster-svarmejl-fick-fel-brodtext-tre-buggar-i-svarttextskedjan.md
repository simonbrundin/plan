---
title: MailMaster: svarmejl fick fel brödtext - tre buggar i svarttextskedjan
type: note
tags: [bug, mailmaster, ai, mcp, projekt:simon]
created: 2026-10-29
updated: 2026-11-17
status: active
related: [notes/mailmaster-granskning-av-ai-atgarder-fore-utforande]
---

Symtom: MailAgent skickade svarmejl men brödtexten var ofta fel (t.ex. bara ordet 'Viktigt' från motiveringen).

Kedjan som gick sönder (AI-svaret -> parser -> executor):
1. normalizeAction (ai/response.go) kunde inte tolka 'ÅTGÄRD: reply, archive' (kommalista) -> action blev 'uncertain'.
2. När action var 'uncertain' kastades det strukturerade svaret bort och parseMotivationForAction() byggde ett nytt svar ur MOTIVERING-texten. Då försvann SVAR-texten (ActionArgs['reply']).
3. extractReplyFromContent() hade strings.Split(replyText, "")[0] - delar på tom streng -> gav bara FÖRSTA TECKNET i svaret (skulle vara '\n').
4. Executor.extractReplyText() föll tillbaka på citattext i motiveringen -> skickade 'Viktigt' (nyckelordet i citat) som brödtext.

Fix: parseActions() stöder kommaseparerade åtgärder (läggs i ActionArgs['actions']); strukturerade argument har företräde framför motivationssniffning; citattext accepteras bara om den är minst två ord; reply/forward-text som inte finns ger ett tydligt fel istället för skräppost; forward-text hämtas från citattext i motiveringen om SVAR saknas; alla åtgärder körs även om en misslyckas (errors.Join).

Regelfilens format stödjer nu: '### Regeln: ...' (H3), '1. Svara med: "text"' och 'Vidarebefordra till X med: {json}'.

Viktigt: regler.md ändrades av användaren under sessionen (5 regler). AI-svaret för 'viktigt'-regeln är ÅTGÄRD: reply, archive + SVAR: <text> - och det parsas nu till exakt rätt svarttext.

MCP-verktygen: alla 7 verktyg är registrerade korrekt (list_emails, read_email, archive_email, reply_to_email, forward_email, move_email, process_emails_with_ai). Standardmappen är nu kontots konfigurerade mapp (INBOX.Simon) i stället för INBOX.

## Relaterade

- [[notes/mailmaster-granskning-av-ai-atgarder-fore-utforande]]


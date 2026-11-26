---
title: ZITADEL: importerad användare krävde init-code-event för att kunna logga in
type: note
tags: [zitadel, bug, lesson, projekt:simon]
---

Efter importen av en användare utan lösenord skapade ZITADEL eventet 'user.human.initialization.code.added', vilket sätter user state till INITIAL (projections.users14.state = 6) och auth.users3.init_required = true. Login v2 svarar då med 'Initial User not supported' och erbjuder ingen lösenordsinloggning - trots att CreateSession i loggen svarade code_0 (dvs. autentiseringen lyckades).

För att en importerad användare ska kunna logga in med lösenord krävs tre saker:
1. ett event 'user.human.password.changed' med payload {"encodedHash": <bcrypt>, "changeRequired": false} i eventstore.events2 (lösenordshashen finns bara i eventstore, inte i någon projection),
2. eventet 'user.human.initialization.check.succeeded' (sätter state ACTIVE och init_required false),
3. auth.users3 med password_set=true, password_init_required=false, init_required=false.

Hashen kan tas direkt från ZITADEL:s export med with_passwords: true, då stämmer exakt det gamla lösenordet. bcrypt (a$...) fungerar i ZITADEL v4.14 precis som admin-kontots hash (a4$...).

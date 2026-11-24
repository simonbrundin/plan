---
title: Escape i datum-popup måste hanteras före leader-listenern
type: note
tags: [bug, keyboard, ui, projekt:plan]
---

Datum-popupens Escape stängdes inte eftersom leader-keydown lyssnar i capture-fasen och konsumerade Escape innan popupens egen listener. Layouten wrapper nu leader-handlern: när datum-popupen är öppen stängs den direkt och stoppar eventet.

@natural-language
Feature: Tolka naturlig text till kalenderhändelser och mål

  @event
  Scenario: Enkel händelse utan datum
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte med Anna"
    Then the title should be "Möte med Anna"
    And the date should be "2026-03-10"
    And the time should be "18:00"
    And no time should be explicit
    And the mode should be "event" by default

  @goal
  Scenario: Ord som talar för en händelse blir en händelse
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Harry Potter Kalas"
    Then the title should be "Harry Potter Kalas"
    And the date should be "2026-03-10"
    And the time should be "18:00"
    And the mode should be "event" by default

  @goal
  Scenario: Kvällen som redan passerat föreslås imorgon
    Given the reference time is "2026-03-10T20:00:00"
    When the user writes "Harry Potter Kalas"
    Then the date should be "2026-03-11"
    And the time should be "18:00"
    And the mode should be "event" by default

  @goal
  Scenario: Ett mål utan händelseord förblir ett mål
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Simhallen bokad"
    Then the mode should be "goal" by default
    And no date should be found

  @event
  Scenario: Datum med "imorgon"
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte imorgon"
    Then the title should be "Möte"
    And the date should be "2026-03-11"
    And no time should be found

  @event
  Scenario: Datum med veckodag
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte torsdag"
    Then the title should be "Möte"
    And the date should be "2026-03-12"

  @event
  Scenario: Klockslag med "kl"
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte kl 14"
    Then the title should be "Möte"
    And the time should be "14:00"

  @event
  Scenario: Klockslag med kolon
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte kl 14:30"
    Then the title should be "Möte"
    And the time should be "14:30"

  @event
  Scenario: Heldag när datum anges utan klockslag
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Konferens 12/03"
    Then the title should be "Konferens"
    And the date should be "2026-03-12"
    And the event should be all day

  @event
  Scenario: Längd i minuter
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte kl 14 i 30 min"
    Then the title should be "Möte"
    And the start should be "2026-03-10T14:00:00"
    And the end should be "2026-03-10T14:30:00"

  @event
  Scenario: Längd i timmar
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte kl 14 i 1,5 h"
    Then the title should be "Möte"
    And the end should be "2026-03-10T15:30:00"

  @event
  Scenario: Upprepning varje dag
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Stretcha varje dag kl 7"
    Then the title should be "Stretcha"
    And the recurrence rule should be "FREQ=DAILY"

  @event
  Scenario: Upprepning varje vecka
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Rapport varje vecka"
    Then the title should be "Rapport"
    And the recurrence rule should be "FREQ=WEEKLY"

  @event
  Scenario: Upprepning varannan dag
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Gymma varannan dag"
    Then the title should be "Gymma"
    And the recurrence rule should be "FREQ=DAILY;INTERVAL=2"

  @goal
  Scenario: Ett återkommande mål i målläge
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Gymma varannan dag"
    Then the goal title should be "Gymma"
    And the recurrence rule should be "FREQ=DAILY;INTERVAL=2"

  @event
  Scenario: Varannan vecka med startdatum och intervall över midnatt
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Jobb varannan vecka 17.00 till 17.00 dagen efter med start mån 30 nov"
    Then the title should be "Jobb"
    And the date should be "2026-11-30"
    And the time should be "17:00"
    And the end should be "2026-12-01T17:00:00"
    And the recurrence rule should be "FREQ=WEEKLY;INTERVAL=2"

  @event
  Scenario: Upprepning på en viss veckodag
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Stretcha varje tisdag"
    Then the title should be "Stretcha"
    And the recurrence rule should be "FREQ=WEEKLY;BYDAY=TU"

  @event
  Scenario: Hela texten med datum och tid
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Lunch med Anna imorgon kl 12"
    Then the title should be "Lunch med Anna"
    And the date should be "2026-03-11"
    And the time should be "12:00"

  @event
  Scenario: Ord som antyder ett klockslag behålls i titeln
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Lunch med Anna"
    Then the title should be "Lunch med Anna"
    And the time should be "12:00"

  @event
  Scenario: Tomma inmatningen ger ingen händelse
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "   "
    Then the title should be ""
    And no date should be found

  @event
  Scenario: Upprepning utan datum har inget utskrivet datum
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Stretcha varje dag"
    Then the title should be "Stretcha"
    And the recurrence rule should be "FREQ=DAILY"
    And the date should not be explicit

  @event
  Scenario: Utskrivet datum markeras som explicit
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte imorgon"
    Then the date should be explicit

  @range
  Scenario: Tidsintervall med "mellan" och "och"
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Sömn varje dag mellan 23.00 och 07.00"
    Then the title should be "Sömn"
    And the recurrence rule should be "FREQ=DAILY"
    And the time should be "23:00"
    And the end should be "2026-03-11T07:00:00"

  @range
  Scenario: Tidsintervall med bindestreck
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte 14.00-15.00"
    Then the title should be "Möte"
    And the time should be "14:00"
    And the end should be "2026-03-10T15:00:00"

  @range
  Scenario: Tidsintervall med "från" och "till"
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Arbetsdag från 09.00 till 17.00"
    Then the title should be "Arbetsdag"
    And the time should be "09:00"
    And the end should be "2026-03-10T17:00:00"

  @range
  Scenario: Tidsintervall med "kl"
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Lunch kl 12-13"
    Then the title should be "Lunch"
    And the time should be "12:00"
    And the end should be "2026-03-10T13:00:00"

  @range
  Scenario: Tidsintervall som passerar midnatt
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Sömn 23.00-07.00"
    Then the end should be "2026-03-11T07:00:00"

  @range
  Scenario: Tidsintervall med timmar utan minuter
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte 14-15"
    Then the title should be "Möte"
    And the time should be "14:00"
    And the end should be "2026-03-10T15:00:00"

  @range
  Scenario: Ett antal är inte ett tidsintervall
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Läs 3-4 böcker"
    Then the title should be "Läs 3-4 böcker"
    And no time should be found

  @range
  Scenario: Ett datumintervall skapas som heldag
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte 12-13 mars"
    Then the title should be "Möte"
    And the event should be all day
    And the start should be "2026-03-12T00:00:00"
    And the end should be "2026-03-14T00:00:00"

  @range
  Scenario: Ett månadsnamn ensamt ger ett datum
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte 13 mars"
    Then the title should be "Möte"
    And the event should be all day
    And the start should be "2026-03-13T00:00:00"

  @mode
  Scenario: Text utan datum blir ett mål
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Läs bok"
    Then the mode should be "goal" by default

  @mode
  Scenario: Text med datum blir en händelse
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte imorgon"
    Then the mode should be "event" by default

  @mode
  Scenario: Text med klockslag blir en händelse
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte kl 14"
    Then the mode should be "event" by default

  @mode
  Scenario: Text med upprepning blir en händelse
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Stretcha varje dag"
    Then the mode should be "event" by default

  @mode
  Scenario: Målläge kan tvingas fram
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte imorgon"
    Then the mode should be "goal" when the mode is "goal"

  @mode
  Scenario: Händelseläge kan tvingas fram
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Läs bok"
    Then the mode should be "event" when the mode is "event"

  @icon
  Scenario: Ett emoji i början blir ikon
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "🏃 Löpning kl 18"
    Then the title should be "Löpning"
    And the icon should be "🏃"
    And the time should be "18:00"

  @icon
  Scenario: Ett emoji i slutet blir ikon
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Löpning kl 18 🏃"
    Then the title should be "Löpning"
    And the icon should be "🏃"

  @icon
  Scenario: Text utan emoji har ingen ikon
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Löpning kl 18"
    Then no icon should be found

  @icon
  Scenario: En emoji-sekvens är en enda ikon
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "👨‍👩‍👧 Familjetid"
    Then the title should be "Familjetid"
    And the icon should be "👨‍👩‍👧"

  @icon
  Scenario: Tecken som inte är emoji blir inte ikon
    Given the reference time is "2026-03-10T09:00:00"
    When the user writes "Möte © 2026"
    Then the title should be "Möte © 2026"
    And no icon should be found

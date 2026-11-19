Feature: Händelserader i den vertikala kalendern
  Som användare vill jag att varje händelse syns som en läsbar rad med ikon
  och titel, oavsett hur mycket tid som syns. Jag vill aldrig behöva gissa
  vad som ligger i kalendern bara för att tidslinjen är utzoomed.

  Scenario: Tidsaxeln följer tidsfönstrets storlek
    Given the visible duration is 6 hours
    Then the scale level should be "hour"

  Scenario: Ett dygn räcker fortfarande som timmar
    Given the visible duration is 24 hours
    Then the scale level should be "hour"

  Scenario: Två dygn blir dagar
    Given the visible duration is 3 days
    Then the scale level should be "day"

  Scenario: En månad räcker fortfarande som dagar
    Given the visible duration is 30 days
    Then the scale level should be "day"

  Scenario: Två månader blir månader
    Given the visible duration is 60 days
    Then the scale level should be "month"

  Scenario: Två år blir år på axeln
    Given the visible duration is 800 days
    Then the scale level should be "year"

  Scenario: Fyra år blir år
    Given the visible duration is 1500 days
    Then the scale level should be "year"

  Scenario: En rad ligger på sin tid i vyn
    Given the calendar view is 600 pixels tall
    And the following calendar events exist:
      | title | start | end   |
      | Fika  | 12:00 | 13:00 |
    When the events are laid out
    Then the first row should start 300 pixels from the top
    And the first row should be 25 pixels tall

  Scenario: En kort händelse får ändå en läsbar rad
    Given the calendar view is 600 pixels tall
    And the visible duration is 30 days
    And the following calendar events exist:
      | title | start | end   |
      | Fika  | 12:00 | 12:30 |
    When the events are laid out
    Then the first row should be at least 16 pixels tall

  Scenario: En lång händelse får sin fulla höjd
    Given the calendar view is 600 pixels tall
    And the following calendar events exist:
      | title | start | end   |
      | Fika  | 09:00 | 11:00 |
    When the events are laid out
    Then the first row should be 50 pixels tall

  Scenario: Överlappande rader hamnar under varandra
    Given the calendar view is 600 pixels tall
    And the following calendar events exist:
      | title | start | end   |
      | Fika  | 10:00 | 11:00 |
      | Lunch | 10:00 | 11:00 |
    When the events are laid out
    Then the rows should not overlap

  Scenario: Alla rader ligger i samma kolumn
    Given the calendar view is 600 pixels tall
    And the following calendar events exist:
      | title | start | end   |
      | Fika  | 10:00 | 11:00 |
      | Lunch | 10:00 | 11:00 |
    When the events are laid out
    Then the rows should not overlap
    And the rows should be 2

  Scenario: Raderna behåller sin ordning i tiden
    Given the calendar view is 600 pixels tall
    And the following calendar events exist:
      | title    | start | end   |
      | Senare   | 12:00 | 13:00 |
      | Tidigare | 01:00 | 02:00 |
    When the events are laid out
    Then the first row should start 300 pixels from the top
    And the second row should start 25 pixels from the top

  Scenario: Rader som inte ryms i vyn markeras
    Given the calendar view is 100 pixels tall
    And the following calendar events exist:
      | title  | start | end   |
      | Fika 1 | 12:00 | 12:30 |
      | Fika 2 | 12:00 | 12:30 |
      | Fika 3 | 12:00 | 12:30 |
      | Fika 4 | 12:00 | 12:30 |
      | Fika 5 | 12:00 | 12:30 |
      | Fika 6 | 12:00 | 12:30 |
      | Fika 7 | 12:00 | 12:30 |
      | Fika 8 | 12:00 | 12:30 |
    When the events are laid out
    Then 6 rows should be clipped
    And the rows should not overlap

  Scenario: En daglig serie döljs när den har fler förekomster än gränsen
    Given the daily visibility limit is 7
    And the following calendar series exist:
      | key   | count | daily |
      | daily | 8     | yes   |
    When the visibility is planned
    Then the series "daily" should be hidden

  Scenario: En daglig serie med få förekomster döljs inte
    Given the daily visibility limit is 7
    And the following calendar series exist:
      | key   | count | daily |
      | daily | 7     | yes   |
    When the visibility is planned
    Then no series should be hidden

  Scenario: Gränsen kan ändras
    Given the daily visibility limit is 3
    And the following calendar series exist:
      | key   | count | daily |
      | daily | 4     | yes   |
    When the visibility is planned
    Then the series "daily" should be hidden

  Scenario: En serie som var dold förblir dold vid gränsen
    Given the daily visibility limit is 7
    And the series "daily" was hidden before
    And the following calendar series exist:
      | key   | count | daily |
      | daily | 7     | yes   |
    When the visibility is planned
    Then the series "daily" should be hidden

  Scenario: Döljning kan stängas av
    Given the daily visibility limit is 7
    And hiding of recurring series is off
    And the following calendar series exist:
      | key   | count | daily |
      | daily | 30    | yes   |
    When the visibility is planned
    Then no series should be hidden

  Scenario: En serie kan visas ensam
    Given the daily visibility limit is 7
    And the series "weekly" is focused
    And the following calendar series exist:
      | key    | count | daily |
      | daily  | 8     | yes   |
      | weekly | 2     | no    |
    When the visibility is planned
    Then the series "daily" should be hidden
    And the series "weekly" should be visible

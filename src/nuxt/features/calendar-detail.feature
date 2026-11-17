Feature: Detaljnivåer i den vertikala kalendern
  Som användare vill jag att kalendern visar olika mycket information
  beroende på hur mycket tid som syns, så att timmar och dagar går att läsa
  i detalj medan månader och år går att överblicka.

  Scenario: Tidsnivåerna följer tidsfönstrets storlek
    Given the visible duration is 6 hours
    Then the scale level should be "hour"
    And the detail level should be "full"

  Scenario: Ett dygn räcker fortfarande som timmar
    Given the visible duration is 24 hours
    Then the scale level should be "hour"
    And the detail level should be "full"

  Scenario: Två dygn blir dagar
    Given the visible duration is 3 days
    Then the scale level should be "day"
    And the detail level should be "compact"

  Scenario: En månad räcker fortfarande som dagar
    Given the visible duration is 30 days
    Then the scale level should be "day"
    And the detail level should be "daily"

  Scenario: Två månader blir månader
    Given the visible duration is 60 days
    Then the scale level should be "month"
    And the detail level should be "monthly"

  Scenario: En vecka räcker som kompakta kort
    Given the visible duration is 5 days
    Then the scale level should be "day"
    And the detail level should be "compact"

  Scenario: Två veckor blir prickar per dag
    Given the visible duration is 14 days
    Then the scale level should be "day"
    And the detail level should be "daily"

  Scenario: Ett halvt år blir månader
    Given the visible duration is 180 days
    Then the scale level should be "month"
    And the detail level should be "monthly"

  Scenario: Två år blir månader
    Given the visible duration is 800 days
    Then the scale level should be "year"
    And the detail level should be "monthly"

  Scenario: Fyra år blir år
    Given the visible duration is 1500 days
    Then the scale level should be "year"
    And the detail level should be "yearly"

  Scenario: Flera år blir år
    Given the visible duration is 3650 days
    Then the scale level should be "year"
    And the detail level should be "yearly"

  Scenario: Nivån fladdrar inte vid gränsen mellan timmar och dagar
    Given the current detail level is "full"
    And the visible duration is 46 hours
    Then the detail level with hysteresis should be "full"
    Given the visible duration is 49 hours
    Then the detail level with hysteresis should be "full"
    Given the visible duration is 46 hours
    Then the detail level with hysteresis should be "full"

  Scenario: Nivån byts först när gränsen är klart passerad
    Given the current detail level is "full"
    And the visible duration is 53 hours
    Then the detail level with hysteresis should be "compact"

  Scenario: Nivån byts direkt när tiden blir kortare
    Given the current detail level is "compact"
    And the visible duration is 24 hours
    Then the detail level with hysteresis should be "full"

  Scenario: Nivån står kvar när tiden svänger kring en tidigare gräns
    Given the current detail level is "daily"
    And the visible duration is 20 days
    Then the detail level with hysteresis should be "daily"
    Given the visible duration is 30 days
    Then the detail level with hysteresis should be "daily"
    Given the visible duration is 20 days
    Then the detail level with hysteresis should be "daily"

  Scenario Outline: Händelser grupperas per dag
    Given the following calendar events exist:
      | title     | start                | end                  |
      | Fika      | 2027-03-12T09:00:00Z | 2027-03-12T10:00:00Z |
      | Lunch     | 2027-03-12T12:00:00Z | 2027-03-12T13:00:00Z |
      | Mötet     | 2027-03-14T09:00:00Z | 2027-03-14T10:00:00Z |
    When the events are grouped by day
    Then there should be 2 day groups
    And the first day group should contain 2 events
    And the second day group should contain 1 event

  Scenario: Antalet händelser visas i gruppens etikett
    Given the following calendar events exist:
      | title     | start                | end                  |
      | Fika      | 2027-03-12T09:00:00Z | 2027-03-12T10:00:00Z |
      | Lunch     | 2027-03-12T12:00:00Z | 2027-03-12T13:00:00Z |
    When the events are grouped by day
    Then the first day group should be titled "2 händelser"

  Scenario: En enda händelse använder entalsform
    Given the following calendar events exist:
      | title     | start                | end                  |
      | Fika      | 2027-03-12T09:00:00Z | 2027-03-12T10:00:00Z |
    When the events are grouped by day
    Then the first day group should be titled "1 händelse"

  Scenario: Händelser grupperas per månad
    Given the following calendar events exist:
      | title     | start                | end                  |
      | Fika      | 2027-03-12T09:00:00Z | 2027-03-12T10:00:00Z |
      | Lunch     | 2027-03-20T12:00:00Z | 2027-03-20T13:00:00Z |
      | Mötet     | 2027-04-01T09:00:00Z | 2027-04-01T10:00:00Z |
    When the events are grouped by month
    Then there should be 2 month groups
    And the first month group should contain 2 events
    And the second month group should contain 1 event

  Scenario: Grupper sorteras i tidsordning
    Given the following calendar events exist:
      | title     | start                | end                  |
      | Senare    | 2027-03-20T12:00:00Z | 2027-03-20T13:00:00Z |
      | Tidigare  | 2027-03-02T12:00:00Z | 2027-03-02T13:00:00Z |
    When the events are grouped by day
    Then the first day group should be titled "1 händelse"
    And the first day group should start before the second day group

  Scenario: Prickar som ligger nära varandra hamnar i olika kolumner
    Given markers are placed at 10, 12, 40 percent
    When the markers are packed into lanes with a minimum gap of 2 percent
    Then the marker lanes should be "0, 1, 0"

  Scenario: Prickar med tillräckligt avstånd kan dela kolumn
    Given markers are placed at 10, 30, 50 percent
    When the markers are packed into lanes with a minimum gap of 2 percent
    Then the marker lanes should be "0, 0, 0"

  Scenario: Kolumnerna återanvänds när det finns plats
    Given markers are placed at 10, 11, 12, 13 percent
    When the markers are packed into lanes with a minimum gap of 2 percent
    Then the marker lanes should be "0, 1, 2, 0"

  Scenario: Antalet kolumner begränsas
    Given markers are placed at 10, 11, 12, 13, 14, 15, 16 percent
    When the markers are packed into lanes with a minimum gap of 10 percent
    Then the marker lanes should be "0, 1, 2, 3, 4, 5, 5"

  Scenario: Intervall som överlappar hamnar i olika kolumner
    Given intervals are placed at 10-20, 12-14, 40-42 percent
    When the intervals are packed into lanes with a minimum gap of 2 percent
    Then the interval lanes should be "0/2, 1/2, 0/1"

  Scenario: Ett intervall som slutar sent blockerar sin kolumn
    Given intervals are placed at 10-40, 12-14 percent
    When the intervals are packed into lanes with a minimum gap of 2 percent
    Then the interval lanes should be "0/2, 1/2"

  Scenario: Intervall med tillräckligt avstånd får en egen kolumn
    Given intervals are placed at 10-20, 25-30, 40-42 percent
    When the intervals are packed into lanes with a minimum gap of 2 percent
    Then the interval lanes should be "0/1, 0/1, 0/1"

  Scenario: Händelser som överlappar delar bredden
    Given intervals are placed at 10-20, 12-14, 40-42 percent
    When the intervals are packed into lanes with a minimum gap of 2 percent
    Then the first event should use 2 columns
    And the third event should use 1 column

  Scenario: Händelser grupperas per år
    Given the following calendar events exist:
      | title     | start                | end                  |
      | Nyår      | 2027-01-01T09:00:00Z | 2027-01-01T10:00:00Z |
      | Sommar    | 2027-07-01T09:00:00Z | 2027-07-01T10:00:00Z |
      | Nyår 2028 | 2028-01-01T09:00:00Z | 2028-01-01T10:00:00Z |
    When the events are grouped by year
    Then there should be 2 year groups
    And the first year group should contain 2 events
    And the second year group should contain 1 event

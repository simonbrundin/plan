@recurring
Feature: Återkommande mål i listorna

  @recurring
  Scenario: Ett återkommande mål ligger kvar när förekomsten är slutförd
    Given the user has a recurring goal "Gymma" whose next occurrence is today
    And the user has a finished goal "Läs bok"
    When the occurrence of "Gymma" is completed
    Then the goal "Gymma" should still be listed
    And the goal "Läs bok" should be hidden from the list
    And the goal "Gymma" should have completed 1 occurrence

  @recurring
  Scenario: Ett återkommande mål väntar utgråat på nästa förekomst
    Given the user has a recurring goal "Gymma" whose next occurrence is tomorrow
    Then the goal "Gymma" should be waiting for its next occurrence

  @recurring
  Scenario: Ett återkommande mål med dagens förekomst går att checka av
    Given the user has a recurring goal "Gymma" whose next occurrence is today
    Then the goal "Gymma" should not be waiting for its next occurrence

  @recurring
  Scenario: Ett återkommande mål lämnar listan när serien tar slut
    Given the user has a recurring goal "Gymma" whose next occurrence is today
    When the series of "Gymma" ends
    Then the goal "Gymma" should be hidden from the list

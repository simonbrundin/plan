@priority
Feature: Markera och filtrera påbörjade mål på prioriteringssidan

  @started
  Scenario: Mark a goal as started
    Given the user is logged in
    And the user has a goal "Träna" that is not started
    When the user marks the goal "Träna" as started
    Then the goal "Träna" should be marked as started

  @started
  Scenario: Unmark a started goal
    Given the user is logged in
    And the user has a goal "Träna" that is started
    When the user unmarks the goal "Träna"
    Then the goal "Träna" should not be marked as started

  @filter
  Scenario: Filter to show only started goals
    Given the user is logged in
    And the user has the goals:
      | title   | started |
      | Träna   | yes    |
      | Läs bok | no     |
    And the filter is "Alla"
    When the user selects the filter "Påbörjade"
    Then only the goal "Träna" should be visible
    And the goal "Läs bok" should not be visible

  @filter
  Scenario: Filter to show only not-started goals
    Given the user is logged in
    And the user has the goals:
      | title   | started |
      | Träna   | yes    |
      | Läs bok | no     |
    And the filter is "Alla"
    When the user selects the filter "Ej påbörjade"
    Then only the goal "Läs bok" should be visible
    And the goal "Träna" should not be visible

  @filter
  Scenario: Filter "All" shows all goals
    Given the user is logged in
    And the user has the goals:
      | title   | started |
      | Träna   | yes    |
      | Läs bok | no     |
    And the filter is "Påbörjade"
    When the user selects the filter "Alla"
    Then both goals "Träna" and "Läs bok" should be visible

  @started
  Scenario: Started goals are marked correctly in the API
    Given the user is logged in
    And a goal has started value null
    When the user marks the goal as started
    Then the API should receive started value "<datum>"

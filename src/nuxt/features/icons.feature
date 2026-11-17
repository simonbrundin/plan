@icons
Feature: Hantera ikoner för kalenderhändelser

  En ikon kan vara antingen en emoji (t.ex. "🏃") eller ett Iconify-namn
  (t.ex. "heroicons:star"). Båda lagras i samma fält och måste gå att
  skilja på vid rendering.

  @emoji
  Scenario: Plocka ut ett emoji i början av texten
    Given the text is "🏃 Löpning kl 18"
    When the emoji is extracted
    Then the extracted emoji should be "🏃"
    And the remaining text should be "Löpning kl 18"

  @emoji
  Scenario: Plocka ut ett emoji i slutet av texten
    Given the text is "Löpning kl 18 🏃"
    When the emoji is extracted
    Then the extracted emoji should be "🏃"
    And the remaining text should be "Löpning kl 18"

  @emoji
  Scenario: En emoji-sekvens är en enda ikon
    Given the text is "👨‍👩‍👧 Familjetid"
    When the emoji is extracted
    Then the extracted emoji should be "👨‍👩‍👧"
    And the remaining text should be "Familjetid"

  @emoji
  Scenario: En flagga är en enda ikon
    Given the text is "🇸🇪 Sverige"
    When the emoji is extracted
    Then the extracted emoji should be "🇸🇪"
    And the remaining text should be "Sverige"

  @emoji
  Scenario: En hudtonmodifierad emoji är en enda ikon
    Given the text is "👍🏽 thumbs up"
    When the emoji is extracted
    Then the extracted emoji should be "👍🏽"
    And the remaining text should be "thumbs up"

  @emoji
  Scenario: Tekniska symboler är inte emoji
    Given the text is "Möte © 2026"
    When the emoji is extracted
    Then no emoji should be found
    And the remaining text should be "Möte © 2026"

  @emoji
  Scenario: Text utan emoji lämnas orörd
    Given the text is "Löpning kl 18"
    When the emoji is extracted
    Then no emoji should be found
    And the remaining text should be "Löpning kl 18"

  @search
  Scenario: Sökning i emoji-katalogen ger träffar
    Given the user searches for emoji "löpning"
    Then the emoji search should find "🏃"

  @search
  Scenario: Sökning i emoji-katalogen ger flera träffar
    Given the user searches for emoji "träning"
    Then the emoji search should find at least 2 emojis

  @search
  Scenario: Sökning utan träffar ger en tom lista
    Given the user searches for emoji "zzzzz"
    Then the emoji search should find no emojis

  @search
  Scenario: Tom sökning ger en tom lista
    Given the user searches for emoji ""
    Then the emoji search should find no emojis

  @iconify
  Scenario: Ett Iconify-namn känns igen
    Given the icon is "heroicons:star"
    Then it should be an Iconify icon

  @iconify
  Scenario: En emoji är inte ett Iconify-namn
    Given the icon is "🏃"
    Then it should not be an Iconify icon

  @iconify
  Scenario: Tomt värde är inte ett Iconify-namn
    Given the icon is ""
    Then it should not be an Iconify icon

  @iconify
  Scenario: Ikonen normaliseras genom att trimmas
    Given the icon is "  heroicons:star  "
    Then the normalized icon should be "heroicons:star"

  @iconify
  Scenario: En tom ikon normaliseras till en tom sträng
    Given the icon is ""
    Then the normalized icon should be ""

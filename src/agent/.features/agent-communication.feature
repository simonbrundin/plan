Feature: Kommunicera med agent via konversation
  Som användare
  Vill jag kunna skicka ett meddelande
  För att sätta agent i arbete

  Scenario: Användare kan skicka iväg meddelande till agent
    Given att användaren befinner sig på konversationssidan
    When användaren skrivit ett meddelande
    And användaren skickar meddelandet
    Then ska meddelandet gå iväg till agenten

  Scenario: Meddelande triggar workflow
    When användaren skrivit ett meddelande
    And användaren skickar meddelandet
    Then ska agenten ta emot meddelandet
    And köra workflow

  Scenario: Svar från agenten visas i konversationslistan
    Given att användaren befinner sig på konversationssidan
    When användaren skrivit ett meddelande
    And användaren skickar meddelandet
    And agenten svarar
    Then ska svaret visas för användaren

  Scenario: Användare får felmeddelande när agenten inte kan svara
    Given att användaren befinner sig på konversationssidan
    When användaren skrivit ett meddelande
    And användaren skickar meddelandet
    And agenten inte kan svara
    Then ska ett felmeddelande visas för användaren

  Scenario: Användare kan inte skicka tomt meddelande
    Given att användaren befinner sig på konversationssidan
    And att input-rutan är tom
    When användaren skickar meddelandet
    Then ska ett felmeddelande visas som säger att man ska skriva något
    And ska inget meddelande skickas


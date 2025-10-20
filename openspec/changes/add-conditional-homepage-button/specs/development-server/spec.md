## ADDED Requirements

### Requirement: Conditional Homepage Navigation
The homepage MUST display a conditional button that directs users to the appropriate page based on their authentication status. If the user is logged in, the button should navigate to the root-goals page. If the user is not logged in, the button should navigate to the login page.

#### Scenario: Logged-in user sees root-goals button
- **GIVEN** a user is logged in
- **WHEN** they visit the homepage
- **THEN** they see a button labeled "Go to Root Goals" or similar
- **AND** clicking the button navigates to `/root-goals`

#### Scenario: Non-logged-in user sees login button
- **GIVEN** a user is not logged in
- **WHEN** they visit the homepage
- **THEN** they see a button labeled "Login" or similar
- **AND** clicking the button navigates to `/login`

#### Scenario: Button styling matches design
- **GIVEN** the conditional button is displayed
- **WHEN** viewing the homepage
- **THEN** the button uses consistent Tailwind CSS styling
- **AND** the button is centered and appropriately sized
- **AND** the button is accessible and responsive
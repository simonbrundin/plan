# development-server Specification

## Purpose
TBD - created by archiving change fix-dev-server-500-error. Update Purpose after archive.
## Requirements
### Requirement: Development Server Startup
The development server MUST start successfully without returning HTTP 500 errors and MUST properly handle SSR without GraphQL client errors. The server MUST allow proper imports of GraphQL composables without Vite import-analysis restrictions. GraphQL operations MUST be available as functions and execute without "GqlSdks[client](...)[operation] is not a function" errors.

#### Scenario: Pages load without SSR GraphQL errors
- **WHEN** accessing pages like /root-goals
- **THEN** the page loads without HTTP 500 errors
- **AND** GraphQL data loads on client-side without useAsyncGql undefined errors
- **AND** GraphQL operations execute as functions without client errors
- **AND** navigation menu functions correctly
- **AND** no Vite import-analysis errors occur for GraphQL composables

#### Scenario: GraphQL imports work correctly
- **WHEN** importing useAsyncGql in Vue components
- **THEN** the import succeeds without "Importing directly from module entry-points is not allowed" errors
- **AND** GraphQL queries execute successfully
- **AND** operations are available as callable functions

#### Scenario: Goals display correctly in root-goals page
- **WHEN** a logged-in user accesses the /root-goals page
- **THEN** user's goals load and display without GraphQL client operation errors
- **AND** no "Component is already mounted" warnings occur for data fetching

### Requirement: GraphQL Composable Availability
The useAsyncGql composable MUST be available in Vue components and Pinia stores without import errors.

#### Scenario: Goals load successfully in root-goals page
- **WHEN** a logged-in user accesses the /root-goals page
- **THEN** useAsyncGql is defined and available
- **AND** goals load without "useAsyncGql is not defined" errors
- **AND** user's goals display correctly

#### Scenario: Goals store functions work
- **WHEN** the goals store calls loadGoals()
- **THEN** useAsyncGql executes GraphQL queries successfully
- **AND** no ReferenceError is thrown
- **AND** goal data is fetched and stored properly</content>
</xai:function_call">Now validate the proposal. 

<xai:function_call name="bash">
<parameter name="command">cd /Users/simon/repos/plan && openspec validate fix-useAsyncGql-undefined-error --strict

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


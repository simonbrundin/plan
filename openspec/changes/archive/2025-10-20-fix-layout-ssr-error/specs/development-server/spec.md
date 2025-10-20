## MODIFIED Requirements

### Requirement: Development Server Startup
The development server MUST start successfully without returning HTTP 500 errors and MUST properly handle SSR without GraphQL client errors.

#### Scenario: Pages load without SSR GraphQL errors
- **WHEN** accessing pages like /root-goals
- **THEN** the page loads without HTTP 500 errors
- **AND** GraphQL data loads on client-side without useAsyncGql undefined errors
- **AND** navigation menu functions correctly
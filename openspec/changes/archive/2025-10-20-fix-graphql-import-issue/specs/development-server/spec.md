## MODIFIED Requirements
### Requirement: Development Server Startup
The development server MUST start successfully without returning HTTP 500 errors and MUST properly handle SSR without GraphQL client errors. The server MUST allow proper imports of GraphQL composables without Vite import-analysis restrictions.

#### Scenario: Pages load without SSR GraphQL errors
- **WHEN** accessing pages like /root-goals
- **THEN** the page loads without HTTP 500 errors
- **AND** GraphQL data loads on client-side without useAsyncGql undefined errors
- **AND** navigation menu functions correctly
- **AND** no Vite import-analysis errors occur for GraphQL composables

#### Scenario: GraphQL imports work correctly
- **WHEN** importing useAsyncGql in Vue components
- **THEN** the import succeeds without "Importing directly from module entry-points is not allowed" errors
- **AND** GraphQL queries execute successfully
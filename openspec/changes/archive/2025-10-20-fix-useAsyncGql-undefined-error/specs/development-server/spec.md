# development-server Specification

## Purpose
The development server must start successfully and handle GraphQL operations without undefined composable errors.

## ADDED Requirements
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
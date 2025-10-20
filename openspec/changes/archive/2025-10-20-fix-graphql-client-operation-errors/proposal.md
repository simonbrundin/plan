## Why
The root-goals page fails to load goals with "GqlSdks[client](...)[operation] is not a function" error and "Component is already mounted, please use $fetch instead" warning. This prevents users from viewing their goals despite the application running and user authentication working correctly.

## What Changes
- Fix GraphQL client operation generation/configuration to ensure operations are properly available as functions
- Address the useAsyncData mounting warning by using appropriate data fetching methods
- Ensure goals load correctly in the root-goals page without GraphQL client errors

## Impact
- Affected specs: development-server
- Affected code: frontend/app/pages/root-goals.vue, GraphQL client configuration
- Fixes critical user-facing bug preventing goal viewing
- No breaking changes to existing functionality
## Why
The root-goals page fails to load goals with "ReferenceError: useAsyncGql is not defined" error, preventing users from seeing their goals even when logged in and having goals assigned to their user account. This breaks the core functionality of viewing goals.

## What Changes
- Fix the useAsyncGql undefined error by ensuring proper import or auto-import of the GraphQL composable
- Ensure goals load correctly in the root-goals page and goals store
- Maintain existing GraphQL functionality and authentication

## Impact
- Affected code: `frontend/app/stores/goals.ts`, `frontend/app/pages/root-goals.vue`
- Fixes critical user-facing bug preventing goal viewing
- No breaking changes to existing functionality
- Improves development server reliability
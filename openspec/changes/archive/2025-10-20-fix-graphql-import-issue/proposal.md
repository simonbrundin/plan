## Why
The root-goals page is failing to build due to a Vite import-analysis error that prevents importing `useAsyncGql` from the `nuxt-graphql-client` module entry-point. This is blocking development and the application from running properly.

## What Changes
- Fix the import statement in `app/pages/root-goals.vue` to use the correct import path or rely on Nuxt's auto-import system for `useAsyncGql`
- Ensure the GraphQL functionality continues to work as expected

## Impact
- Affected code: `frontend/app/pages/root-goals.vue`
- No changes to existing functionality or API
- Resolves build error that prevents the application from running
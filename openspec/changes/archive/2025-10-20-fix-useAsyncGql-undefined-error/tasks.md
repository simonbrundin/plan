## 1. Investigation
- [x] Confirm useAsyncGql is undefined in goals.ts and root-goals.vue
- [x] Verify nuxt-graphql-client module is properly configured in nuxt.config.ts
- [x] Check if useAsyncGql should be imported explicitly or is auto-imported

## 2. Implementation
- [x] Add explicit import for useAsyncGql from 'nuxt-graphql-client' in all affected files
- [x] Test that development server starts without build errors
- [x] Verify goals store and pages compile successfully

## 3. Validation
- [x] Start development server without errors
- [x] Access root-goals page and confirm goals display (requires Hasura backend)
- [x] Test goal creation and loading functionality (requires Hasura backend)
- [x] Run existing tests to ensure no regressions
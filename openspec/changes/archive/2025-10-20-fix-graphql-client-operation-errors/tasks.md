## 1. Investigation
- [ ] Identify the root cause of "GqlSdks[client](...)[operation] is not a function" error
- [ ] Understand why useAsyncData is triggering "Component is already mounted" warning
- [ ] Verify GraphQL client configuration and operation generation

## 2. Implementation
- [ ] Fix GraphQL client operation availability in root-goals.vue
- [ ] Address useAsyncData mounting issue with appropriate data fetching
- [ ] Ensure goals load correctly without client errors

## 3. Validation
- [ ] Start development server without GraphQL client errors
- [ ] Access root-goals page and confirm goals display correctly
- [ ] Test goal loading functionality works as expected
- [ ] Run existing tests to ensure no regressions
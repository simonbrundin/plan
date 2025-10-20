## 1. Diagnosis
- [x] Identify root cause: nuxt-graphql-client module disabled but code uses useAsyncGql
- [x] Confirm GraphQL usage throughout codebase (stores, pages, layouts)

## 2. Implementation
- [x] Re-enable nuxt-graphql-client module in nuxt.config.ts
- [x] Test dev server startup without errors
- [ ] Verify GraphQL queries work (may need Hasura running)

## 3. Validation
- [x] Start dev server successfully
- [x] Access application without 500 errors
- [ ] Confirm GraphQL data loading works (requires Hasura backend)</content>
</xai:function_call">Now I need to create the spec delta. Since this is fixing a bug in the development server, I'll create a new capability for "development-server" with a requirement that it should start without 500 errors. 

<xai:function_call name="write">
<parameter name="filePath">openspec/changes/fix-dev-server-500-error/specs/development-server/spec.md
## Why
The `/root-goals` page and other pages are returning HTTP 500 errors with "useAsyncGql is not defined" because the default layout is trying to execute GraphQL queries during SSR, but the GraphQL client is not available server-side.

## What Changes
- Fix the default layout to properly handle GraphQL data loading without SSR issues
- Move GraphQL calls to client-side only execution to prevent server-side errors
- Ensure pages can load without GraphQL client availability during SSR

## Impact
- Affected specs: development-server
- Affected code: default layout GraphQL data loading
- No breaking changes to existing functionality</content>
</xai:function_call">Now I'll create the tasks.md file. The tasks should be:
1. Analyze the current layout issue
2. Fix the GraphQL call in default layout
3. Test the root-goals page
4. Validate that the error is resolved

Let me create the tasks file. Then I'll create the spec delta. The spec should modify the existing development-server requirement to include proper SSR handling. Let me check what the current spec looks like first. 

<xai:function_call name="read">
<parameter name="filePath">openspec/specs/development-server/spec.md
## 1. Diagnosis
- [x] Identify root cause: Incorrect hydration logic using Object.setPrototypeOf()
- [x] Confirm error occurs during Pinia store serialization

## 2. Implementation
- [x] Remove problematic Object.setPrototypeOf() calls from goals store hydration
- [x] Implement proper object reconstruction during hydration
- [x] Fix useAsyncGql usage in layout to avoid SSR issues
- [ ] Test dev server startup without hydration errors

## 3. Validation
- [x] Start dev server successfully without 500 errors
- [x] Verify Pinia store hydration works correctly
- [x] Confirm application loads without hasOwnProperty errors</content>
</xai:function_call">Now I'll create the spec delta. Since this is modifying the existing development-server spec, I need to use MODIFIED Requirements. Let me create the spec delta. 

<xai:function_call name="write">
<parameter name="filePath">openspec/changes/fix-pinia-hydration-error/specs/development-server/spec.md
# Add Conditional Homepage Button

## Summary
Add a conditional button on the homepage that directs users to either the root-goals page (if logged in) or the login page (if not logged in).

## Motivation
Currently, the homepage only displays "Test" with no navigation options. Users need a clear way to access the main functionality of the application based on their authentication status.

## Impact
- Improves user experience by providing clear navigation from the homepage
- Reduces friction for both logged-in and non-logged-in users
- Maintains existing functionality while adding new UI elements

## Implementation Approach
- Modify `frontend/app/pages/index.vue` to include conditional button logic
- Use `useUserSession()` composable to check authentication status
- Add appropriate styling with Tailwind CSS to match existing design
- Ensure buttons are accessible and responsive

## Risks
- Minimal risk as this is a frontend-only change
- No database or API modifications required
- Existing functionality remains unchanged

## Testing
- Verify button appears and functions correctly for logged-in users
- Verify button appears and functions correctly for non-logged-in users
- Ensure no regression in existing homepage behavior
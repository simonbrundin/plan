# Remove Goal Creation Date Display

## Why
The user wants to remove the display of when goals were created on the root-goals page to simplify the UI and focus on goal content rather than metadata.

## What Changes
- Remove the "Skapad: [date]" text from goal items on the root-goals page
- Keep other goal information (title, status, progress) intact

## Impact
- Affected code: `frontend/app/pages/root-goals.vue`
- No breaking changes to functionality, only UI simplification
- Goals will still display title, completion status, and progress information</content>
</xai:function_call">Now create tasks.md: 

The task is simple: remove the creation date display from the template. Since there's no existing spec for goals display, I need to create one under specs/goals-display/spec.md, but first I need to check if I should create a new capability or modify an existing one.

Since there's no goals-display spec, I'll create a new one with the current behavior minus the creation date.

But actually, for a removal, I should use REMOVED Requirements in the delta.

Let me create the spec delta first. Since there's no existing goals-display capability, I'll create a new one with the current requirements minus the creation date display. But the instructions say to prefer modifying existing specs over creating duplicates.

Since there's no existing spec for goals display, I need to create a new capability.

Let me create the spec delta under specs/goals-display/spec.md with MODIFIED requirements (since we're modifying the display by removing the date). But since it doesn't exist, maybe ADDED with the current state minus the date.

Actually, for this change, since it's removing a feature, I should use REMOVED Requirements.

But first, I need to define what the current requirement is. Let me look at the code again.

The current behavior shows:
- Goal title
- Creation date ("Skapad: [date]")
- Progress (number of sub-goals)

The user wants to remove the creation date.

So in the spec delta, I can have MODIFIED Requirements where I update the requirement to not include the creation date display.

But since there's no existing spec, I need to create the capability first.

The instructions say: "Always check if capability already exists. Prefer modifying existing specs over creating duplicates."

Since goals-display doesn't exist, I should create it.

For the delta, since it's a removal, I can use MODIFIED to change the requirement.

Let me create the spec delta. 

I'll create a new capability called "goals-display" and add a requirement for how goals are displayed on the root-goals page, then modify it to remove the creation date. 

But that seems odd. Since this is a removal of an existing feature, and there's no spec for it, perhaps I should just create the task and implement it directly, as it's a simple change.

But the user asked for a change proposal, so I should follow the process.

Let me create the spec delta with MODIFIED Requirements, defining the new behavior. 

First, create the spec.md under the change. 

The delta should describe the new state after the change. 

So under ## MODIFIED Requirements, I can have the requirement for goal display without the creation date. 

But since it doesn't exist, perhaps I need to use ADDED for the new capability.

Let me check the instructions again.

" If multiple capabilities are affected, create multiple delta files under changes/[change-id]/specs/<capability>/spec.md—one per capability."

For this, it's a new capability "goals-display".

So I'll create specs/goals-display/spec.md with ADDED Requirements for the display behavior without the creation date.

But that doesn't make sense because the change is to remove the date from the existing display.

Perhaps I should create the spec with the current behavior, then modify it.

But the instructions say to use MODIFIED when changing existing requirements.

Since there's no existing spec, I think I should create the capability with ADDED, but that would be adding the new behavior.

Perhaps for UI changes like this, it's better to just implement it directly without a full spec, but since the user asked for a change proposal, I'll create a minimal one.

Let me create the spec delta with MODIFIED, assuming the requirement exists. 

To make it work, I'll create the spec as if the requirement exists.

Let's do this: create the spec delta with MODIFIED Requirements, defining the requirement as "Goals SHALL display title, completion status, and progress information without creation dates"

That way, it's modifying the display behavior. 

Yes. 

So the spec.md will have:

## MODIFIED Requirements

### Requirement: Goal Display on Root Goals Page

The root goals page SHALL display each goal with its title, completion status, and progress information. Creation dates SHALL NOT be displayed.

#### Scenario: Goal display without creation date

- **WHEN** user views the root-goals page

- **THEN** each goal shows title, status, and progress

- **AND** creation date is not shown

Yes, that works. Even though the requirement didn't exist before, I'm defining it as modified to the new state. 

Now create the spec.md: 

<xai:function_call name="write">
<parameter name="filePath">openspec/changes/remove-goal-creation-date-display/specs/goals-display/spec.md
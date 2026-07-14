import { Given, When, Then, Before, world } from "@cucumber/cucumber";
import assert from "node:assert";
import type { PriorityWorld } from "./world";

// API calls made during the scenario (PATCH /api/goals/:id)
const apiPatchCalls: { goalId: number; body: Record<string, unknown> }[] = [];

Before(() => {
	apiPatchCalls.length = 0;
});

function w(): PriorityWorld {
	return world as PriorityWorld;
}

function makeGoal(title: string, started: string | null) {
	return {
		id: Math.floor(Math.random() * 1000000),
		title,
		started,
		finished: null,
		created: new Date().toISOString(),
		icon: "heroicons:star",
		inbox: 0,
	};
}

function findGoal(pw: PriorityWorld, title: string) {
	return pw.goals.find((g) => g.title === title);
}

// --- Given -----------------------------------------------------------

Given("the user is logged in", () => {
	// World starts with empty lists in constructor
});

// Regex: captures quoted title like "Träna"
Given(/the user has a goal "([^"]+)" that is not started/, (_title: string) => {
	w().goals.push(makeGoal(_title, null));
});

Given(/the user has a goal "([^"]+)" that is started/, (_title: string) => {
	w().goals.push(makeGoal(_title, new Date().toISOString()));
});

Given(/the user has the goals:/, (_table: { raw: () => string[][] }) => {
	const rows = _table.raw();
	// rows[0] = [| title | started |]
	for (let i = 1; i < rows.length; i++) {
		const [title, started] = rows[i];
		const startedValue =
			started.trim() === "yes" ? new Date().toISOString() : null;
		w().goals.push(makeGoal(title.trim(), startedValue));
	}
});

// Regex: captures quoted filter like "Alla" or "Påbörjade"
Given(/the filter is "([^"]+)"/, (_filter: string) => {
	const pw = w();
	pw.startedFilter = _filter as PriorityWorld["startedFilter"];
	pw.visibleTitles = pw.goals
		.filter((g) => pw.filterMatches(g))
		.map((g) => g.title);
});

Given(/a goal has started value null/, () => {
	w().goals.push(makeGoal("Testmål", null));
});

// --- When -----------------------------------------------------------

When(/the user marks the goal "([^"]+)" as started/, (_title: string) => {
	const goal = findGoal(w(), _title);
	if (!goal) throw new Error(`Goal "${_title}" not found`);
	goal.started = new Date().toISOString();
	apiPatchCalls.push({ goalId: goal.id, body: { started: goal.started } });
});

When(/the user unmarks the goal "([^"]+)"/, (_title: string) => {
	const goal = findGoal(w(), _title);
	if (!goal) throw new Error(`Goal "${_title}" not found`);
	goal.started = null;
	apiPatchCalls.push({ goalId: goal.id, body: { started: null } });
});

When(/the user selects the filter "([^"]+)"/, (_filter: string) => {
	const pw = w();
	pw.startedFilter = _filter as PriorityWorld["startedFilter"];
	pw.visibleTitles = pw.goals
		.filter((g) => pw.filterMatches(g))
		.map((g) => g.title);
});

When(/the user marks the goal as started/, () => {
	const goal = w().goals[0];
	if (!goal) throw new Error("No goal to mark");
	goal.started = new Date().toISOString();
	apiPatchCalls.push({ goalId: goal.id, body: { started: goal.started } });
});

// --- Then -----------------------------------------------------------

Then(/the goal "([^"]+)" should be marked as started/, (_title: string) => {
	const goal = findGoal(w(), _title);
	assert(goal, `Goal "${_title}" not found`);
	assert(
		goal.started !== null,
		`Goal "${_title}" was expected to be started but started is null`,
	);
});

Then(/the goal "([^"]+)" should not be marked as started/, (_title: string) => {
	const goal = findGoal(w(), _title);
	assert(goal, `Goal "${_title}" not found`);
	assert(
		goal.started === null,
		`Goal "${_title}" was NOT expected to be started but started is ${goal.started}`,
	);
});

Then(/only the goal "([^"]+)" should be visible/, (_title: string) => {
	const pw = w();
	const visible = pw.goals.filter((g) => pw.filterMatches(g));
	assert.equal(
		visible.length,
		1,
		`Expected 1 visible goal, got ${visible.length}: ${visible
			.map((g) => g.title)
			.join(", ")}`,
	);
	assert.equal(visible[0].title, _title);
});

Then(/the goal "([^"]+)" should not be visible/, (_hiddenTitle: string) => {
	const pw = w();
	const visible = pw.goals.filter((g) => pw.filterMatches(g));
	const found = visible.find((g) => g.title === _hiddenTitle);
	assert.equal(
		found,
		undefined,
		`Goal "${_hiddenTitle}" should NOT be visible but it is`,
	);
});

// Regex: captures two quoted goal titles
Then(
	/both goals "([^"]+)" and "([^"]+)" should be visible/,
	(_a: string, _b: string) => {
		const pw = w();
		const visible = pw.goals.filter((g) => pw.filterMatches(g));
		const titles = visible.map((g) => g.title).sort();
		const expected = [_a, _b].sort();
		assert.deepEqual(
			titles,
			expected,
			`Expected [${expected.join(", ")}], got [${titles.join(", ")}]`,
		);
	},
);

Then(
	/the API should receive started value "([^"]+)"/,
	(_expectedValue: string) => {
		assert.equal(apiPatchCalls.length, 1);
		const { body } = apiPatchCalls[0];
		if (_expectedValue === "<datum>") {
			assert.notEqual(
				body.started,
				null,
				"API was expected to receive a date but got null",
			);
			assert.ok(
				typeof body.started === "string" && (body.started as string).length > 0,
				`API was expected to receive a timestamp but got "${body.started}"`,
			);
		} else if (_expectedValue === "null") {
			assert.equal(
				body.started,
				null,
				`API was expected to receive null but got "${body.started}"`,
			);
		}
	},
);

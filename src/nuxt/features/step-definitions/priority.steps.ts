import { Given, When, Then, Before, setWorldConstructor, World } from "@cucumber/cucumber";
import assert from "node:assert";

// --- PriorityWorld class ---

export type StartedFilter = "all" | "started" | "not_started";

function matchesStartedFilter(
	goal: { started: string | null },
	filter: StartedFilter,
): boolean {
	if (filter === "all") return true;
	if (filter === "started") return goal.started !== null;
	return goal.started === null;
}

export class PriorityWorld extends World {
	readonly goals: {
		id: number;
		title: string;
		started: string | null;
		finished: string | null;
		created: string;
		icon: string;
		inbox: number;
	}[] = [];
	visibleTitles: string[] = [];
	startedFilter: StartedFilter = "all";

	filterMatches(g: { started: string | null }): boolean {
		return matchesStartedFilter(g, this.startedFilter);
	}
}

setWorldConstructor(PriorityWorld);

// --- API call tracker ---

const apiPatchCalls: { goalId: number; body: Record<string, unknown> }[] = [];

Before(() => {
	apiPatchCalls.length = 0;
});

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

// --- Given steps ---

Given("the user is logged in", () => {
	// World starts with empty lists in constructor
});

Given(
	/the user has a goal "([^"]+)" that is not started/,
	function (this: PriorityWorld, _title: string) {
		this.goals.push(makeGoal(_title, null));
	},
);

Given(
	/the user has a goal "([^"]+)" that is started/,
	function (this: PriorityWorld, _title: string) {
		this.goals.push(makeGoal(_title, new Date().toISOString()));
	},
);

Given(
	/the user has the goals:/,
	function (this: PriorityWorld, _table: { raw: () => string[][] }) {
		const rows = _table.raw();
		for (let i = 1; i < rows.length; i++) {
			const [title, started] = rows[i];
			const startedValue =
				started.trim() === "yes" ? new Date().toISOString() : null;
			this.goals.push(makeGoal(title.trim(), startedValue));
		}
	},
);

Given(
	/the filter is "([^"]+)"/,
	function (this: PriorityWorld, _filter: string) {
		const filterMap: Record<string, StartedFilter> = {
Alla: "all",
Påbörjade: "started",
"Ej påbörjade": "not_started",
		};
		this.startedFilter = filterMap[_filter] ?? "all";
		this.visibleTitles = this.goals
			.filter((g) => this.filterMatches(g))
			.map((g) => g.title);
	},
);

Given(
	/a goal has started value null/,
	function (this: PriorityWorld) {
		this.goals.push(makeGoal("Testmål", null));
	},
);

// --- When steps ---

When(
	/the user marks the goal "([^"]+)" as started/,
	function (this: PriorityWorld, _title: string) {
		const goal = findGoal(this, _title);
		if (!goal) throw new Error(`Goal "${_title}" not found`);
		goal.started = new Date().toISOString();
		apiPatchCalls.push({ goalId: goal.id, body: { started: goal.started } });
	},
);

When(
	/the user unmarks the goal "([^"]+)"/,
	function (this: PriorityWorld, _title: string) {
		const goal = findGoal(this, _title);
		if (!goal) throw new Error(`Goal "${_title}" not found`);
		goal.started = null;
		apiPatchCalls.push({ goalId: goal.id, body: { started: null } });
	},
);

When(
	/the user selects the filter "([^"]+)"/,
	function (this: PriorityWorld, _filter: string) {
		const filterMap: Record<string, StartedFilter> = {
			Alla: "all",
			Påbörjade: "started",
			"Ej påbörjade": "not_started",
		};
		this.startedFilter = filterMap[_filter] ?? "all";
		this.visibleTitles = this.goals
			.filter((g) => this.filterMatches(g))
			.map((g) => g.title);
	},
);

When(
	/the user marks the goal as started/,
	function (this: PriorityWorld) {
		const goal = this.goals[0];
		if (!goal) throw new Error("No goal to mark");
		goal.started = new Date().toISOString();
		apiPatchCalls.push({ goalId: goal.id, body: { started: goal.started } });
	},
);

// --- Then steps ---

Then(
	/the goal "([^"]+)" should be marked as started/,
	function (this: PriorityWorld, _title: string) {
		const goal = findGoal(this, _title);
		assert(goal, `Goal "${_title}" not found`);
		assert(
			goal.started !== null,
			`Goal "${_title}" was expected to be started but started is null`,
		);
	},
);

Then(
	/the goal "([^"]+)" should not be marked as started/,
	function (this: PriorityWorld, _title: string) {
		const goal = findGoal(this, _title);
		assert(goal, `Goal "${_title}" not found`);
		assert(
			goal.started === null,
			`Goal "${_title}" was NOT expected to be started but started is ${goal.started}`,
		);
	},
);

Then(
	/only the goal "([^"]+)" should be visible/,
	function (this: PriorityWorld, _title: string) {
		const visible = this.goals.filter((g) => this.filterMatches(g));
		assert.equal(
			visible.length,
			1,
			`Expected 1 visible goal, got ${visible.length}: ${visible
				.map((g) => g.title)
				.join(", ")}`,
		);
		assert.equal(visible[0].title, _title);
	},
);

Then(
	/the goal "([^"]+)" should not be visible/,
	function (this: PriorityWorld, _hiddenTitle: string) {
		const visible = this.goals.filter((g) => this.filterMatches(g));
		const found = visible.find((g) => g.title === _hiddenTitle);
		assert.equal(
			found,
			undefined,
			`Goal "${_hiddenTitle}" should NOT be visible but it is`,
		);
	},
);

Then(
	/both goals "([^"]+)" and "([^"]+)" should be visible/,
	function (this: PriorityWorld, _a: string, _b: string) {
		const visible = this.goals.filter((g) => this.filterMatches(g));
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
				typeof body.started === "string" &&
					(body.started as string).length > 0,
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

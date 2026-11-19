import { Given, Then, When } from "@cucumber/cucumber";
import assert from "node:assert";

import { isUnfinishedGoal } from "../../app/utils/goalFilters.ts";
import {
	isRecurringGoal,
	isWaitingForNextOccurrence,
	recurrenceCount,
} from "../../app/utils/recurrence.ts";

interface RecurringGoal {
	id: number;
	title: string;
	finished: string | null;
	recurrence_rule: string | null;
	recurrence_next: string | null;
	recurrence_count: number;
}

// Cucumber har bara en World-klass per körning, så den här filen delar
// inte constructor. Tillståndet hålls i en WeakMap per scenario.
const states = new WeakMap<object, RecurringGoal[]>();

const goalsOf = (world: object): RecurringGoal[] => {
	let goals = states.get(world);
	if (!goals) {
		goals = [];
		states.set(world, goals);
	}
	return goals;
};

const localDate = (offsetDays: number): string => {
	const date = new Date();
	date.setDate(date.getDate() + offsetDays);
	const pad = (value: number) => String(value).padStart(2, "0");
	return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
};

const findGoal = (world: object, title: string): RecurringGoal => {
	const goal = goalsOf(world).find((g) => g.title === title);
	if (!goal) throw new Error(`Goal "${title}" not found`);
	return goal;
};

let nextGoalId = 1;

const makeGoal = (
	title: string,
	overrides: Partial<RecurringGoal> = {},
): RecurringGoal => ({
	id: nextGoalId++,
	title,
	finished: null,
	recurrence_rule: null,
	recurrence_next: null,
	recurrence_count: 0,
	...overrides,
});

Given(
	/the user has a recurring goal "([^"]+)" whose next occurrence is (today|tomorrow|yesterday)/,
	function (this: object, title: string, when: string) {
		const offset = when === "today" ? 0 : when === "tomorrow" ? 1 : -1;
		goalsOf(this).push(
			makeGoal(title, {
				recurrence_rule: "FREQ=DAILY;INTERVAL=2",
				recurrence_next: localDate(offset),
			}),
		);
	},
);

Given(
	/the user has a finished goal "([^"]+)"/,
	function (this: object, title: string) {
		goalsOf(this).push(
			makeGoal(title, { finished: new Date().toISOString() }),
		);
	},
);

When(
	/the occurrence of "([^"]+)" is completed/,
	function (this: object, title: string) {
		const goal = findGoal(this, title);
		goal.recurrence_count += 1;
	},
);

When(/the series of "([^"]+)" ends/, function (this: object, title: string) {
	const goal = findGoal(this, title);
	goal.recurrence_next = null;
	goal.finished = new Date().toISOString();
});

Then(
	/the goal "([^"]+)" should still be listed/,
	function (this: object, title: string) {
		const goal = findGoal(this, title);
		assert.equal(
			isUnfinishedGoal(goal),
			true,
			`Goal "${title}" was expected to be listed but it is finished`,
		);
	},
);

Then(
	/the goal "([^"]+)" should be hidden from the list/,
	function (this: object, title: string) {
		const goal = findGoal(this, title);
		assert.equal(
			isUnfinishedGoal(goal),
			false,
			`Goal "${title}" was expected to be hidden but it is open`,
		);
	},
);

Then(
	/the goal "([^"]+)" should be waiting for its next occurrence/,
	function (this: object, title: string) {
		const goal = findGoal(this, title);
		assert.equal(isRecurringGoal(goal), true);
		assert.equal(
			isWaitingForNextOccurrence(goal),
			true,
			`Goal "${title}" was expected to wait for its next occurrence`,
		);
	},
);

Then(
	/the goal "([^"]+)" should not be waiting for its next occurrence/,
	function (this: object, title: string) {
		const goal = findGoal(this, title);
		assert.equal(
			isWaitingForNextOccurrence(goal),
			false,
			`Goal "${title}" was not expected to wait for its next occurrence`,
		);
	},
);

Then(
	/the goal "([^"]+)" should have completed (\d+) occurrence/,
	function (this: object, title: string, count: string) {
		const goal = findGoal(this, title);
		assert.equal(recurrenceCount(goal), Number(count));
	},
);

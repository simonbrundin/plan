import { Given, Then, When } from "@cucumber/cucumber";
import assert from "node:assert";

import {
	describeParsedEntry,
	parseNaturalLanguage,
	resolveQuickAddMode,
	type ParsedEntry,
} from "../../app/utils/naturalLanguage.ts";

// Cucumber har bara en World-klass per körning, så den här filen delar
// inte constructor. Tillståndet hålls i en WeakMap per scenario.
interface NaturalLanguageState {
	referenceTime: Date;
	entry: ParsedEntry | null;
}

const states = new WeakMap<object, NaturalLanguageState>();

const stateOf = (world: object): NaturalLanguageState => {
	let state = states.get(world);
	if (!state) {
		state = { referenceTime: new Date("2026-03-10T09:00:00"), entry: null };
		states.set(world, state);
	}
	return state;
};

const entryOf = (world: object): ParsedEntry => {
	const entry = stateOf(world).entry;
	if (!entry) throw new Error("No parsed entry");
	return entry;
};

const pad = (value: number) => String(value).padStart(2, "0");

const formatDate = (date: Date): string =>
	`${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;

const formatTime = (date: Date): string =>
	`${pad(date.getHours())}:${pad(date.getMinutes())}`;

const formatDateTime = (date: Date): string =>
	`${formatDate(date)}T${formatTime(date)}:${pad(date.getSeconds())}`;

Given(/the reference time is "([^"]+)"/, function (this: object, value: string) {
	stateOf(this).referenceTime = new Date(value);
});

When(/the user writes "([^"]*)"/, function (this: object, value: string) {
	const state = stateOf(this);
	state.entry = parseNaturalLanguage(value, state.referenceTime);
});

Then(/the title should be "([^"]*)"/, function (this: object, value: string) {
	assert.equal(entryOf(this).title, value);
});

Then(/no date should be found/, function (this: object) {
	assert.equal(entryOf(this).hasDate, false);
});

Then(/no time should be found/, function (this: object) {
	assert.equal(entryOf(this).hasTime, false);
});

Then(/no time should be explicit/, function (this: object) {
	assert.equal(entryOf(this).hasTime, false);
});

Then(/the date should be "([^"]+)"/, function (this: object, value: string) {
	const entry = entryOf(this);
	assert.ok(entry.startAt, "Expected a start time");
	assert.equal(formatDate(entry.startAt), value);
});

Then(/the time should be "([^"]+)"/, function (this: object, value: string) {
	const entry = entryOf(this);
	assert.ok(entry.startAt, "Expected a start time");
	assert.equal(formatTime(entry.startAt), value);
});

Then(/the start should be "([^"]+)"/, function (this: object, value: string) {
	const entry = entryOf(this);
	assert.ok(entry.startAt, "Expected a start time");
	assert.equal(formatDateTime(entry.startAt), value);
});

Then(/the end should be "([^"]+)"/, function (this: object, value: string) {
	const entry = entryOf(this);
	assert.ok(entry.endAt, "Expected an end time");
	assert.equal(formatDateTime(entry.endAt), value);
});

Then(/the event should be all day/, function (this: object) {
	assert.equal(entryOf(this).allDay, true);
});

Then(/the recurrence rule should be "([^"]+)"/, function (this: object, value: string) {
	assert.equal(entryOf(this).recurrenceRule, value);
});

Then(/the date should not be explicit/, function (this: object) {
	assert.equal(entryOf(this).dateExplicit, false);
});

Then(/the date should be explicit/, function (this: object) {
	assert.equal(entryOf(this).dateExplicit, true);
});

Then(/the mode should be "([^"]+)" by default/, function (this: object, value: string) {
	assert.equal(resolveQuickAddMode(entryOf(this)), value);
});

Then(/the icon should be "([^"]*)"/, function (this: object, value: string) {
	assert.equal(entryOf(this).icon, value);
});

Then(/no icon should be found/, function (this: object) {
	assert.equal(entryOf(this).icon, null);
});

Then(
	/the mode should be "([^"]+)" when the mode is "([^"]+)"/,
	function (this: object, value: string, mode: string) {
		assert.equal(
			resolveQuickAddMode(
				entryOf(this),
				mode as "auto" | "event" | "goal",
			),
			value,
		);
	},
);

Then(/the description should mention "([^"]+)"/, function (this: object, value: string) {
	const description = describeParsedEntry(entryOf(this));
	assert.ok(
		description.includes(value),
		`Expected description "${description}" to include "${value}"`,
	);
});

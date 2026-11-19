import { Given, Then, When } from "@cucumber/cucumber";
import assert from "node:assert";

import {
	layoutEventRows,
	type EventRow,
} from "../../app/utils/calendarLayout.ts";
import { resolveScaleLevel } from "../../app/utils/calendarScale.ts";
import {
	planVisibility,
	type VisibilitySeries,
} from "../../app/utils/calendarVisibility.ts";

/**
 * Steg för kalenderns händelserader. Tillståndet hålls i en WeakMap per
 * scenario, precis som i de övriga step-definitionerna.
 */
interface CalendarRowsState {
	duration: number;
	height: number;
	events: Array<{ start: number; end: number; seriesKey: string | null }>;
	rows: EventRow[];
	dailyLimit: number;
	hidingEnabled: boolean;
	focusKey: string | null;
	hiddenBefore: string[];
	series: VisibilitySeries[];
	hidden: string[];
}

const states = new WeakMap<object, CalendarRowsState>();

const stateOf = (world: object): CalendarRowsState => {
	let state = states.get(world);
	if (!state) {
		state = {
			duration: 24 * 60 * 60 * 1000,
			height: 600,
			events: [],
			rows: [],
			dailyLimit: 7,
			hidingEnabled: true,
			focusKey: null,
			hiddenBefore: [],
			series: [],
			hidden: [],
		};
		states.set(world, state);
	}
	return state;
};

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

Given(
	/the visible duration is (\d+) (hours|days)/,
	function (this: object, amount: string, unit: string) {
		const value = Number(amount);
		stateOf(this).duration = unit === "hours" ? value * HOUR : value * DAY;
	},
);

Then(
	/the scale level should be "([^"]*)"/,
	function (this: object, expected: string) {
		assert.equal(resolveScaleLevel(stateOf(this).duration), expected);
	},
);

Given(
	/the calendar view is (\d+) pixels tall/,
	function (this: object, height: string) {
		stateOf(this).height = Number(height);
	},
);

/**
 * Läser klockslag som "HH:MM" och gör dem till millisekunder från
 * tidsfönstrets början.
 */
const toMilliseconds = (value: string): number => {
	const [hours, minutes] = value.trim().split(":").map(Number);
	return (hours ?? 0) * HOUR + (minutes ?? 0) * 60 * 1000;
};

Given(
	/the following calendar events exist:/,
	function (this: object, table: { raw: () => string[][] }) {
		const rows = table.raw();
		const events: CalendarRowsState["events"] = [];

		for (let i = 1; i < rows.length; i++) {
			const [, start, end, seriesKey] = rows[i];
			events.push({
				start: toMilliseconds(start ?? ""),
				end: toMilliseconds(end ?? ""),
				seriesKey: seriesKey?.trim() ? seriesKey.trim() : null,
			});
		}

		stateOf(this).events = events;
	},
);

When(/the events are laid out/, function (this: object) {
	const state = stateOf(this);

	state.rows = layoutEventRows(
		state.events.map((event) => ({ start: event.start, end: event.end })),
		{
			from: 0,
			to: state.duration,
			height: state.height,
		},
	);
});

Then(
	/the (first|second) row should start (\d+) pixels from the top/,
	function (this: object, which: string, expected: string) {
		const rows = stateOf(this).rows;
		const row = which === "first" ? rows[0] : rows[1];

		assert.equal(row?.top, Number(expected));
	},
);

Then(
	/the (first|second) row should be (\d+) pixels tall/,
	function (this: object, which: string, expected: string) {
		const rows = stateOf(this).rows;
		const row = which === "first" ? rows[0] : rows[1];

		assert.equal(row?.height, Number(expected));
	},
);

Then(
	/the first row should be at least (\d+) pixels tall/,
	function (this: object, expected: string) {
		const row = stateOf(this).rows[0];

		assert.ok(
			(row?.height ?? 0) >= Number(expected),
			`Raden ska vara minst ${expected} pixlar hög men var ${row?.height}`,
		);
	},
);

/**
 * Två rader får inte ligga ovanpå varandra. Alla rader ligger i samma
 * kolumn, så en rad som krockar skjuts ned under den som redan står.
 */
Then(/the rows should not overlap/, function (this: object) {
	const rows = [...stateOf(this).rows].sort((left, right) => left.top - right.top);

	for (let i = 1; i < rows.length; i += 1) {
		const previous = rows[i - 1]!;
		const current = rows[i]!;

		assert.ok(
			current.top >= previous.top + previous.height,
			`Rad ${i} börjar på ${current.top} pixlar, men raden ovanför slutar på ${
				previous.top + previous.height
			}`,
		);
	}
});

Then(/the rows should be (\d+)/, function (this: object, expected: string) {
	assert.equal(stateOf(this).rows.length, Number(expected));
});

Then(/(\d+) rows should be clipped/, function (this: object, expected: string) {
	const count = stateOf(this).rows.filter((row) => row.clipped).length;

	assert.equal(count, Number(expected));
});

Given(
	/the daily visibility limit is (\d+)/,
	function (this: object, limit: string) {
		stateOf(this).dailyLimit = Number(limit);
	},
);

Given(
	/the following calendar series exist:/,
	function (this: object, table: { raw: () => string[][] }) {
		const rows = table.raw();
		const series: VisibilitySeries[] = [];

		for (let i = 1; i < rows.length; i++) {
			const [key, count, daily] = rows[i];
			series.push({
				key: (key ?? "").trim(),
				title: (key ?? "").trim(),
				count: Number(count),
				daily: (daily ?? "").trim().toLowerCase() === "yes",
			});
		}

		stateOf(this).series = series;
	},
);

Given(
	/the series "([^"]*)" was hidden before/,
	function (this: object, key: string) {
		stateOf(this).hiddenBefore = [key];
	},
);

Given(/hiding of recurring series is off/, function (this: object) {
	stateOf(this).hidingEnabled = false;
});

Given(
	/the series "([^"]*)" is focused/,
	function (this: object, key: string) {
		stateOf(this).focusKey = key;
	},
);

When(/the visibility is planned/, function (this: object) {
	const state = stateOf(this);

	const plan = planVisibility({
		events: state.events,
		series: state.series,
		dailyLimit: state.dailyLimit,
		hidingEnabled: state.hidingEnabled,
		focusKey: state.focusKey,
		hiddenBefore: state.hiddenBefore,
		from: 0,
		to: state.duration,
		height: state.height,
	});

	state.hidden = plan.hidden;
});

Then(
	/the series "([^"]*)" should be hidden/,
	function (this: object, key: string) {
		assert.ok(
			stateOf(this).hidden.includes(key),
			`Serien ${key} borde vara dold, men dolda är ${stateOf(this).hidden.join(", ")}`,
		);
	},
);

Then(
	/the series "([^"]*)" should be visible/,
	function (this: object, key: string) {
		assert.ok(
			!stateOf(this).hidden.includes(key),
			`Serien ${key} borde vara synlig, men dolda är ${stateOf(this).hidden.join(", ")}`,
		);
	},
);

Then(/no series should be hidden/, function (this: object) {
	assert.deepEqual(stateOf(this).hidden, []);
});

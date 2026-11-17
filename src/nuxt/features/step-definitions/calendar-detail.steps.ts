import { Given, Then, When } from "@cucumber/cucumber";
import assert from "node:assert";

import {
	assignClusteredLanes,
	assignLanes,
	groupEventsByDay,
	groupEventsByMonth,
	groupEventsByYear,
	resolveDetailLevel,
	resolveDetailLevelWithHysteresis,
	resolveScaleLevel,
	type CalendarEventLike,
	type CalendarMarker,
	type DetailLevel,
} from "../../app/utils/calendarDetail.ts";

/**
 * Steg för kalenderns detaljnivåer. Tillståndet hålls i en WeakMap per
 * scenario, precis som i de övriga step-definitionerna.
 */
interface CalendarDetailState {
	duration: number;
	events: CalendarEventLike[];
	dayGroups: CalendarMarker[];
	monthGroups: CalendarMarker[];
	yearGroups: CalendarMarker[];
	positions: number[];
	lanes: number[];
	intervals: Array<[number, number]>;
	intervalLanes: string[];
	previousLevel: DetailLevel;
}

const states = new WeakMap<object, CalendarDetailState>();

const stateOf = (world: object): CalendarDetailState => {
	let state = states.get(world);
	if (!state) {
		state = {
			duration: 0,
			events: [],
			dayGroups: [],
			monthGroups: [],
			yearGroups: [],
			positions: [],
			lanes: [],
			intervals: [],
			intervalLanes: [],
			previousLevel: "full",
		};
		states.set(world, state);
	}
	return state;
};

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

Given(/the visible duration is (\d+) (hours|days)/, function (this: object, amount: string, unit: string) {
	const value = Number(amount);
	stateOf(this).duration = unit === "hours" ? value * HOUR : value * DAY;
});

Then(/the scale level should be "([^"]*)"/, function (this: object, expected: string) {
	assert.equal(resolveScaleLevel(stateOf(this).duration), expected);
});

Then(/the detail level should be "([^"]*)"/, function (this: object, expected: string) {
	assert.equal(resolveDetailLevel(stateOf(this).duration), expected);
});

Given(/the current detail level is "([^"]*)"/, function (this: object, level: string) {
	stateOf(this).previousLevel = level as DetailLevel;
});

Then(/the detail level with hysteresis should be "([^"]*)"/, function (this: object, expected: string) {
	const state = stateOf(this);
	const resolved = resolveDetailLevelWithHysteresis(state.duration, state.previousLevel);

	assert.equal(resolved, expected);

	// Nästa steg utgår från den nivå som faktiskt gäller.
	state.previousLevel = resolved;
});

Given(
	/the following calendar events exist:/,
	function (this: object, table: { raw: () => string[][] }) {
		const rows = table.raw();
		const events: CalendarEventLike[] = [];

		for (let i = 1; i < rows.length; i++) {
			const [title, start, end, recurrence, series] = rows[i];
			const event: CalendarEventLike = {
				// Kolumnen "series" låter flera rader dela samma serie-id,
				// vilket är hur förekomster från en återkommande serie ser
				// ut i verkligheten.
				id: series?.trim() ? Number(series.trim()) : events.length + 1,
				title: title.trim(),
				start_at: start.trim(),
				end_at: end.trim(),
			};

			if (recurrence?.trim()) event.recurrence_rule = recurrence.trim();

			events.push(event);
		}

		stateOf(this).events = events;
	},
);

When(/the events are grouped by day/, function (this: object) {
	stateOf(this).dayGroups = groupEventsByDay(stateOf(this).events);
});

When(/the events are grouped by month/, function (this: object) {
	stateOf(this).monthGroups = groupEventsByMonth(stateOf(this).events);
});

When(/the events are grouped by year/, function (this: object) {
	stateOf(this).yearGroups = groupEventsByYear(stateOf(this).events);
});

Then(/there should be (\d+) day groups?/, function (this: object, expected: string) {
	assert.equal(stateOf(this).dayGroups.length, Number(expected));
});

Then(/there should be (\d+) month groups?/, function (this: object, expected: string) {
	assert.equal(stateOf(this).monthGroups.length, Number(expected));
});

Then(/there should be (\d+) year groups?/, function (this: object, expected: string) {
	assert.equal(stateOf(this).yearGroups.length, Number(expected));
});

Then(
	/the (first|second) year group should contain (\d+) events?/,
	function (this: object, which: string, expected: string) {
		const groups = stateOf(this).yearGroups;
		const group = which === "first" ? groups[0] : groups[1];
		assert.equal(group?.count, Number(expected));
	},
);

Then(
	/the (first|second) day group should contain (\d+) events?/,
	function (this: object, which: string, expected: string) {
		const groups = stateOf(this).dayGroups;
		const group = which === "first" ? groups[0] : groups[1];
		assert.equal(group?.count, Number(expected));
	},
);

Then(
	/the (first|second) month group should contain (\d+) events?/,
	function (this: object, which: string, expected: string) {
		const groups = stateOf(this).monthGroups;
		const group = which === "first" ? groups[0] : groups[1];
		assert.equal(group?.count, Number(expected));
	},
);

Then(
	/the (first|second) day group should be titled "([^"]*)"/,
	function (this: object, which: string, expected: string) {
		const groups = stateOf(this).dayGroups;
		const group = which === "first" ? groups[0] : groups[1];
		assert.ok(
			group?.title.startsWith(expected),
			`Förväntade att etiketten skulle börja med "${expected}" men fick "${group?.title}"`,
		);
	},
);

Then(
	/the first day group should start before the second day group/,
	function (this: object) {
		const groups = stateOf(this).dayGroups;
		assert.ok(
			(groups[0]?.startAt ?? 0) < (groups[1]?.startAt ?? 0),
			"Den första gruppen ska ligga före den andra i tiden",
		);
	},
);

Then(
	/the (first|second) (day|month|year) group should mention "([^"]*)"/,
	function (this: object, which: string, period: string, expected: string) {
		const state = stateOf(this);
		const groups =
			period === "day"
				? state.dayGroups
				: period === "month"
					? state.monthGroups
					: state.yearGroups;
		const group = which === "first" ? groups[0] : groups[1];

		assert.ok(
			group?.title.includes(expected),
			`Förväntade att etiketten skulle innehålla "${expected}" men fick "${group?.title}"`,
		);
	},
);

Then(
	/the (first|second) (day|month|year) group should be recurring/,
	function (this: object, which: string, period: string) {
		const state = stateOf(this);
		const groups =
			period === "day"
				? state.dayGroups
				: period === "month"
					? state.monthGroups
					: state.yearGroups;
		const group = which === "first" ? groups[0] : groups[1];

		assert.equal(group?.recurring, true);
	},
);

Given(/markers are placed at ([\d, ]+) percent/, function (this: object, list: string) {
	stateOf(this).positions = list.split(",").map((value) => Number(value.trim()));
});

When(
	/the markers are packed into lanes with a minimum gap of (\d+) percent/,
	function (this: object, gap: string) {
		stateOf(this).lanes = assignLanes(stateOf(this).positions, Number(gap));
	},
);

Then(/the marker lanes should be "([^"]*)"/, function (this: object, expected: string) {
	assert.equal(stateOf(this).lanes.join(", "), expected);
});

Given(
	/intervals are placed at ([\d, -]+) percent/,
	function (this: object, list: string) {
		const intervals = list.split(",").map((entry): [number, number] => {
			const [from, to] = entry.trim().split("-").map(Number);
			return [from, to];
		});
		stateOf(this).intervals = intervals;
	},
);

When(
	/the intervals are packed into lanes with a minimum gap of (\d+) percent/,
	function (this: object, gap: string) {
		stateOf(this).intervalLanes = assignClusteredLanes(
			stateOf(this).intervals,
			Number(gap),
		).map((placement) => `${placement.lane}/${placement.lanes}`);
	},
);

Then(/the interval lanes should be "([^"]*)"/, function (this: object, expected: string) {
	assert.equal(stateOf(this).intervalLanes.join(", "), expected);
});

Then(
	/the (first|second|third) event should use (\d+) columns?/,
	function (this: object, which: string, expected: string) {
		const lanes = stateOf(this).intervalLanes;
		const index = which === "first" ? 0 : which === "second" ? 1 : 2;
		const lanesOfEvent = lanes[index]?.split("/")[1];
		assert.equal(lanesOfEvent, expected);
	},
);

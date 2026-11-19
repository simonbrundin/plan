import { describe, expect, it } from "bun:test";

import {
	planVisibility,
	type VisibilityEvent,
	type VisibilityInput,
	type VisibilitySeries,
} from "../../app/utils/calendarVisibility";

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

const series = (key: string, count: number, daily = true): VisibilitySeries => ({
	key,
	title: key,
	daily,
	count,
});

const event = (
	start: number,
	end: number,
	seriesKey: string | null = null,
): VisibilityEvent => ({ start, end, seriesKey });

const plan = (input: Partial<VisibilityInput> = {}) =>
	planVisibility({
		events: [],
		series: [],
		dailyLimit: 7,
		hidingEnabled: true,
		focusKey: null,
		hiddenBefore: [],
		from: 0,
		to: DAY,
		height: 600,
		...input,
	});

describe("planVisibility", () => {
	it("döljer en daglig serie när den har fler förekomster än gränsen", () => {
		const result = plan({
			series: [series("daily", 8)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
		});

		expect(result.hidden).toEqual(["daily"]);
	});

	it("visar en daglig serie när den har färre förekomster än gränsen", () => {
		const result = plan({
			series: [series("daily", 7)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
		});

		expect(result.hidden).toEqual([]);
	});

	it("håller en serie dold när antalet förekomster precis når gränsen", () => {
		const result = plan({
			series: [series("daily", 7)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
			hiddenBefore: ["daily"],
		});

		expect(result.hidden).toEqual(["daily"]);
	});

	it("visar serien igen när antalet faller under gränsen", () => {
		const result = plan({
			series: [series("daily", 6)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
			hiddenBefore: ["daily"],
		});

		expect(result.hidden).toEqual([]);
	});

	it("döljer ingenting när döljning är avstängd", () => {
		const result = plan({
			series: [series("daily", 30)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
			hidingEnabled: false,
		});

		expect(result.hidden).toEqual([]);
	});

	it("visar bara den serie som är vald", () => {
		const result = plan({
			series: [series("daily", 8), series("weekly", 2, false)],
			events: [
				event(9 * HOUR, 10 * HOUR, "daily"),
				event(11 * HOUR, 12 * HOUR, "weekly"),
			],
			focusKey: "weekly",
		});

		expect(result.hidden).toEqual(["daily"]);
	});

	it("döljer den serien som har flest förekomster när allt inte ryms", () => {
		const events: VisibilityEvent[] = [];

		// En serie som återkommer varje dag i tio dagar.
		for (let day = 0; day < 10; day += 1) {
			events.push(event(day * DAY + 9 * HOUR, day * DAY + 10 * HOUR, "daily"));
		}

		// En veckoserien som återkommer två gånger.
		events.push(event(11 * HOUR, 12 * HOUR, "weekly"));
		events.push(event(7 * DAY + 11 * HOUR, 7 * DAY + 12 * HOUR, "weekly"));

		const result = plan({
			series: [series("daily", 10), series("weekly", 2, false)],
			events,
			to: 10 * DAY,
			height: 600,
		});

		expect(result.hidden).toContain("daily");
		expect(result.hidden).not.toContain("weekly");
	});

	it("döljer aldrig enstaka händelser", () => {
		const events: VisibilityEvent[] = [];

		for (let hour = 0; hour < 40; hour += 1) {
			events.push(event(hour * HOUR, hour * HOUR + 30 * 60 * 1000));
		}

		const result = plan({ events });

		expect(result.hidden).toEqual([]);
		expect(result.overflowing).toBe(true);
	});

	it("döljer en serie med en enda förekomst inte", () => {
		const events: VisibilityEvent[] = [];

		for (let day = 0; day < 30; day += 1) {
			events.push(event(day * DAY + 9 * HOUR, day * DAY + 10 * HOUR, "daily"));
		}

		const result = plan({
			series: [series("daily", 30)],
			events,
			to: 30 * DAY,
			height: 600,
		});

		expect(result.hidden).toContain("daily");
	});
});

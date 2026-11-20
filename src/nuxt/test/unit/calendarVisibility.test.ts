import { describe, expect, it } from "bun:test";

import {
	planVisibility,
	type VisibilityEvent,
	type VisibilityInput,
	type VisibilitySeries,
} from "../../app/utils/calendarVisibility";

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

const series = (key: string, count: number): VisibilitySeries => ({
	key,
	title: key,
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
		maxInstances: 7,
		hidingEnabled: true,
		hiddenBefore: [],
		from: 0,
		to: DAY,
		height: 600,
		...input,
	});

describe("planVisibility", () => {
	it("döljer en serie när den har fler förekomster än gränsen", () => {
		const result = plan({
			series: [series("daily", 8)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
		});

		expect(result.hidden).toEqual(["daily"]);
	});

	it("visar en serie när den har högst så många förekomster som gränsen", () => {
		const result = plan({
			series: [series("daily", 7)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
		});

		expect(result.hidden).toEqual([]);
	});

	it("gränsen kan ändras", () => {
		const result = plan({
			series: [series("daily", 4)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
			maxInstances: 3,
		});

		expect(result.hidden).toEqual(["daily"]);
	});

	it("döljer ingenting när döljning är avstängd", () => {
		const result = plan({
			series: [series("daily", 30)],
			events: [event(9 * HOUR, 10 * HOUR, "daily")],
			hidingEnabled: false,
		});

		expect(result.hidden).toEqual([]);
	});

	it("döljer den serien som har flest förekomster när allt inte ryms", () => {
		const events: VisibilityEvent[] = [];

		// Fyrattio samtidiga förekomster ryms inte i en vy på 600 pixlar.
		for (let i = 0; i < 40; i += 1) {
			events.push(event(9 * HOUR, 9 * HOUR + 30 * 60 * 1000, "daily"));
		}

		events.push(event(11 * HOUR, 12 * HOUR, "weekly"));
		events.push(event(7 * DAY + 11 * HOUR, 7 * DAY + 12 * HOUR, "weekly"));

		const result = plan({
			series: [series("daily", 40), series("weekly", 2)],
			events,
			to: 10 * DAY,
			height: 600,
		});

		expect(result.hidden).toContain("daily");
		expect(result.hidden).not.toContain("weekly");
	});

	it("håller samma serie dold när flera är lika stora", () => {
		const events: VisibilityEvent[] = [];
		const keys = ["a", "b", "c", "d", "e"];

		// Fem serier med sju samtidiga förekomster vardera. Det ryms inte i
		// kolumnen, så någon serie måste döljas.
		for (const key of keys) {
			for (let i = 0; i < 7; i += 1) {
				events.push(event(9 * HOUR, 9 * HOUR + 30 * 60 * 1000, key));
			}
		}

		const result = plan({
			series: keys.map((key) => series(key, 7)),
			events,
			hiddenBefore: ["c"],
			height: 600,
		});

		// Serien som var dold sist döljs först, så att det är samma serie som
		// försvinner och kommer tillbaka. De övriga döljs i alfabetisk
		// ordning, så att de inte hoppar om varandra.
		expect(result.hidden).toEqual(["c", "a"]);
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

		for (let i = 0; i < 40; i += 1) {
			events.push(event(9 * HOUR, 9 * HOUR + 30 * 60 * 1000, "daily"));
		}

		const result = plan({
			series: [series("daily", 40), series("weekly", 1)],
			events,
		});

		expect(result.hidden).toContain("daily");
		expect(result.hidden).not.toContain("weekly");
	});
});

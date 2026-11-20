import { describe, expect, it } from "bun:test";

import {
	MIN_ROW_PX,
	ROW_GAP_PX,
	layoutEventRows,
} from "../../app/utils/calendarLayout";

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

/**
 * Packar en lista med [start, slut] i millisekunder.
 */
const layout = (
	items: Array<[number, number]>,
	options: {
		from?: number;
		to?: number;
		height?: number;
	} = {},
) =>
	layoutEventRows(
		items.map(([start, end]) => ({ start, end })),
		{
			from: options.from ?? 0,
			to: options.to ?? DAY,
			height: options.height ?? 600,
		},
	);

describe("layoutEventRows", () => {
	it("placerar en rad efter händelsens tid", () => {
		const rows = layout([[12 * HOUR, 13 * HOUR]]);

		// Ett dygn är 600 pixlar, så en timme är 25 pixlar.
		expect(rows[0].top).toBe(300);
		expect(rows[0].height).toBe(25);
	});

	it("ger en kort händelse minsta lästa höjd", () => {
		const rows = layout([[12 * HOUR, 13 * HOUR]], {
			to: 30 * DAY,
			height: 600,
		});

		expect(rows[0].height).toBe(MIN_ROW_PX);
	});

	it("behåller en lång händelses höjd", () => {
		const rows = layout([[9 * HOUR, 11 * HOUR]]);

		expect(rows[0].height).toBe(50);
	});

	it("begränsar en lång händelse så att den inte tar över kolumnen", () => {
		const rows = layout([[0, 20 * HOUR]], { height: 600 });

		expect(rows[0].height).toBeLessThanOrEqual(150);
	});

	it("skjuter en rad som krockar ned under den som redan står", () => {
		const rows = layout([
			[10 * HOUR, 11 * HOUR],
			[10 * HOUR, 11 * HOUR],
		]);

		// Den andra raden får inte ligga ovanpå den första. Bara
		// etikettbandet reserverar plats, så den hamnar strax under.
		expect(rows[1].top).toBe(rows[0].top + MIN_ROW_PX + ROW_GAP_PX);
	});

	it("lägger en rad under den som redan står, även utan överlapp", () => {
		const rows = layout([
			[10 * HOUR, 11 * HOUR],
			[10 * HOUR + 30 * 60 * 1000, 11 * HOUR + 30 * 60 * 1000],
		]);

		expect(rows[1].top).toBe(rows[0].top + MIN_ROW_PX + ROW_GAP_PX);
	});

	it("behåller en händelse på sin tid inne i en lång händelse", () => {
		const rows = layout([
			[9 * HOUR, 16 * HOUR],
			[12 * HOUR, 13 * HOUR],
		]);

		// 12:00 är 300 pixlar ned i en vy som är 600 pixlar hög.
		expect(rows[1].top).toBe(300);
	});

	it("behåller en händelse på sin tid efter en lång händelse", () => {
		const rows = layout([
			[9 * HOUR, 16 * HOUR],
			[18 * HOUR, 19 * HOUR],
		]);

		expect(rows[1].top).toBe(450);
	});

	it("låter etikettbanden aldrig överlappa", () => {
		const rows = layout([
			[9 * HOUR, 16 * HOUR],
			[12 * HOUR, 13 * HOUR],
			[12 * HOUR, 13 * HOUR],
			[12 * HOUR, 13 * HOUR],
		]);

		const sorted = [...rows].sort((left, right) => left.top - right.top);

		for (let i = 1; i < sorted.length; i += 1) {
			const previous = sorted[i - 1]!;
			const current = sorted[i]!;

			expect(current.top).toBeGreaterThanOrEqual(
				previous.top + Math.min(previous.height, MIN_ROW_PX),
			);
		}
	});

	it("behåller radernas ordning i tiden", () => {
		const rows = layout([
			[12 * HOUR, 13 * HOUR],
			[1 * HOUR, 2 * HOUR],
		]);

		expect(rows[0].top).toBeGreaterThan(rows[1].top);
	});

	it("flyttar en rad som börjar före fönstret till toppen", () => {
		const rows = layout([[-2 * HOUR, HOUR]]);

		// Bara delen som syns räknas, alltså en timme.
		expect(rows[0].top).toBe(0);
		expect(rows[0].height).toBe(25);
	});

	it("märker rader som inte ryms i vyn", () => {
		const rows = layout(
			Array.from({ length: 60 }, (_, index) => [
				index * 10 * 60 * 1000,
				index * 10 * 60 * 1000 + 30 * 60 * 1000,
			]) as Array<[number, number]>,
		);

		const clipped = rows.filter((row) => row.clipped);

		expect(clipped.length).toBeGreaterThan(0);
		expect(rows[0].clipped).toBe(false);
	});

	it("begränsar en rad till fönstrets höjd", () => {
		const rows = layout([[23 * HOUR, 25 * HOUR]]);

		expect(rows[0].top + rows[0].height).toBeLessThanOrEqual(600);
	});
});

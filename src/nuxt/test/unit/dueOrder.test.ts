import { describe, expect, it } from "bun:test";
import { sortByDueFirst } from "../../app/utils/dueOrder";

describe("sortByDueFirst", () => {
	it("lägger mål med deadline överst", () => {
		const sorterade = sortByDueFirst([
			{ id: 1, due: null },
			{ id: 2, due: "2026-12-01" },
			{ id: 3, due: null },
		]);

		expect(sorterade.map((mål) => mål.id)).toEqual([2, 1, 3]);
	});

	it("sorterar deadlines efter datum, den närmaste fristen först", () => {
		const sorterade = sortByDueFirst([
			{ id: 1, due: "2026-12-24" },
			{ id: 2, due: "2026-11-01" },
			{ id: 3, due: "2026-12-01" },
		]);

		expect(sorterade.map((mål) => mål.id)).toEqual([2, 3, 1]);
	});

	it("behåller den manuella ordningen för mål med samma deadline", () => {
		const sorterade = sortByDueFirst([
			{ id: 1, due: "2026-12-01" },
			{ id: 2, due: "2026-12-01" },
			{ id: 3, due: null },
		]);

		expect(sorterade.map((mål) => mål.id)).toEqual([1, 2, 3]);
	});

	it("behåller den manuella ordningen bland mål utan deadline", () => {
		const sorterade = sortByDueFirst([
			{ id: 1, due: null },
			{ id: 2, due: undefined },
			{ id: 3, due: null },
		]);

		expect(sorterade.map((mål) => mål.id)).toEqual([1, 2, 3]);
	});

	it("ändrar inte den lista som skickas in", () => {
		const mål = [
			{ id: 1, due: null },
			{ id: 2, due: "2026-12-01" },
		];

		sortByDueFirst(mål);

		expect(mål.map((m) => m.id)).toEqual([1, 2]);
	});
});

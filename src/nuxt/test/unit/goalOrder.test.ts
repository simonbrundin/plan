import { describe, expect, it } from "bun:test";
import { compareGoalOrder, sortGoalsByOrder } from "../../app/utils/goalOrder";

describe("sortGoalsByOrder", () => {
	it("sorterar grundmål efter order", () => {
		const goals = [
			{ id: 1, order: 2, created: "2026-01-01T00:00:00Z" },
			{ id: 2, order: 0, created: "2026-01-02T00:00:00Z" },
			{ id: 3, order: 1, created: "2026-01-03T00:00:00Z" },
		];
		expect(sortGoalsByOrder(goals).map((g) => g.id)).toEqual([2, 3, 1]);
	});

	it("lägger mål med samma order sist, nyast först", () => {
		const goals = [
			{ id: 1, order: 0, created: "2026-01-01T00:00:00Z" },
			{ id: 2, order: 0, created: "2026-01-03T00:00:00Z" },
			{ id: 3, order: 0, created: "2026-01-02T00:00:00Z" },
		];
		expect(sortGoalsByOrder(goals).map((g) => g.id)).toEqual([2, 3, 1]);
	});

	it("räknar saknad order som noll", () => {
		const goals = [
			{ id: 1, created: "2026-01-01T00:00:00Z" },
			{ id: 2, order: 1, created: "2026-01-02T00:00:00Z" },
		];
		expect(sortGoalsByOrder(goals).map((g) => g.id)).toEqual([1, 2]);
	});

	it("ändrar inte ursprungsarrayen", () => {
		const goals = [
			{ id: 1, order: 1, created: "2026-01-01T00:00:00Z" },
			{ id: 2, order: 0, created: "2026-01-02T00:00:00Z" },
		];
		sortGoalsByOrder(goals);
		expect(goals.map((g) => g.id)).toEqual([1, 2]);
	});
});

describe("compareGoalOrder", () => {
	it("jämför order före skapadatum", () => {
		const older = { order: 1, created: "2026-01-01T00:00:00Z" };
		const newer = { order: 0, created: "2026-01-05T00:00:00Z" };
		expect(compareGoalOrder(older, newer)).toBeGreaterThan(0);
	});
});

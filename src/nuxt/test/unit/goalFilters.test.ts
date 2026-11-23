import { describe, expect, it } from "bun:test";
import {
	isRootGoal,
	isUnfinishedGoal,
	matchesNowFilter,
	matchesStartedFilter,
} from "../../app/utils/goalFilters";

describe("matchesStartedFilter", () => {
	it("visar alla mål när filtret är 'all'", () => {
		expect(matchesStartedFilter({ started: null }, "all")).toBe(true);
		expect(
			matchesStartedFilter({ started: "2026-01-01T10:00:00Z" }, "all"),
		).toBe(true);
	});

	it("visar endast mål med startdatum när filtret är 'started'", () => {
		expect(
			matchesStartedFilter({ started: "2026-01-01T10:00:00Z" }, "started"),
		).toBe(true);
		expect(matchesStartedFilter({ started: null }, "started")).toBe(false);
	});

	it("räknar saknat startfält som inte påbörjat", () => {
		expect(matchesStartedFilter({}, "started")).toBe(false);
		expect(matchesStartedFilter({ started: undefined }, "started")).toBe(false);
		expect(matchesStartedFilter({}, "not_started")).toBe(true);
		expect(matchesStartedFilter({ started: undefined }, "not_started")).toBe(
			true,
		);
	});

	it("visar endast mål utan startdatum när filtret är 'not_started'", () => {
		expect(matchesStartedFilter({ started: null }, "not_started")).toBe(true);
		expect(
			matchesStartedFilter({ started: "2026-01-01T10:00:00Z" }, "not_started"),
		).toBe(false);
	});
});

describe("matchesNowFilter", () => {
	it("visar endast mål med nu-märke när filtret är 'now'", () => {
		expect(matchesNowFilter({ now: true }, "now")).toBe(true);
		expect(matchesNowFilter({ now: false }, "now")).toBe(false);
		expect(matchesNowFilter({}, "now")).toBe(false);
	});

	it("visar mål med nu-märke även om de inte är påbörjade", () => {
		expect(matchesNowFilter({ now: true, started: null }, "now")).toBe(true);
	});

	it("faller tillbaka på startfiltret för övriga filter", () => {
		expect(matchesNowFilter({ now: true, started: null }, "all")).toBe(true);
		expect(
			matchesNowFilter({ now: true, started: "2026-01-01T10:00:00Z" }, "started"),
		).toBe(true);
		expect(matchesNowFilter({ now: false, started: null }, "not_started")).toBe(
			true,
		);
	});
});

describe("isRootGoal", () => {
	it("räknar mål utan förälder som grundmål", () => {
		expect(isRootGoal({ has_parent: false })).toBe(true);
	});

	it("räknar mål med förälder som undermål", () => {
		expect(isRootGoal({ has_parent: true })).toBe(false);
	});

	it("räknar mål utan has_parent som grundmål", () => {
		expect(isRootGoal({})).toBe(true);
		expect(isRootGoal({ has_parent: undefined })).toBe(true);
	});
});

describe("isUnfinishedGoal", () => {
	it("räknar mål utan finished som öppna", () => {
		expect(isUnfinishedGoal({ finished: null })).toBe(true);
		expect(isUnfinishedGoal({ finished: undefined })).toBe(true);
	});

	it("räknar slutförda mål som stängda", () => {
		expect(isUnfinishedGoal({ finished: "2026-01-01T10:00:00Z" })).toBe(false);
	});
});

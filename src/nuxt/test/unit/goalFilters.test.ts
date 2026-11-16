import { describe, expect, it } from "bun:test";
import { isRootGoal } from "../../app/utils/goalFilters";

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

import { describe, expect, it } from "bun:test";
import { canBeFinished } from "../../app/utils/goalRules";

describe("canBeFinished", () => {
	it("tillåter att vanliga mål slutförs", () => {
		expect(canBeFinished({ eternal: false })).toBe(true);
	});

	it("förbjuder att eviga mål slutförs", () => {
		expect(canBeFinished({ eternal: true })).toBe(false);
	});

	it("tillåter mål utan evigt-flagga", () => {
		expect(canBeFinished({})).toBe(true);
		expect(canBeFinished({ eternal: null })).toBe(true);
	});
});

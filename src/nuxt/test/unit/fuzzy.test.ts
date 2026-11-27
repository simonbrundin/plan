import { describe, expect, it } from "bun:test";
import { fuzzyMatch, fuzzyScore } from "../../app/utils/fuzzy";

describe("fuzzyMatch", () => {
	it("matchar en enkel substring", () => {
		expect(fuzzyMatch("möt", "Mötet i rummet")).not.toBeNull();
	});

	it("matchar tecken i valfri ordning, som fzf", () => {
		expect(fuzzyMatch("mt", "Mötet i rummet")).not.toBeNull();
		expect(fuzzyMatch("tm", "Mötet i rummet")).not.toBeNull();
	});

	it("matchar inte tecken som saknas", () => {
		expect(fuzzyMatch("mk", "Läkare")).toBeNull();
		expect(fuzzyMatch("xyz", "Mötet i rummet")).toBeNull();
	});

	it("matchar diagonalt och bortser från diacriter", () => {
		expect(fuzzyMatch("mal", "Mål")).not.toBeNull();
		expect(fuzzyMatch("mål", "Mal")).not.toBeNull();
	});

	it("kräver att alla ord i sökordet finns", () => {
		expect(fuzzyMatch("mötet rum", "Mötet i rummet")).not.toBeNull();
		expect(fuzzyMatch("mötet städa", "Mötet i rummet")).toBeNull();
	});

	it("rankar prefix högre än en träff mitt i texten", () => {
		const prefix = fuzzyScore("möt", "Mötet i rummet") ?? 0;
		const middle = fuzzyScore("möt", "Städa till möten") ?? 0;

		expect(prefix).toBeGreaterThan(middle);
	});

	it("rankar ordgränser högre än en träff inne i ett ord", () => {
		const boundary = fuzzyScore("rum", "x rum") ?? 0;
		const inside = fuzzyScore("rum", "xbrum") ?? 0;

		expect(boundary).toBeGreaterThan(inside);
	});

	it("rankar sammanhängande träffar högre", () => {
		const together = fuzzyScore("möt", "Mötet") ?? 0;
		const apart = fuzzyScore("möt", "M t") ?? 0;

		expect(together).toBeGreaterThan(apart);
	});

	it("kräver versaler i sökordet när det själv har versaler", () => {
		expect(fuzzyMatch("MÖT", "mötet")).toBeNull();
		expect(fuzzyMatch("möt", "mötet")).not.toBeNull();
	});

	it("ett tomt sökord matchar allt med poängen noll", () => {
		expect(fuzzyMatch("", "Mötet")).toEqual({ score: 0, positions: [] });
	});
});

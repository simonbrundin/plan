import { describe, expect, it } from "bun:test";
import { descriptionPreview } from "../../app/utils/goalDescription";

describe("descriptionPreview", () => {
	it("returnerar tom sträng när det saknas beskrivning", () => {
		expect(descriptionPreview(undefined)).toBe("");
		expect(descriptionPreview(null)).toBe("");
		expect(descriptionPreview("")).toBe("");
	});

	it("tar bort radbrytningar och upprepade mellanrum", () => {
		expect(descriptionPreview("  första raden\n\nandra raden  ")).toBe(
			"första raden andra raden",
		);
	});

	it("kortar ned långa beskrivningar med en ellipsis", () => {
		const preview = descriptionPreview("a".repeat(200), 10);
		expect(preview).toBe(`${"a".repeat(9)}…`);
	});

	it("behåller korta beskrivningar oförändrade", () => {
		expect(descriptionPreview("kort", 10)).toBe("kort");
	});
});

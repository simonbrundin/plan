import { describe, expect, it } from "bun:test";
import { moveSearchHighlight } from "../../app/utils/searchHighlight";

describe("moveSearchHighlight", () => {
	it("returnerar -1 när listan är tom", () => {
		expect(moveSearchHighlight(-1, 0, 1)).toBe(-1);
		expect(moveSearchHighlight(2, 0, -1)).toBe(-1);
	});

	it("markerar första träffen när nedåt trycks utan markering", () => {
		expect(moveSearchHighlight(-1, 3, 1)).toBe(0);
	});

	it("markerar sista träffen när uppåt trycks utan markering", () => {
		expect(moveSearchHighlight(-1, 3, -1)).toBe(2);
	});

	it("flyttar markeringen nedåt och uppåt", () => {
		expect(moveSearchHighlight(0, 3, 1)).toBe(1);
		expect(moveSearchHighlight(1, 3, -1)).toBe(0);
	});

	it("stannar inom listan", () => {
		expect(moveSearchHighlight(2, 3, 1)).toBe(2);
		expect(moveSearchHighlight(0, 3, -1)).toBe(0);
	});
});

import { describe, expect, it } from "bun:test";
import { moveChildId, nextChildOrder } from "../../app/utils/childOrder";

describe("moveChildId", () => {
	it("flyttar ett undermål ett steg upp", () => {
		const all = [1, 2, 3];
		expect(moveChildId(all, [1, 2, 3], 3, "up")).toEqual([1, 3, 2]);
	});

	it("flyttar ett undermål ett steg ner", () => {
		const all = [1, 2, 3];
		expect(moveChildId(all, [1, 2, 3], 1, "down")).toEqual([2, 1, 3]);
	});

	it("flyttar ett undermål flera steg", () => {
		const all = [1, 2, 3, 4];
		expect(moveChildId(all, [1, 2, 3, 4], 1, "up")).toBeNull();
		expect(moveChildId(all, [1, 2, 3, 4], 4, "down")).toBeNull();
		expect(moveChildId(all, [1, 2, 3, 4], 2, "up")).toEqual([2, 1, 3, 4]);
	});

	it("behåller dolda undermål på sina platser", () => {
		// 2 är dolt (klart) och visas inte i listan
		const all = [1, 2, 3];
		const visible = [1, 3];
		expect(moveChildId(all, visible, 3, "up")).toEqual([3, 2, 1]);
	});

	it("returnerar null när målet inte finns", () => {
		expect(moveChildId([1, 2], [1, 2], 99, "up")).toBeNull();
	});
});

describe("nextChildOrder", () => {
	it("ger nästa lediga ordning", () => {
		expect(nextChildOrder([0, 1, 2])).toBe(3);
		expect(nextChildOrder([])).toBe(0);
		expect(nextChildOrder([3, 7])).toBe(8);
	});
});

import { describe, expect, it } from "bun:test";
import {
	insertChildId,
	moveChildId,
	newChildInsertIndex,
	nextChildOrder,
} from "../../app/utils/childOrder";

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

describe("newChildInsertIndex", () => {
	it("lägger det nya målet under det markerade undermålet", () => {
		expect(newChildInsertIndex([1, 2, 3], [1, 2, 3], 0, false)).toBe(1);
		expect(newChildInsertIndex([1, 2, 3], [1, 2, 3], 1, false)).toBe(2);
		expect(newChildInsertIndex([1, 2, 3], [1, 2, 3], 2, false)).toBe(3);
	});

	it("lägger det nya målet först när huvudmålet är markerat", () => {
		expect(newChildInsertIndex([1, 2, 3], [1, 2, 3], 0, true)).toBe(0);
	});

	it("räknar rätt när listan är filtrerad", () => {
		// 2 är dolt (klart) och visas inte i listan
		expect(newChildInsertIndex([1, 2, 3], [1, 3], 1, false)).toBe(3);
		expect(newChildInsertIndex([1, 2, 3], [1, 3], 0, false)).toBe(1);
	});

	it("lägger målet sist när inget undermål är markerat", () => {
		expect(newChildInsertIndex([1, 2, 3], [1, 2, 3], -1, false)).toBe(3);
	});
});

describe("insertChildId", () => {
	it("insätter det nya målet på given plats", () => {
		expect(insertChildId([1, 2, 3], 1, 9)).toEqual([1, 9, 2, 3]);
		expect(insertChildId([1, 2, 3], 0, 9)).toEqual([9, 1, 2, 3]);
		expect(insertChildId([1, 2, 3], 3, 9)).toEqual([1, 2, 3, 9]);
	});

	it("klampar in indexet inom listan", () => {
		expect(insertChildId([1, 2, 3], 99, 9)).toEqual([1, 2, 3, 9]);
		expect(insertChildId([1, 2, 3], -5, 9)).toEqual([9, 1, 2, 3]);
	});

	it("muterar inte originalet", () => {
		const all = [1, 2, 3];
		insertChildId(all, 1, 9);
		expect(all).toEqual([1, 2, 3]);
	});
});

describe("nextChildOrder", () => {
	it("ger nästa lediga ordning", () => {
		expect(nextChildOrder([0, 1, 2])).toBe(3);
		expect(nextChildOrder([])).toBe(0);
		expect(nextChildOrder([3, 7])).toBe(8);
	});
});

import { describe, expect, it } from "bun:test";
import {
	buildDisplayTree,
	buildHierarchy,
	swapIds,
} from "../../app/utils/nowHierarchy";

describe("buildHierarchy", () => {
	it("lägger mål under den första föräldern som finns i listan", () => {
		const hierarchy = buildHierarchy([
			{ id: 1, parents: [] },
			{ id: 2, parents: [{ id: 1 }] },
			{ id: 3, parents: [{ id: 2 }] },
		]);

		expect(hierarchy.parentOf.get(1)).toBeNull();
		expect(hierarchy.parentOf.get(2)).toBe(1);
		expect(hierarchy.parentOf.get(3)).toBe(2);
		expect(hierarchy.roots).toEqual([1]);
	});

	it("bortser från föräldrar som inte finns i listan", () => {
		const hierarchy = buildHierarchy([
			{ id: 1, parents: [{ id: 99 }] },
		]);

		expect(hierarchy.parentOf.get(1)).toBeNull();
		expect(hierarchy.roots).toEqual([1]);
	});

	it("tar den tyngsta föräldern när målet har flera", () => {
		const hierarchy = buildHierarchy([
			{ id: 1, parents: [] },
			{ id: 2, parents: [] },
			{ id: 3, parents: [{ id: 2 }, { id: 1 }] },
		]);

		// Föräldrarna kommer sorterade på vikt - den första används.
		expect(hierarchy.parentOf.get(3)).toBe(2);
	});

	it("bryter cykler så att målen inte försvinner", () => {
		const hierarchy = buildHierarchy([
			{ id: 1, parents: [{ id: 2 }] },
			{ id: 2, parents: [{ id: 1 }] },
		]);

		const roots = hierarchy.roots;
		expect(roots.length).toBeGreaterThan(0);
		expect(hierarchy.parentOf.get(roots[0])).toBeNull();
	});
});

describe("buildDisplayTree", () => {
	it("visar barn under föräldern", () => {
		const goals = [
			{ id: 1, parents: [] },
			{ id: 2, parents: [{ id: 1 }] },
			{ id: 3, parents: [] },
		];
		const hierarchy = buildHierarchy(goals);
		const tree = buildDisplayTree([1, 2, 3], hierarchy);

		expect(tree.rows.map((row) => row.id)).toEqual([1, 2, 3]);
		expect(tree.rows.map((row) => row.depth)).toEqual([0, 1, 0]);
		expect(tree.rows[1].parentId).toBe(1);
	});

	it("lyfter upp barn vars förälder inte syns", () => {
		const goals = [
			{ id: 1, parents: [] },
			{ id: 2, parents: [{ id: 1 }] },
		];
		const hierarchy = buildHierarchy(goals);
		// Föräldern filtreras bort - barnet hamnar på toppnivå.
		const tree = buildDisplayTree([2], hierarchy);

		expect(tree.rows.map((row) => row.id)).toEqual([2]);
		expect(tree.rows[0].depth).toBe(0);
		expect(tree.rows[0].parentId).toBeNull();
	});

	it("grupperar syskon under samma förälder", () => {
		const goals = [
			{ id: 1, parents: [] },
			{ id: 2, parents: [{ id: 1 }] },
			{ id: 3, parents: [{ id: 1 }] },
			{ id: 4, parents: [] },
		];
		const hierarchy = buildHierarchy(goals);
		const tree = buildDisplayTree([1, 2, 3, 4], hierarchy);

		expect(tree.siblingsOf.get(1)).toEqual([1, 4]);
		expect(tree.siblingsOf.get(2)).toEqual([2, 3]);
		expect(tree.siblingsOf.get(3)).toEqual([2, 3]);
	});

	it("visar barnbarn under sin förälder", () => {
		const goals = [
			{ id: 1, parents: [] },
			{ id: 2, parents: [{ id: 1 }] },
			{ id: 3, parents: [{ id: 2 }] },
		];
		const hierarchy = buildHierarchy(goals);
		const tree = buildDisplayTree([1, 2, 3], hierarchy);

		expect(tree.rows.map((row) => row.depth)).toEqual([0, 1, 2]);
		expect(tree.rows[2].parentId).toBe(2);
	});
});

describe("swapIds", () => {
	it("byter plats på två mål", () => {
		expect(swapIds([1, 2, 3], 1, 3)).toEqual([3, 2, 1]);
	});

	it("returnerar null om ett mål inte finns", () => {
		expect(swapIds([1, 2, 3], 1, 9)).toBeNull();
	});
});

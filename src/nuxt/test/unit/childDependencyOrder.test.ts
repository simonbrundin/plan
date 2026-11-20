import { describe, expect, it } from "bun:test";
import {
	buildDependencyTree,
	moveChildInTree,
} from "../../app/utils/childDependencyOrder";

function child(
	id: number,
	dependsOn: number[] = [],
	finished: string | null = null,
) {
	return {
		id,
		finished,
		dependsOn: dependsOn.map((depId) => ({ id: depId })),
	};
}

function shape(children: ReturnType<typeof child>[]) {
	return buildDependencyTree(children).map((item) => [
		item.child.id,
		item.depth,
		item.parentId,
	]);
}

describe("buildDependencyTree", () => {
	it("lägger ett mål som väntar under det mål det väntar på", () => {
		expect(shape([child(2, [1]), child(1)])).toEqual([
			[1, 0, null],
			[2, 1, 1],
		]);
	});

	it("hänger ihop en kedja av beroenden", () => {
		expect(shape([child(3, [2]), child(2, [1]), child(1)])).toEqual([
			[1, 0, null],
			[2, 1, 1],
			[3, 2, 2],
		]);
	});

	it("lägger flera mål under samma beroende", () => {
		expect(shape([child(1), child(2, [1]), child(3, [1])])).toEqual([
			[1, 0, null],
			[2, 1, 1],
			[3, 1, 1],
		]);
	});

	it("ignorerar beroenden till mål utanför listan", () => {
		expect(shape([child(2, [99]), child(1)])).toEqual([
			[2, 0, null],
			[1, 0, null],
		]);
	});

	it("ignorerar beroenden på redan klara mål", () => {
		expect(
			shape([child(2, [1]), child(1, [], "2026-01-01T00:00:00Z")]),
		).toEqual([
			[2, 0, null],
			[1, 0, null],
		]);
	});

	it("väljer det beroende som ligger längst ned i kedjan", () => {
		// Mål 3 väntar på 1 och 2, där 2 väntar på 1 - då hänger 3 under 2.
		expect(shape([child(1), child(2, [1]), child(3, [1, 2])])).toEqual([
			[1, 0, null],
			[2, 1, 1],
			[3, 2, 2],
		]);
	});

	it("hanterar cykler utan att hänga", () => {
		const tree = shape([child(1, [2]), child(2, [1])]);
		expect(tree.map(([id]) => id)).toEqual([1, 2]);
		// Cykeln bryts - det ena målet blir rot.
		expect(tree[0]).toEqual([1, 0, null]);
	});

	it("behåller ordningen mellan mål utan beroenden", () => {
		expect(shape([child(3), child(1), child(2)])).toEqual([
			[3, 0, null],
			[1, 0, null],
			[2, 0, null],
		]);
	});
});

describe("moveChildInTree", () => {
	const children = [child(1), child(2, [1]), child(3, [1]), child(4)];

	it("flyttar ett mål på översta nivån tillsammans med gruppen", () => {
		const tree = buildDependencyTree(children);
		const allIds = children.map((c) => c.id);

		const result = moveChildInTree(allIds, tree, 1, "down");

		expect(result).toEqual([4, 1, 2, 3]);
	});

	it("flyttar ett undermål inom sin grupp", () => {
		const tree = buildDependencyTree(children);
		const allIds = children.map((c) => c.id);

		const result = moveChildInTree(allIds, tree, 2, "down");

		expect(result).toEqual([1, 3, 2, 4]);
	});

	it("flyttar inte ett undermål ur sin grupp", () => {
		const tree = buildDependencyTree(children);
		const allIds = children.map((c) => c.id);

		// Mål 3 är sist i sin grupp - det finns ingen plats under.
		expect(moveChildInTree(allIds, tree, 3, "down")).toBeNull();
	});

	it("flyttar inte det första undermålet i gruppen över gruppen", () => {
		const tree = buildDependencyTree(children);
		const allIds = children.map((c) => c.id);

		expect(moveChildInTree(allIds, tree, 2, "up")).toBeNull();
	});

	it("flyttar inte det sista målet i listan nedåt", () => {
		const tree = buildDependencyTree(children);
		const allIds = children.map((c) => c.id);

		expect(moveChildInTree(allIds, tree, 4, "down")).toBeNull();
	});
});

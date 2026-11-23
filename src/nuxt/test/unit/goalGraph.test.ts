import { describe, expect, it } from "bun:test";
import { buildGoalData } from "../../app/utils/goalGraph";
import type { Goal, GoalDependency, GoalRelation } from "../../app/types/goal";

function goal(id: number, title: string): Goal {
	return {
		id,
		title,
		description: "",
		icon: "",
		status_id: 1,
		created: "2026-01-01T00:00:00Z",
		started: null,
		finished: null,
		due: null,
		inbox: 0,
		eternal: false,
	};
}

const goals: Goal[] = [
	goal(1, "Huvudmål"),
	goal(2, "Undermål A"),
	goal(3, "Undermål B"),
	goal(4, "Förälder"),
	goal(5, "Beroende"),
];

const relations: GoalRelation[] = [
	{ parent_id: 1, child_id: 3, order: 1, weight: 5 },
	{ parent_id: 1, child_id: 2, order: 0, weight: 10 },
	{ parent_id: 4, child_id: 1, order: 0, weight: 0 },
];

const dependencies: GoalDependency[] = [
	{ id: 1, goal_id: 2, depends_on_id: 5, created: "2026-01-01T00:00:00Z" },
];

const graph = { goals, relations, dependencies };

describe("buildGoalData", () => {
	it("returnerar null för ett mål som inte finns", () => {
		expect(buildGoalData(graph, 99)).toBeNull();
	});

	it("bygger barnen i ordning", () => {
		const data = buildGoalData(graph, 1);

		expect(data?.children.map((child) => child.id)).toEqual([2, 3]);
		expect(data?.children.map((child) => child.order)).toEqual([0, 1]);
		expect(data?.children.map((child) => child.weight)).toEqual([10, 5]);
	});

	it("bygger föräldrarna", () => {
		const data = buildGoalData(graph, 1);

		expect(data?.parents.map((parent) => parent.id)).toEqual([4]);
		expect(data?.goal.parentRelations.map((parent) => parent.id)).toEqual([4]);
	});

	it("bygger beroenden åt båda håll", () => {
		const data = buildGoalData(graph, 2);

		expect(data?.dependsOn.map((g) => g.id)).toEqual([5]);
		expect(data?.blocking.map((g) => g.id)).toEqual([]);
	});

	it("bygger barnens beroenden", () => {
		const data = buildGoalData(graph, 1);

		expect(data?.children[0]?.dependsOn.map((g) => g.id)).toEqual([5]);
		expect(data?.children[1]?.dependsOn.map((g) => g.id)).toEqual([]);
	});

	it("bygger blockerande mål", () => {
		const data = buildGoalData(graph, 5);

		expect(data?.blocking.map((g) => g.id)).toEqual([2]);
	});

	it("bygger relationerna till barnen", () => {
		const data = buildGoalData(graph, 1);

		expect(data?.goal.childRelations).toEqual([
			{ child_id: 2, order: 0, weight: 10 },
			{ child_id: 3, order: 1, weight: 5 },
		]);
	});
});

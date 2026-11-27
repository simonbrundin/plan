import { describe, expect, it } from "bun:test";
import {
	GOAL_PHASES,
	getGoalPhasePresentation,
	nextPhase,
} from "../../app/utils/goalPhase";

describe("goal phase presentation", () => {
	it("defines a distinct icon and color for every phase", () => {
		const presentations = GOAL_PHASES.map(getGoalPhasePresentation);

		expect(presentations.map(({ name }) => name)).toEqual([
			"define",
			"plan",
			"execute",
			"verify",
			"completed",
		]);
		expect(new Set(presentations.map(({ icon }) => icon)).size).toBe(5);
		expect(new Set(presentations.map(({ colorClass }) => colorClass)).size).toBe(5);
	});

	it("returns a neutral presentation for an unknown phase", () => {
		expect(getGoalPhasePresentation("unknown")).toEqual({
			name: "unknown",
			label: "Okänd fas",
			icon: "lucide:circle-dot",
			colorClass: "text-ink-dim",
		});
	});
});

describe("next phase", () => {
	it("steps forward through the phases and wraps around", () => {
		expect(nextPhase("define")).toBe("plan");
		expect(nextPhase("plan")).toBe("execute");
		expect(nextPhase("execute")).toBe("verify");
		expect(nextPhase("verify")).toBe("completed");
		expect(nextPhase("completed")).toBe("define");
	});

	it("steps backward through the phases and wraps around", () => {
		expect(nextPhase("plan", -1)).toBe("define");
		expect(nextPhase("define", -1)).toBe("completed");
	});

	it("starts at the first phase for an unknown phase", () => {
		expect(nextPhase("unknown")).toBe("plan");
		expect(nextPhase(null)).toBe("plan");
		expect(nextPhase(undefined, -1)).toBe("completed");
	});
});

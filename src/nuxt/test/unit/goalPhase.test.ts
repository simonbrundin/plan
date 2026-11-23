import { describe, expect, it } from "bun:test";
import { GOAL_PHASES, getGoalPhasePresentation } from "../../app/utils/goalPhase";

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

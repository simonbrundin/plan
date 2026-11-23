export const GOAL_PHASES = [
	"define",
	"plan",
	"execute",
	"verify",
	"completed",
] as const;

export type GoalPhase = (typeof GOAL_PHASES)[number];

export interface GoalPhasePresentation {
	name: string;
	label: string;
	icon: string;
	colorClass: string;
}

const phasePresentations: Record<GoalPhase, GoalPhasePresentation> = {
	define: {
		name: "define",
		label: "Definiera",
		icon: "lucide:file-pen",
		colorClass: "text-sky-400",
	},
	plan: {
		name: "plan",
		label: "Planera",
		icon: "lucide:clipboard-list",
		colorClass: "text-violet-400",
	},
	execute: {
		name: "execute",
		label: "Genomföra",
		icon: "lucide:hammer",
		colorClass: "text-amber-400",
	},
	verify: {
		name: "verify",
		label: "Verifiera",
		icon: "lucide:shield-check",
		colorClass: "text-cyan-400",
	},
	completed: {
		name: "completed",
		label: "Klar",
		icon: "lucide:circle-check",
		colorClass: "text-emerald-400",
	},
};

export function getGoalPhasePresentation(
	phase: string | null | undefined,
): GoalPhasePresentation {
	if (phase && Object.hasOwn(phasePresentations, phase)) {
		return phasePresentations[phase as GoalPhase];
	}

	return {
		name: phase ?? "",
		label: "Okänd fas",
		icon: "lucide:circle-dot",
		colorClass: "text-ink-dim",
	};
}

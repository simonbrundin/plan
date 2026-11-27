/**
 * Verifiering och bevis.
 *
 * Planen är valfri - ett mål utan plan fungerar precis som ett mål med
 * plan. därför är allt här oförarligt: en typ som inte finns i listan
 * ska inte krascha vyn, bara visas som okänd.
 */

export const VERIFICATION_TYPES = [
	"checklista",
	"delmal",
	"kalender",
	"bevis",
	"matvarde",
	"kod",
] as const;

export type VerificationType = (typeof VERIFICATION_TYPES)[number];

export const VERIFICATION_STRENGTHS = ["stark", "svag"] as const;

export type VerificationStrength = (typeof VERIFICATION_STRENGTHS)[number];

export interface VerificationPlan {
	id: number;
	goal_id: number;
	type: string;
	criterion: string;
	strength: string;
	outcome?: string;
	decided_by?: number | null;
	decided_at?: string | null;
}

export interface GoalEvidence {
	id: number;
	goal_id: number;
	verification_id?: number | null;
	type: string;
	value: string;
	source?: string;
	created_at: string;
}

export interface VerificationResult {
	id: number;
	goal_id: number;
	verification_id?: number | null;
	outcome: string;
	note?: string;
	decided_by?: number | null;
	decided_at?: string | null;
	created_at: string;
}

export interface VerificationPresentation {
	label: string;
	icon: string;
	colorClass: string;
}

const typePresentations: Record<VerificationType, VerificationPresentation> = {
	checklista: {
		label: "Checklista",
		icon: "lucide:list-checks",
		colorClass: "text-sky-400",
	},
	delmal: {
		label: "Delmål",
		icon: "lucide:git-branch",
		colorClass: "text-violet-400",
	},
	kalender: {
		label: "Kalender",
		icon: "lucide:calendar-check",
		colorClass: "text-amber-400",
	},
	bevis: {
		label: "Bevis",
		icon: "lucide:paperclip",
		colorClass: "text-cyan-400",
	},
	matvarde: {
		label: "Mätvärde",
		icon: "lucide:gauge",
		colorClass: "text-teal-400",
	},
	kod: {
		label: "Kod",
		icon: "lucide:code",
		colorClass: "text-emerald-400",
	},
};

const strengthPresentations: Record<VerificationStrength, VerificationPresentation> = {
	stark: {
		label: "Starkt bevis",
		icon: "lucide:shield-check",
		colorClass: "text-emerald-400",
	},
	svag: {
		label: "Svagt bevis",
		icon: "lucide:shield-question",
		colorClass: "text-ink-dim",
	},
};

const outcomePresentations: Record<string, VerificationPresentation> = {
	passed: {
		label: "Godkänd",
		icon: "lucide:circle-check",
		colorClass: "text-emerald-400",
	},
	failed: {
		label: "Underkänd",
		icon: "lucide:circle-x",
		colorClass: "text-rose-400",
	},
	avslagen: {
		label: "Avslaget",
		icon: "lucide:circle-x",
		colorClass: "text-error",
	},
	pending: {
		label: "Väntar på bevis",
		icon: "lucide:circle-dashed",
		colorClass: "text-ink-dim",
	},
};

/**
 * Presentationen för en verifieringstyp.
 *
 * En okänd typ får en generisk presentation - den ska inte krascha
 * vyn, bara tala om att typen inte känns igen.
 */
export function getVerificationTypePresentation(
	type: string | null | undefined,
): VerificationPresentation {
	if (type && Object.hasOwn(typePresentations, type)) {
		return typePresentations[type as VerificationType] as VerificationPresentation;
	}

	return {
		label: type ? `Okänd typ (${type})` : "Ingen plan",
		icon: "lucide:circle-help",
		colorClass: "text-ink-dim",
	};
}

/** Presentationen för bevisets styrka. */
export function getVerificationStrengthPresentation(
	strength: string | null | undefined,
): VerificationPresentation {
	if (strength && Object.hasOwn(strengthPresentations, strength)) {
		return strengthPresentations[strength as VerificationStrength] as VerificationPresentation;
	}

	return {
		label: "Okänd styrka",
		icon: "lucide:circle-help",
		colorClass: "text-ink-dim",
	};
}

/** Presentationen för ett verifieringsomdöme. */
export function getVerificationOutcomePresentation(
	outcome: string | null | undefined,
): VerificationPresentation {
	if (outcome && Object.hasOwn(outcomePresentations, outcome)) {
		return outcomePresentations[outcome] as VerificationPresentation;
	}

	return {
		label: "Inget omdöme",
		icon: "lucide:circle-dashed",
		colorClass: "text-ink-dim",
	};
}

/**
 * Hur målet avslutades.
 *
 * "deklarerat" betyder att en människa sa att målet var klart, utan att
 * något verifierades. "verifierat" betyder att en godkänd verifiering
 * gått igenom. Skillnaden är det enda som skiljer "jag tror det är
 * klart" från "det här är bevisat".
 */
export function describeCompletion(
	kind: string | null | undefined,
): { label: string; description: string } | null {
	if (kind === "verifierat") {
		return {
			label: "Verifierat",
			description: "Avslutades efter en godkänd verifiering.",
		};
	}
	if (kind === "deklarerat") {
		return {
			label: "Deklarerat",
			description: "Avslutades av en människa, utan verifiering.",
		};
	}
	return null;
}

/**
 * Är ett omdöme väntande på en människa?
 *
 * Ett omdöme som ingen godkänt är inte giltigt - så länge det väntar
 * ska gränssnittet be om en människa.
 */
export function needsApproval(result: VerificationResult | null): boolean {
	if (!result) return false;
	return !result.decided_at;
}

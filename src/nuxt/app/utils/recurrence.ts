import type { Goal, RecurrenceMode } from "../types/goal";
import { parseDueDate } from "./dueDate.ts";

// Schemaläggningsstrategi för ett återkommande mål.
export type { RecurrenceMode };

const WEEKDAYS: Record<string, string> = {
	MO: "måndag",
	TU: "tisdag",
	WE: "onsdag",
	TH: "torsdag",
	FR: "fredag",
	SA: "lördag",
	SU: "söndag",
};

const ORDINALS: Record<number, string> = {
	2: "annan",
	3: "tredje",
	4: "fjärde",
	5: "femte",
	6: "sjätte",
	7: "sjunde",
	8: "åttonde",
	9: "nionde",
	10: "tionde",
};

/**
 * Avgör om ett mål är återkommande, dvs. har en upprepningsregel.
 */
export function isRecurringGoal(
	goal: Pick<Goal, "recurrence_rule">,
): boolean {
	return Boolean(goal.recurrence_rule);
}

/**
 * Avgör om ett återkommande mål väntar på nästa upprepning.
 *
 * Det gäller när nästa förekomst ligger i framtiden - då finns inget att
 * göra ännu, men målet ligger kvar i listan. Ett mål med en förfallen
 * förekomst är inte väntande - det är förfallet och ska synas.
 */
export function isWaitingForNextOccurrence(
	goal: Pick<Goal, "recurrence_rule" | "recurrence_next">,
): boolean {
	if (!isRecurringGoal(goal)) return false;
	if (!goal.recurrence_next) return false;

	const next = parseDueDate(goal.recurrence_next);
	if (!next) return false;

	const today = new Date();
	today.setHours(0, 0, 0, 0);

	return next.getTime() > today.getTime();
}

/**
 * Antalet slutförda upprepningar för ett mål.
 */
export function recurrenceCount(
	goal: Pick<Goal, "recurrence_count">,
): number {
	return goal.recurrence_count ?? 0;
}

/**
 * Beskriver en upprepningsregel på svenska, t.ex. "varannan dag".
 *
 * Okända regler visas som de är, så att inget går förlorat.
 */
export function describeGoalRecurrence(
	rule: string | null | undefined,
): string {
	if (!rule) return "";

	const parts = new Map<string, string>();
	for (const part of rule.replace(/^RRULE:/i, "").split(";")) {
		const [key, value] = part.split("=");
		if (key && value) {
			parts.set(key.trim().toUpperCase(), value.trim().toUpperCase());
		}
	}

	const frequency = parts.get("FREQ");
	const interval = Number(parts.get("INTERVAL") ?? "1") || 1;

	switch (frequency) {
		case "DAILY":
			return interval === 1 ? "varje dag" : `var ${ordinal(interval)} dag`;
		case "WEEKLY": {
			const days = (parts.get("BYDAY") ?? "")
				.split(",")
				.map((day) => WEEKDAYS[day.trim()])
				.filter((day): day is string => Boolean(day));
			if (days.length === 0) {
				return interval === 1
					? "varje vecka"
					: `var ${ordinal(interval)} vecka`;
			}
			if (interval === 1) {
				return `varje ${joinWords(days)}`;
			}
			return `var ${ordinal(interval)} vecka (${joinWords(days)})`;
		}
		case "MONTHLY":
			return interval === 1
				? "varje månad"
				: `var ${ordinal(interval)} månad`;
		case "YEARLY":
			return interval === 1 ? "varje år" : `vart ${ordinal(interval)} år`;
		default:
			return rule;
	}
}

function ordinal(value: number): string {
	return ORDINALS[value] ?? `${value}:e`;
}

function joinWords(words: string[]): string {
	if (words.length === 1) return words[0] ?? "";

	const last = words[words.length - 1] ?? "";
	return `${words.slice(0, -1).join(", ")} och ${last}`;
}

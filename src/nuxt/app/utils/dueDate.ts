// Due-datum hanteras som kalenderdatum ("YYYY-MM-DD") utan tidszon.
// new Date("2026-11-20") tolkas som UTC-midnatt, vilket kan visa fel dag
// i tidszoner bakom UTC - därför byggs datumet lokalt från delarna.

export type DueTone = "overdue" | "today" | "upcoming";

export function parseDueDate(due: string | null | undefined): Date | null {
	if (!due) return null;

	const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(due);
	if (!match) return null;

	const date = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]));
	return Number.isNaN(date.getTime()) ? null : date;
}

// Kort form av datumet, t.ex. "20 nov"
export function formatDueDate(due: string | null | undefined): string {
	const date = parseDueDate(due);
	if (!date) return "";

	return date.toLocaleDateString("sv-SE", { day: "numeric", month: "short" });
}

// Klassificerar datumet: förgått, idag eller kommande.
export function dueTone(due: string | null | undefined): DueTone {
	const date = parseDueDate(due);
	if (!date) return "upcoming";

	const today = new Date();
	today.setHours(0, 0, 0, 0);

	const days = Math.round((date.getTime() - today.getTime()) / 86400000);
	if (days < 0) return "overdue";
	if (days === 0) return "today";

	return "upcoming";
}

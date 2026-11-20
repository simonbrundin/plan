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

// Dagens datum som kalenderdatum ("YYYY-MM-DD") i lokal tid.
export function todayDateString(): string {
	const today = new Date();
	const month = String(today.getMonth() + 1).padStart(2, "0");
	const day = String(today.getDate()).padStart(2, "0");
	return `${today.getFullYear()}-${month}-${day}`;
}

// Sant om datumet ligger efter idag. Ett startdatum i framtiden betyder
// att målet inte får dyka upp i prioriteringslistan ännu - dagens datum
// räknas alltså inte som framtid, då målet syns från och med idag.
export function isFutureDate(date: string | null | undefined): boolean {
	const parsed = parseDueDate(date);
	if (!parsed) return false;

	const today = new Date();
	today.setHours(0, 0, 0, 0);

	return parsed.getTime() > today.getTime();
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

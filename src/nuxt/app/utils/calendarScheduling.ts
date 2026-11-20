/**
 * Schemaläggning av mål i kalendern.
 *
 * Två saker hör hemma här: att bära ett mål från en lista till kalendern,
 * och att flytta en händelse i tidslinjen. Båda bygger på samma sak - att
 * översätta en pekarposition till en tid, och att snappa tiden till en
 * rimlig upplösning.
 *
 * Logiken ligger här, utan Vue, så att den går att testa.
 */

import { formatEstimate } from "./estimate";

/** MIME-typen som bär ett mål från en lista till kalendern. */
export const GOAL_DRAG_MIME = "application/x-plan-goal";

/** Vad som följer med när ett mål dras. */
export interface GoalDragPayload {
	id: number;
	title: string;
	icon?: string;
	/** Målets uppskattade tidsåtgång i minuter, om den är satt. */
	estimated_minutes?: number | null;
}

/**
 * Gör om ett mål till den text som följer med under dragningen.
 *
 * Texten innehåller bara det som behövs för att skapa en händelse, så att
 * inget internt läcker ut i klippbordet om någon kopierar den.
 */
export function serializeGoalDrag(goal: GoalDragPayload): string {
	return JSON.stringify({
		id: goal.id,
		title: goal.title,
		icon: goal.icon,
		estimated_minutes: goal.estimated_minutes ?? null,
	});
}

/**
 * Startar en dragning av ett mål.
 *
 * Målet läggs i klippbordet under dragningen, så att kalendern kan läsa
 * det när målet släpps. Texten innehåller bara målets id, titel, ikon och
 * uppskattning - inget mer.
 */
export function startGoalDrag(
	event: DragEvent,
	goal: GoalDragPayload,
): void {
	const transfer = event.dataTransfer;
	if (!transfer) return;

	transfer.effectAllowed = "copy";
	transfer.setData(GOAL_DRAG_MIME, serializeGoalDrag(goal));
	// Vissa webbläsare vägrar släppa om inte den vanliga texten också
	// sätts, så den läggs med trots att kalendern inte läser den.
	transfer.setData("text/plain", goal.title);
}

/**
 * Sant om en dragning bär ett mål.
 *
 * Kalendern ska bara reagera på mål som dras till den - inte på något annat
 * som dras i sidan.
 */
export function carriesGoal(event: DragEvent): boolean {
	const types = event.dataTransfer?.types;
	if (!types) return false;

	return Array.from(types).includes(GOAL_DRAG_MIME);
}

/**
 * Läser tillbaka ett mål från en dragad text.
 *
 * Returnerar null om texten inte är ett mål - då kan den komma från något
 * annat som dragas i sidan, och då ska kalendern inte reagera.
 */
export function parseGoalDrag(data: string | null | undefined): GoalDragPayload | null {
	if (!data) return null;

	let parsed: unknown;
	try {
		parsed = JSON.parse(data);
	} catch {
		return null;
	}

	if (typeof parsed !== "object" || parsed === null) return null;

	const candidate = parsed as Record<string, unknown>;
	if (typeof candidate.id !== "number" || !Number.isFinite(candidate.id)) {
		return null;
	}
	if (typeof candidate.title !== "string" || candidate.title.length === 0) {
		return null;
	}

	const estimate = candidate.estimated_minutes;

	return {
		id: candidate.id,
		title: candidate.title,
		icon: typeof candidate.icon === "string" ? candidate.icon : undefined,
		estimated_minutes:
			typeof estimate === "number" && Number.isFinite(estimate)
				? estimate
				: null,
	};
}

/** Ett tidsalternativ i listan som visas när ett mål släpps. */
export interface DurationOption {
	/** Längden i minuter. */
	minutes: number;
	/** Etiketten som visas, t.ex. "1 h 30 min". */
	label: string;
	/**
	 * Sant om alternativet kommer från målets egen uppskattning.
	 *
	 * Det alternativet står först i listan, och markeras så att det går att
	 * se varför det finns med.
	 */
	suggested: boolean;
}

/** De vanligaste längderna, i minuter. */
const STANDARD_DURATIONS = [15, 30, 60, 120, 240];

/**
 * Listan med tidsalternativ som visas när ett mål släpps i kalendern.
 *
 * Målets egen uppskattning ligger först, så att det vanliga fallet - att
 * målet redan har en uppskattad tidsåtgång - bara kräver ett klick.
 * Därefter följer de vanligaste längderna.
 */
export function durationOptions(
	estimatedMinutes?: number | null,
): DurationOption[] {
	const options: DurationOption[] = [];
	const seen = new Set<number>();

	const add = (minutes: number, suggested: boolean) => {
		if (!Number.isFinite(minutes) || minutes <= 0) return;
		const rounded = Math.round(minutes);
		if (seen.has(rounded)) return;

		seen.add(rounded);
		options.push({
			minutes: rounded,
			label: formatEstimate(rounded) || `${rounded} min`,
			suggested,
		});
	};

	add(estimatedMinutes ?? null, true);
	for (const minutes of STANDARD_DURATIONS) add(minutes, false);

	return options;
}

/**
 * Hur mycket tid en pixel i tidslinjen motsvarar, i millisekunder.
 *
 * Tidslinjen visar hela fönstret över vyns höjd, så en pixel är lika stor
 * del av tiden oavsett hur utzoomat vyn är.
 */
export function millisecondsPerPixel(
	visibleDuration: number,
	height: number,
): number {
	if (height <= 0) return 0;
	return visibleDuration / height;
}

/**
 * Räknar ut vilken tid en pekarposition i vyn motsvarar.
 *
 * Positionen räknas i pixlar uppifrån vyn. Resultatet kan ligga utanför
 * tidsfönstret - då klipps det in, så att en händelse inte hamnar långt
 * före eller efter det som syns.
 */
export function timeAtOffset(options: {
	/** Pekarens avstånd från vyns överkant, i pixlar. */
	offsetY: number;
	/** Vyns höjd, i pixlar. */
	height: number;
	/** Tidsfönstrets början, i millisekunder. */
	startTime: number;
	/** Tidsfönstrets längd, i millisekunder. */
	visibleDuration: number;
}): number {
	const { offsetY, height, startTime, visibleDuration } = options;

	if (height <= 0) return startTime;

	const raw = startTime + (offsetY / height) * visibleDuration;
	return Math.min(Math.max(raw, startTime), startTime + visibleDuration);
}

/**
 * Hur stor upplösning tiden snappas till, givet hur utzoomat vyn är.
 *
 * I en timme måste tiden fälla på minuter, i en månad på timmar, och i ett
 * år på dagar. Annars skulle en etikett som visar "14:20" inte betyda
 * något när vyn visar halva året.
 */
export function snapStepMs(visibleDuration: number): number {
	const minute = 60 * 1000;
	const hour = 60 * minute;
	const day = 24 * hour;

	if (visibleDuration <= 2 * day) return 15 * minute;
	if (visibleDuration <= 10 * day) return 30 * minute;
	if (visibleDuration <= 90 * day) return hour;
	return day;
}

/**
 * Snappar en tid till närmaste multipel av steget.
 */
export function snapTimestamp(timestamp: number, stepMs: number): number {
	if (stepMs <= 0) return timestamp;
	return Math.round(timestamp / stepMs) * stepMs;
}

/**
 * Flyttar en händelse och behåller dess längd.
 *
 * Tiden flyttas så att händelsens början hamnar på den nya tiden, och
 * sluttiden följer med samma avstånd.
 */
export function shiftEventTimes(options: {
	/** Händelsens början, i millisekunder. */
	start: number;
	/** Händelsens slut, i millisekunder. */
	end: number;
	/** Hur mycket händelsen ska flyttas, i millisekunder. */
	deltaMs: number;
}): { start: number; end: number } {
	const { start, end, deltaMs } = options;

	return {
		start: start + deltaMs,
		end: end + deltaMs,
	};
}

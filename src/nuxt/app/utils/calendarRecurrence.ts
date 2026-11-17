/**
 * Lokal expansion av återkommande kalenderhändelser.
 *
 * Kalendern hämtar alla händelser en gång när den öppnas. Serien kommer då
 * som rader med en `recurrence_rule`, och det är här som serien expanderas
 * till enskilda förekomster för det tidsfönster som just nu syns.
 *
 * Syftet är att scrolla och zooma ska vara snabbt och inte kräva nya
 * anrop till servern: expanderingen sker lokalt, samma sätt som rrule-go
 * gör det på servern (båda biblioteken är portar av samma datumalgoritm,
 * så en serie som expanderas här får samma förekomster som servern skulle
 * ha gett).
 */

import { RRule } from "rrule";

import type { CalendarEvent } from "../composables/useCalendarApi";

/** Ett dygn i millisekunder. */
const DAY = 24 * 60 * 60 * 1000;

/**
 * Max antal förekomster som expanderas för en och samma serie.
 *
 * En serie som upprepas varje dag i ett århundrade ger fler förekomster än
 * det finns användning för. Capen skyddar både minnet och renderingen.
 */
export const MAX_OCCURRENCES_PER_SERIES = 5000;

/**
 * Max tidsfönster som expanderas för en serie.
 *
 * Kalendern kan zoomas ut till ett nästan obegränsat tidsfönster. Utan en
 * gräns skulle en daglig serie generera miljontals förekomster.
 */
export const MAX_EXPANSION_SPAN = 50 * 365 * DAY;

interface CachedRule {
	source: string;
	rule: RRule;
}

// Regelinstanserna är dyra att skapa och används för varje bildruta när
// tidsfönstret flyttas, så de behålls mellan anropen. Nyckeln är seriens
// id, och källsträngen kontrolleras så att en ändrad serie inte återanvänder
// en gammal regel.
const ruleCache = new Map<number, CachedRule>();

const parseRule = (series: CalendarEvent): RRule | null => {
	const source = series.recurrence_rule;

	if (!source) return null;

	const cached = ruleCache.get(series.id);
	if (cached && cached.source === source) return cached.rule;

	try {
		const parsed = RRule.parseString(source);
		const rule = new RRule({ ...parsed, dtstart: new Date(series.start_at) });
		ruleCache.set(series.id, { source, rule });
		return rule;
	} catch {
		// En trasig regel ska inte krascha kalendern - serien visas då
		// bara som en enstaka händelse.
		return null;
	}
};

const overlaps = (event: CalendarEvent, from: number, to: number): boolean => {
	const start = Date.parse(event.start_at);
	const end = Date.parse(event.end_at);

	return start < to && end > from;
};

/**
 * Fördelar förekomster jämnt över tidsfönstret när de blir för många.
 *
 * Att behålla de första förekomsterna skulle lämna resten av fönstret tomt.
 * Änträdet är i stället att hoppa över var n:te förekomst, så att serien
 * finns representerad i hela fönstret.
 */
const spread = (occurrences: Date[]): Date[] => {
	const step = Math.ceil(occurrences.length / MAX_OCCURRENCES_PER_SERIES);
	const result: Date[] = [];

	for (let index = 0; index < occurrences.length; index += step) {
		result.push(occurrences[index]);
	}

	return result;
};

const clampWindow = (from: number, to: number): { from: number; to: number } => {
	const span = to - from;

	if (span <= MAX_EXPANSION_SPAN) return { from, to };

	const centre = (from + to) / 2;
	return {
		from: centre - MAX_EXPANSION_SPAN / 2,
		to: centre + MAX_EXPANSION_SPAN / 2,
	};
};

/**
 * Expanderar en serie till de förekomster som ligger i tidsfönstret.
 *
 * Serien börjar på `start_at` och varje förekomst är lika lång som
 * händelsen. En förekomst som börjar före fönstrets början men slutar
 * inne i fönstret räknas med, precis som servern gör det.
 */
export function expandSeries(
	series: CalendarEvent,
	from: number,
	to: number,
): CalendarEvent[] {
	const rule = parseRule(series);

	if (!rule) return [];

	const start = Date.parse(series.start_at);
	const end = Date.parse(series.end_at);
	const duration = Math.max(0, end - start);
	const window = clampWindow(from, to);

	const raw = rule.between(
		new Date(window.from - duration),
		new Date(window.to),
		true,
	);
	const occurrences = raw.length > MAX_OCCURRENCES_PER_SERIES ? spread(raw) : raw;

	const result: CalendarEvent[] = [];

	for (const occurrence of occurrences) {
		const occurrenceStart = occurrence.getTime();
		const occurrenceEnd = occurrenceStart + duration;

		if (occurrenceStart >= window.to || occurrenceEnd <= window.from) continue;

		result.push({
			...series,
			occurrence_id: `${series.id}-${occurrenceStart}`,
			original_start_at: new Date(occurrenceStart).toISOString(),
			start_at: new Date(occurrenceStart).toISOString(),
			end_at: new Date(occurrenceEnd).toISOString(),
		});
	}

	return result;
}

const overrideKey = (parentId: number, originalStartAt: string): string =>
	`${parentId}|${Date.parse(originalStartAt)}`;

/**
 * Expanderar hela datasetet till de händelser som syns i tidsfönstret.
 *
 * Datasetet innehåller både serier (rader med `recurrence_rule`) och
 * override-rader (rader med `parent_event_id`). En override ersätter den
 * förekomst den hör till, och en avstängd override tar bort förekomsten.
 */
export function expandCalendarEvents(
	events: CalendarEvent[],
	from: number,
	to: number,
): CalendarEvent[] {
	const masters = events.filter((event) => !event.parent_event_id);
	const overrides = events.filter((event) => event.parent_event_id);

	const overrideByKey = new Map<string, CalendarEvent>();
	for (const override of overrides) {
		if (!override.parent_event_id || !override.original_start_at) continue;
		overrideByKey.set(
			overrideKey(override.parent_event_id, override.original_start_at),
			override,
		);
	}

	const unmatched = new Set(overrides);
	const result: CalendarEvent[] = [];

	for (const master of masters) {
		if (!master.recurrence_rule) {
			if (overlaps(master, from, to)) result.push(master);
			continue;
		}

		for (const occurrence of expandSeries(master, from, to)) {
			const key = overrideKey(master.id, occurrence.start_at);
			const override = overrideByKey.get(key);

			if (!override) {
				result.push(occurrence);
				continue;
			}

			unmatched.delete(override);

			// En avstängd override tar bort förekomsten helt.
			if (override.status === "cancelled") continue;

			result.push({
				...override,
				occurrence_id: occurrence.occurrence_id,
			});
		}
	}

	// Overrides som inte motsvarar någon förekomst i fönstret - till exempel
	// en förekomst som har flyttats hit från ett annat ställe - visas ändå,
	// så länge de ligger i fönstret.
	for (const override of unmatched) {
		if (override.status === "cancelled") continue;
		if (!overlaps(override, from, to)) continue;

		result.push({
			...override,
			occurrence_id: `${override.id}-${Date.parse(override.start_at)}`,
		});
	}

	return result.sort(
		(left, right) => Date.parse(left.start_at) - Date.parse(right.start_at),
	);
}

/**
 * Expanderar datasetet till synliga händelser, med cachning.
 *
 * Tidsfönstret flyttas under scroll och zoom, och varje bildruta kan utlösa
 * en ny beräkning. Fönstret avrundas därför ned till hela dygn innan
 * expansionen görs, så att samma cache-nyckel återanvänds så länge
 * visningen ligger inom samma dygn.
 */
export class CalendarExpander {
	private events: CalendarEvent[] = [];
	private revision = 0;
	private readonly cache = new Map<string, CalendarEvent[]>();

	/** Sätter in de händelser som ska expanderas och tömmer cachen. */
	setEvents(events: CalendarEvent[]): void {
		this.events = events;
		this.revision += 1;
		this.cache.clear();
	}

	/** De händelser som ligger i tidsfönstret. */
	visible(from: number, to: number): CalendarEvent[] {
		const bucketFrom = Math.floor(from / DAY) * DAY;
		const bucketTo = Math.ceil(to / DAY) * DAY;
		const key = `${this.revision}:${bucketFrom}:${bucketTo}`;

		let cached = this.cache.get(key);

		if (!cached) {
			cached = expandCalendarEvents(this.events, bucketFrom, bucketTo);
			this.cache.set(key, cached);

			// Cacheen ska inte växa obegränsat när användaren rör sig i tiden.
			if (this.cache.size > 8) {
				const oldest = this.cache.keys().next().value;
				if (oldest !== undefined) this.cache.delete(oldest);
			}
		}

		return cached.filter((event) => overlaps(event, from, to));
	}
}

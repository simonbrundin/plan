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

/** Max antal tidsfönster som cachas samtidigt. */
const MAX_CACHED_WINDOWS = 8;

/** Ett tidsfönster i millisekunder. */
export interface TimeWindow {
	from: number;
	to: number;
}

/**
 * Regeln för en serie, eller null om regeln inte går att tolka.
 *
 * Att tolka en regel kostar någon tiondels millisekund, så det görs varje
 * gång serien expanderas. Det håller funktionerna rena från tillstånd.
 */
const ruleFor = (series: CalendarEvent): RRule | null => {
	const source = series.recurrence_rule;

	if (!source) return null;

	try {
		const parsed = RRule.parseString(source);
		return new RRule({ ...parsed, dtstart: new Date(series.start_at) });
	} catch {
		// En trasig regel ska inte krascha kalendern - serien visas då
		// bara som en enstaka händelse.
		return null;
	}
};

/** Sant om serien har en regel som går att tolka. */
export function canExpandSeries(series: CalendarEvent): boolean {
	return ruleFor(series) !== null;
}

/** Sant om händelsen pågår inom tidsfönstret. */
const overlaps = (event: CalendarEvent, from: number, to: number): boolean => {
	const start = Date.parse(event.start_at);
	const end = Date.parse(event.end_at);

	return start < to && end > from;
};

/**
 * Perioden en samplad förekomst står för.
 *
 * Representanten är den första förekomsten i perioden, så perioden går att
 * härleda från dess starttid. Det görs i lokal tid, precis som `thinOut`
 * grupperar förekomsterna.
 */
const sampleRange = (
	start: number,
	period: SamplePeriod,
): { from: number; to: number } => {
	const date = new Date(start);

	if (period === "day") {
		return {
			from: new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime(),
			to: new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1).getTime(),
		};
	}

	return {
		from: new Date(date.getFullYear(), date.getMonth(), 1).getTime(),
		to: new Date(date.getFullYear(), date.getMonth() + 1, 1).getTime(),
	};
};

/**
 * Sant om händelsen syns i tidsfönstret.
 *
 * En samplad förekomst står för en hel period. Även om representanten
 * ligger utanför fönstret ska den synas, så länge perioden overlappar
 * fönstret - annars skulle periodens antal försvinna ur vyn, och det är
 * precis det serierna inte ska göra.
 */
const isVisible = (event: CalendarEvent, from: number, to: number): boolean => {
	const weight = event.occurrence_weight ?? 1;
	const period = event.occurrence_period;

	if (weight <= 1 || !period) return overlaps(event, from, to);

	const range = sampleRange(Date.parse(event.start_at), period);

	return range.from < to && range.to > from;
};

/** Händelsens längd i millisekunder. */
const durationOf = (event: CalendarEvent): number =>
	Math.max(0, Date.parse(event.end_at) - Date.parse(event.start_at));

/**
 * Fördelar förekomsterna över de perioder de ligger i.
 *
 * Att behålla de första förekomsterna skulle lämna resten av fönstret tomt.
 * Änträdet är i stället att behålla en förekomst per period, med en vikt som
 * motsvarar hur många förekomster perioden faktiskt hade. Då finns serien
 * representerad i hela fönstret, och antalen per dag, månad och år blir
 * ändå exakta.
 *
 * Perioden är dagen så länge antalet dagar ryms under capen. Då blir även
 * antalet per dag exakt. En serie som sträcker sig över så många dagar att
 * de inte ryms fördelas per månad i stället - på den nivån är vyn så
 * överblickbar att enstaka dagar inte längre syns, och antalet per år är
 * fortfarande exakt.
 *
 * Förekomster som har en override måste behållas, annars skulle override
 * bli överflödiga och deras förekomst räknas in två gånger. De räknas inte
 * in i periodens vikt, eftersom de visas för sig.
 */
const thinOut = (
	starts: number[],
	mustKeep?: (start: number) => boolean,
): Array<{ start: number; weight: number; period: SamplePeriod }> => {
	const kept: number[] = [];
	const rest: number[] = [];

	for (const start of starts) {
		if (mustKeep?.(start)) {
			kept.push(start);
		} else {
			rest.push(start);
		}
	}

	const byDay = new Map<string, number[]>();
	const byMonth = new Map<string, number[]>();

	for (const start of rest) {
		const date = new Date(start);
		const dayKey = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
		const monthKey = `${date.getFullYear()}-${date.getMonth()}`;

		const day = byDay.get(dayKey);

		if (day) {
			day.push(start);
		} else {
			byDay.set(dayKey, [start]);
		}

		const month = byMonth.get(monthKey);

		if (month) {
			month.push(start);
		} else {
			byMonth.set(monthKey, [start]);
		}
	}

	const buckets = byDay.size <= MAX_OCCURRENCES_PER_SERIES ? byDay : byMonth;
	const period: SamplePeriod = buckets === byDay ? "day" : "month";

	const result: Array<{ start: number; weight: number; period: SamplePeriod }> = [];

	for (const bucket of buckets.values()) {
		const [first] = bucket;

		if (first === undefined) continue;

		result.push({ start: first, weight: bucket.length, period });
	}

	// Förekomster med override behålls var för sig. De är inte samplade,
	// så de har ingen period - de räknas för sig i stället för i vikten.
	for (const start of kept) {
		result.push({ start, weight: 1, period: "day" });
	}

	return result.sort((left, right) => left.start - right.start);
};

/** Tidsfönstret begränsat till högst MAX_EXPANSION_SPAN. */
const clampWindow = (from: number, to: number): TimeWindow => {
	const span = to - from;

	if (span <= MAX_EXPANSION_SPAN) return { from, to };

	const centre = (from + to) / 2;
	return {
		from: centre - MAX_EXPANSION_SPAN / 2,
		to: centre + MAX_EXPANSION_SPAN / 2,
	};
};

/**
 * Sant om förekomsten räknas in i fönstret.
 *
 * En förekomst som börjar före fönstrets början men slutar inne i fönstret
 * räknas med, precis som servern gör det.
 */
const isInsideWindow = (
	start: number,
	duration: number,
	window: TimeWindow,
): boolean => start < window.to && start + duration > window.from;

/**
 * Den period en uttunnad serie har tunnats ut i.
 *
 * Varje visad förekomst står för alla förekomster i sin period, så det
 * behövs veta vilken period det är för att kunna beskriva urvalet rätt.
 */
export type SamplePeriod = "day" | "month";

/** Etikett för en samplingsperiod. */
const PERIOD_LABELS: Record<SamplePeriod, string> = {
	day: "dag",
	month: "månad",
};

/** En förekomst som en händelse. */
const occurrenceOf = (
	series: CalendarEvent,
	start: number,
	duration: number,
	weight = 1,
	total?: number,
	period?: SamplePeriod,
): CalendarEvent => ({
	...series,
	occurrence_id: `${series.id}-${start}`,
	original_start_at: new Date(start).toISOString(),
	start_at: new Date(start).toISOString(),
	end_at: new Date(start + duration).toISOString(),
	occurrence_weight: weight,
	...(total === undefined ? {} : { occurrence_total: total }),
	...(period === undefined ? {} : { occurrence_period: period }),
});

/**
 * Seriens förekomster i tidsfönstret, som starttid.
 *
 * Returnerar `null` om regeln inte går att tolka - då ska serien visas som
 * en enstaka händelse i stället för att försvinna ur kalendern.
 *
 * När serien har fler förekomster än capen tunnas den ut: en förekomst per
 * period behålls, med en vikt som gör att antalen blir rätta ändå. Det
 * sanna antalet förekomster i fönstret finns då i `occurrence_total`.
 */
const occurrenceStarts = (
	series: CalendarEvent,
	from: number,
	to: number,
	mustKeep?: (start: number) => boolean,
): Array<{ start: number; weight: number; total?: number; period?: SamplePeriod }> | null => {
	const rule = ruleFor(series);

	if (!rule) return null;

	const duration = durationOf(series);
	const window = clampWindow(from, to);
	const occurrences = rule.between(
		new Date(window.from - duration),
		new Date(window.to),
		true,
	);
	const inside = occurrences
		.map((occurrence) => occurrence.getTime())
		.filter((start) => isInsideWindow(start, duration, window));

	if (inside.length <= MAX_OCCURRENCES_PER_SERIES) {
		return inside.map((start) => ({ start, weight: 1 }));
	}

	return thinOut(inside, mustKeep).map((entry) => ({
		...entry,
		total: inside.length,
	}));
};

/**
 * Expanderar en serie till de förekomster som ligger i tidsfönstret.
 *
 * Serien börjar på `start_at` och varje förekomst är lika lång som
 * händelsen.
 */
export function expandSeries(
	series: CalendarEvent,
	from: number,
	to: number,
): CalendarEvent[] {
	const starts = occurrenceStarts(series, from, to);

	if (!starts) return [];

	const duration = durationOf(series);

	return starts.map(({ start, weight, total, period }) =>
		occurrenceOf(series, start, duration, weight, total, period),
	);
}

/**
 * Nyckel för en förekomst: vilken serie den hör till och vilken start den
 * hade innan den eventuellt flyttades.
 */
const occurrenceKey = (parentId: number, originalStartAt: string): string =>
	`${parentId}|${Date.parse(originalStartAt)}`;

/** Nyckel för en förekomst, given starttid i millisekunder. */
const occurrenceKeyAt = (parentId: number, startAt: number): string =>
	`${parentId}|${startAt}`;

/**
 * Register över override-rader.
 *
 * En override hör till en förekomst i en serie, och identifieras av seriens
 * id och förekomstens ursprungliga start.
 */
class OverrideIndex {
	private readonly byOccurrence = new Map<string, CalendarEvent>();
	private readonly used = new Set<CalendarEvent>();

	constructor(private readonly overrides: CalendarEvent[]) {
		for (const override of overrides) {
			if (!override.parent_event_id || !override.original_start_at) continue;

			this.byOccurrence.set(
				occurrenceKey(override.parent_event_id, override.original_start_at),
				override,
			);
		}
	}

	/** Override-raden som hör till förekomsten, om det finns någon. */
	find(parentId: number, originalStartAt: string): CalendarEvent | undefined {
		return this.byOccurrence.get(occurrenceKey(parentId, originalStartAt));
	}

	/** Sant om det finns en override för förekomsten med given start. */
	has(parentId: number, startAt: number): boolean {
		return this.byOccurrence.has(occurrenceKeyAt(parentId, startAt));
	}

	/** Markerar att override-raden har använts mot en förekomst. */
	markUsed(override: CalendarEvent): void {
		this.used.add(override);
	}

	/** De override-rader som inte har motsvarat någon förekomst. */
	unused(): CalendarEvent[] {
		return this.overrides.filter((override) => !this.used.has(override));
	}
}

/** Delar upp datasetet i serier och override-rader. */
const splitEvents = (
	events: CalendarEvent[],
): { series: CalendarEvent[]; overrides: CalendarEvent[] } => ({
	series: events.filter((event) => !event.parent_event_id),
	overrides: events.filter((event) => event.parent_event_id),
});

/** Händelsen om den ligger i fönstret, annars ingenting. */
const visibleInside = (
	event: CalendarEvent,
	from: number,
	to: number,
): CalendarEvent[] => (overlaps(event, from, to) ? [event] : []);

/**
 * Ersätter en förekomst med sin override, om det finns någon.
 *
 * En avstängd override tar bort förekomsten helt.
 */
const applyOverride = (
	occurrence: CalendarEvent,
	masterId: number,
	overrides: OverrideIndex,
): CalendarEvent[] => {
	const override = overrides.find(masterId, occurrence.start_at);

	if (!override) return [occurrence];

	overrides.markUsed(override);

	if (override.status === "cancelled") return [];

	// Vikten förs vidare från förekomsten. En override ersätter en
	// förekomst, men den förekomsten kan representera flera.
	return [
		{
			...override,
			occurrence_id: occurrence.occurrence_id,
			occurrence_weight: occurrence.occurrence_weight,
			occurrence_total: occurrence.occurrence_total,
			occurrence_period: occurrence.occurrence_period,
		},
	];
};

/**
 * Expanderar en masterhändelse till de händelser som syns i fönstret.
 *
 * En serie expanderas till sina förekomster, medan en enstaka händelse bara
 * visas om den ligger i fönstret. En override ersätter den förekomst den
 * hör till, och en avstängd override tar bort förekomsten.
 */
const expandMaster = (
	master: CalendarEvent,
	from: number,
	to: number,
	overrides: OverrideIndex,
): CalendarEvent[] => {
	// Förekomster med override behålls var för sig när serien tunnas ut,
	// annars skulle override sakna sin förekomst och visas dubbelt.
	const starts = occurrenceStarts(master, from, to, (start) =>
		overrides.has(master.id, start),
	);

	// En serie utan regel, eller med en regel som inte går att tolka, visas
	// som en enstaka händelse - den får inte försvinna ur kalendern.
	if (!starts) return visibleInside(master, from, to);

	const duration = durationOf(master);

	return starts
		.map(({ start, weight, total, period }) =>
			occurrenceOf(master, start, duration, weight, total, period),
		)
		.flatMap((occurrence) => applyOverride(occurrence, master.id, overrides));
};

/**
 * Override-rader som inte motsvarar någon förekomst i fönstret.
 *
 * Det kan till exempel vara en förekomst som har flyttats hit från ett
 * annat ställe. Den visas ändå, så länge den ligger i fönstret.
 */
const unmatchedOverrides = (
	overrides: OverrideIndex,
	from: number,
	to: number,
): CalendarEvent[] =>
	overrides
		.unused()
		.filter((override) => override.status !== "cancelled")
		.filter((override) => overlaps(override, from, to))
		.map((override) => ({
			...override,
			occurrence_id: `${override.id}-${Date.parse(override.start_at)}`,
		}));

/**
 * Expanderar hela datasetet till de händelser som syns i tidsfönstret.
 *
 * Datasetet innehåller både serier (rader med `recurrence_rule`) och
 * override-rader (rader med `parent_event_id`).
 */
export function expandCalendarEvents(
	events: CalendarEvent[],
	from: number,
	to: number,
): CalendarEvent[] {
	const { series, overrides } = splitEvents(events);
	const index = new OverrideIndex(overrides);
	const result: CalendarEvent[] = [];

	for (const master of series) {
		result.push(...expandMaster(master, from, to, index));
	}

	result.push(...unmatchedOverrides(index, from, to));

	return result.sort(
		(left, right) => Date.parse(left.start_at) - Date.parse(right.start_at),
	);
}

/** Tidsfönstret avrundat till hela dygn. */
const dayWindow = (from: number, to: number): TimeWindow => ({
	from: Math.floor(from / DAY) * DAY,
	to: Math.ceil(to / DAY) * DAY,
});

/**
 * Cache av expanderade tidsfönster.
 *
 * Tidsfönstret flyttas under scroll och zoom, och varje bildruta kan utlösa
 * en ny beräkning. Fönstret avrundas därför ned till hela dygn innan
 * expansionen görs, så att samma cache-nyckel återanvänds så länge
 * visningen ligger inom samma dygn.
 *
 * Cachen hänger på händelsearrayen: en ny hämtning ger en ny array och då
 * börjar cachen om från början.
 */
class WindowCache {
	private readonly entries = new WeakMap<
		CalendarEvent[],
		Map<string, CalendarEvent[]>
	>();

	/** De expanderade händelserna för fönstret, beräknade vid behov. */
	get(
		events: CalendarEvent[],
		window: TimeWindow,
		expand: () => CalendarEvent[],
	): CalendarEvent[] {
		const key = `${window.from}:${window.to}`;

		let cache = this.entries.get(events);

		if (!cache) {
			cache = new Map();
			this.entries.set(events, cache);
		}

		const cached = cache.get(key);

		if (cached) return cached;

		const expanded = expand();
		cache.set(key, expanded);

		// Cacheen ska inte växa obegränsat när användaren rör sig i tiden.
		if (cache.size > MAX_CACHED_WINDOWS) {
			const oldest = cache.keys().next().value;
			if (oldest !== undefined) cache.delete(oldest);
		}

		return expanded;
	}
}

/**
 * Expanderar datasetet till synliga händelser, med cachning.
 *
 * Händelserna skickas in i varje anrop i stället för att hållas inne i
 * expanderaren. Då beror resultatet på de händelser som faktiskt är laddade,
 * och vyn räknas om så fort en ny hämtning kommer in.
 */
export class CalendarExpander {
	private readonly cache = new WindowCache();

	/** De händelser som ligger i tidsfönstret. */
	visible(
		events: CalendarEvent[],
		from: number,
		to: number,
	): CalendarEvent[] {
		const window = dayWindow(from, to);
		const expanded = this.cache.get(events, window, () =>
			expandCalendarEvents(events, window.from, window.to),
		);

		return expanded.filter((event) => isVisible(event, from, to));
	}
}

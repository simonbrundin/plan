/**
 * Vilka återkommande serier som ska döljas i tidslinjen.
 *
 * Tidslinjen visar alla händelser som rader i en kolumn. Ibland ryms de inte:
 * en daglig rutin som syns trettio gånger i en månadsvy tar över kolumnen, och
 * då blir de enskilda händelserna omöjliga att läsa. Då döljs serier, först
 * de dagliga, sedan de som har flest förekomster i vyn.
 *
 * Döljningen är aldrig tyst. En dold serie listas i sammanfattningen ovanför
 * tidslinjen, märkt som "dold", och kan visas ensam genom ett klick.
 */

import { layoutEventRows, type EventRowInput } from "./calendarLayout.ts";

/** En återkommande serie i vyn. */
export interface VisibilitySeries {
  /** Seriens nyckel. */
  key: string;
  /** Seriens titel. */
  title: string;
  /** Sant om serien återkommer varje dag. */
  daily: boolean;
  /** Antalet förekomster i vyn. */
  count: number;
}

/** En händelse i vyn, med serietillhörighet. */
export interface VisibilityEvent extends EventRowInput {
  /** Seriens nyckel, eller null för en enstak händelse. */
  seriesKey: string | null;
}

/** Underlaget för en plan. */
export interface VisibilityInput {
  /** Alla händelser i vyn. */
  events: VisibilityEvent[];
  /** Serierna i vyn. */
  series: VisibilitySeries[];
  /** Hur många dagliga förekomster som får synas innan serien döljs. */
  dailyLimit: number;
  /** Om serier får döljas alls. */
  hidingEnabled: boolean;
  /** Serien användaren har valt att visa ensam, om någon. */
  focusKey: string | null;
  /** Serierna som var dolda när planen senast gjordes. */
  hiddenBefore: readonly string[];
  /** Tidsfönstrets början, i millisekunder. */
  from: number;
  /** Tidsfönstrets slut, i millisekunder. */
  to: number;
  /** Vyns höjd, i pixlar. */
  height: number;
}

/** Resultatet av en plan. */
export interface VisibilityPlan {
  /** Serierna som ska döljas. */
  hidden: string[];
  /**
   * Sant om allt inte ryms i vyn, även efter döljning.
   *
   * Då är det enstaka händelser som inte hinner med, och de behöver visas
   * som ett antal så att de inte försvinner tyst.
   */
  overflowing: boolean;
}

/**
 * Planerar vilka serier som ska döljas.
 *
 * Planen är ren: samma underlag ger samma svar. Hysteresisen ligger i
 * `hiddenBefore`, så att en serie inte hoppar fram och tillbaka när antalet
 * förekomster passerar gränsen.
 */
export function planVisibility(input: VisibilityInput): VisibilityPlan {
  const hidden = new Set<string>();

  const fits = (hiddenKeys: ReadonlySet<string>): boolean => {
    const items = input.events
      .filter(
        (event) =>
          event.seriesKey === null || !hiddenKeys.has(event.seriesKey),
      )
      .map((event) => ({ start: event.start, end: event.end }));

    if (items.length === 0) return true;

    return layoutEventRows(items, {
      from: input.from,
      to: input.to,
      height: input.height,
    }).every((row) => !row.clipped);
  };

  // Fokus vinner över allt annat: användaren har bett om att få se just
  // den här serien.
  if (input.focusKey) {
    for (const series of input.series) {
      if (series.key !== input.focusKey) hidden.add(series.key);
    }

    return { hidden: [...hidden], overflowing: !fits(hidden) };
  }

  if (!input.hidingEnabled) {
    return { hidden: [], overflowing: !fits(hidden) };
  }

  const limit = Math.max(1, Math.round(input.dailyLimit));
  const wasHidden = new Set(input.hiddenBefore);

  // Regel 1: en daglig serie med fler förekomster än gränsen döljs. Den
  // visas igen först när antalet faller under gränsen, så att serien inte
  // fladdrar när man scrollar över gränsen.
  for (const series of input.series) {
    if (!series.daily) continue;

    if (series.count > limit) hidden.add(series.key);
    else if (wasHidden.has(series.key) && series.count >= limit)
      hidden.add(series.key);
  }

  // Regel 2: om allt inte ryms i vyn, döljs den serien som har flest
  // förekomster. En serie med en enda förekomst i vyn döljs inte - den är
  // ju inte brusan, bara olycklig.
  const candidates = input.series
    .filter((series) => !hidden.has(series.key) && series.count > 1)
    .sort(
      (left, right) =>
        right.count - left.count || left.key.localeCompare(right.key),
    );

  while (candidates.length > 0 && !fits(hidden)) {
    const next = candidates.shift();
    if (!next) break;
    hidden.add(next.key);
  }

  return { hidden: [...hidden], overflowing: !fits(hidden) };
}

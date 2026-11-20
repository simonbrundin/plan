/**
 * Vilka återkommande serier som ska döljas i tidslinjen.
 *
 * Tidslinjen visar alla händelser som rader i en kolumn. Ibland ryms de inte:
 * en daglig rutin som syns trettio gånger i en månadsvy tar över kolumnen, och
 * då blir de enskilda händelserna omöjliga att läsa.
 *
 * Två regler tar hand om det, i den ordningen:
 *
 * 1. **Antal.** En serie visas med högst ett bestämt antal förekomster. Har
 *    den fler, döljs hela serien - den vanligaste källan till brus.
 * 2. **Döljning.** Räcker det inte, döljs den serie som har flest förekomster
 *    i vyn, så att de enskilda händelserna går att läsa.
 *
 * Döljningen är tyst: en dold serie syns inte i tidslinjen. Det är användarens
 * val - sammanfattningen av serierna togs bort eftersom den skräpade ner
 * vyn.
 */

import { layoutEventRows, type EventRowInput } from "./calendarLayout.ts";

/** En återkommande serie i vyn. */
export interface VisibilitySeries {
  /** Seriens nyckel. */
  key: string;
  /** Seriens titel. */
  title: string;
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
  /**
   * Hur många förekomster en serie får ha innan hela serien döljs.
   *
   * Standard är sju: en serie med åtta förekomster i vyn döljs.
   */
  maxInstances: number;
  /** Om serier får döljas alls. */
  hidingEnabled: boolean;
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
 * Planen är ren: samma underlag ger samma svar. Serierna som var dolda sist
 * går före i urvalet, så att serien inte hoppar fram och tillbaka när antalet
 * förekomster ändras.
 *
 * Antalet förekomster per serie är redan begränsat av urvalet i
 * `sampleSeriesEvents`, så den här planen behöver bara se till att det som
 * är kvar ryms i kolumnen.
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

    // Det som räknas är om kolumnen är full, alltså om rader har trängts ut
    // av att andra rader ligger ovanför. En rad vid fönstrets kant räknas
    // inte - den ligger där den ligger.
    return layoutEventRows(items, {
      from: input.from,
      to: input.to,
      height: input.height,
    }).every((row) => !row.overflow);
  };

  if (!input.hidingEnabled) {
    return { hidden: [], overflowing: !fits(hidden) };
  }

  const limit = Math.max(1, Math.round(input.maxInstances));

  // Regel 1: en serie med fler förekomster än gränsen döljs. Det är den
  // vanligaste källan till brus - en daglig rutin som syns trettio gånger i
  // en månadsvy.
  for (const series of input.series) {
    if (series.count > limit) hidden.add(series.key);
  }

  const wasHidden = new Set(input.hiddenBefore);

  // Regel 2: om allt inte ryms i vyn, döljs serier. Den som har flest
  // förekomster döljs först - den är ju den som tar över kolumnen. En serie
  // med en enda förekomst i vyn döljs inte: den är inte brusan, bara
  // olycklig.
  //
  // Vid lika många förekomster vinner den som var dold sist, så att det är
  // samma serie som försvinner och kommer tillbaka - inte olika serier som
  // växlar plats.
  const candidates = input.series
    .filter((series) => !hidden.has(series.key) && series.count > 1)
    .sort(
      (left, right) =>
        right.count - left.count ||
        Number(wasHidden.has(right.key)) - Number(wasHidden.has(left.key)) ||
        left.key.localeCompare(right.key),
    );

  while (candidates.length > 0 && !fits(hidden)) {
    const next = candidates.shift();
    if (!next) break;
    hidden.add(next.key);
  }

  return { hidden: [...hidden], overflowing: !fits(hidden) };
}

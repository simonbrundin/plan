/**
 * Packning av händelserader i den vertikala kalendern.
 *
 * Varje händelse visas som en rad med ikon, titel och starttid. En rad ska
 * alltid gå att läsa, så den får aldrig bli lägre än `MIN_ROW_PX` - även om
 * händelsen i verkligheten är kort och tidsaxeln är utzoomed. Utan det skulle
 * en timmes händelse i ett månadsfönster bli en pixel hög, och titeln skulle
 * försvinna.
 *
 * Alla rader ligger i en enda kolumn. Två rader kan inte ligga ovanpå varandra,
 * så en rad som krockar med en tidigare rad skjuts ned tills den inte gör
 * det. Radens höjd följer händelsens längd, men begränsas så att en lång
 * händelse inte tar över hela kolumnen.
 *
 * Allt räknas i pixlar, inte i procent. En rad har ett absolut höjdbehov i
 * pixlar, och bara pixlarna kan avgöra om en rad hinner rymmas under en annan.
 */

/** Minsta höjd på en rad, i pixlar. */
export const MIN_ROW_PX = 16;

/** Minsta avstånd mellan två rader, i pixlar. */
export const ROW_GAP_PX = 2;

/** En händelse som ska packas, med start och slut i millisekunder. */
export interface EventRowInput {
  /** Starttid i millisekunder. */
  start: number;
  /** Sluttid i millisekunder. */
  end: number;
}

/** En händelses placering i vyn. */
export interface EventRow {
  /** Övre kanten, i pixlar från toppen av vyn. */
  top: number;
  /** Höjden, i pixlar. */
  height: number;
  /** Sant om raden inte ryms i vyn. */
  clipped: boolean;
}

/**
 * Packar händelser till rader i en enda kolumn.
 *
 * Raderna behåller sin ordning i tiden, men deras placering i vyn räknas
 * utifrån hur stort tidsfönstret är och hur hög vyn är.
 */
export function layoutEventRows(
  items: readonly EventRowInput[],
  options: {
    /** Tidsfönstrets början, i millisekunder. */
    from: number;
    /** Tidsfönstrets slut, i millisekunder. */
    to: number;
    /** Vyns höjd, i pixlar. */
    height: number;
    /** Minsta höjd på en rad, i pixlar. */
    minRow?: number;
    /** Minsta avstånd mellan två rader, i pixlar. */
    gap?: number;
  },
): EventRow[] {
  const { from, to, height } = options;
  const minRow = options.minRow ?? MIN_ROW_PX;
  const gap = options.gap ?? ROW_GAP_PX;
  const span = Math.max(1, to - from);
  const scale = height / span;
  // En rad får aldrig ta mer än en fjärdedel av vyn, så att en lång händelse
  // inte ska ut hela kolumnen.
  const maxRow = Math.max(minRow * 4, height / 4);

  const order = items
    .map((_, index) => index)
    .sort(
      (left, right) =>
        items[left]!.start - items[right]!.start || left - right,
    );

  const rows: EventRow[] = new Array(items.length);
  // Var nästa rad får börja. Raderna staplas under varandra, så att ingen
  // rad döljer en annan.
  let cursor = 0;

  for (const index of order) {
    const item = items[index]!;
    // Bara den del av händelsen som ligger i fönstret räknas. En händelse som
    // går över midnatt ska inte sticka ut under kanten.
    const start = Math.max(item.start, from);
    const end = Math.min(item.end, to);
    const timeTop = Math.max(0, Math.min(height, (start - from) * scale));
    const duration = Math.max(0, (end - start) * scale);

    // Raden börjar på sin tid, men skjuts ned om den skulle hamna under en
    // tidigare rad.
    const top = Math.max(timeTop, cursor);
    const rowHeight = Math.min(Math.max(duration, minRow), maxRow);
    const bottom = top + rowHeight;

    rows[index] = {
      top,
      height: rowHeight,
      clipped: bottom > height,
    };

    cursor = bottom + gap;
  }

  return rows;
}

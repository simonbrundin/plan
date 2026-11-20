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
 * det. Radens höjd följer händelsens längd, exakt proportionellt mot tidsaxeln
 * i vyn - ingen övre gräns. En tidigare version begränsade höjden till en
 * fjärdedel av vyn, och då såg en sex timmars lång händelse lika lång ut som
 * en tolftimmars händelse. Längden stämde då inte med det som visades på
 * axeln, och det blev omöjligt att läsa av hur lång händelsen var.
 *
 * Bara etikettbandet - de översta pixlarna där titeln står - reserverar plats.
 * Annars skulle en lång händelse skjuta undan alla andra händelser som ligger
 * inne i den, och de skulle hamna långt nedanför sin egen tid.
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
  /**
   * Höjden, i pixlar.
   *
   * Följer händelsens längd i förhållande till hela tidsfönstret, men är
   * aldrig lägre än `MIN_ROW_PX`.
   */
  height: number;
  /**
   * Sant om radens titelband inte ryms i vyn.
   *
   * Då visas raden inte i tidslinjen, utan i listan över de händelser som
   * inte hinner med.
   */
  clipped: boolean;
  /**
   * Sant om raden har trängts ut ur vyn av att andra rader ligger ovanför.
   *
   * Det betyder att kolumnen är full. En rad vid fönstrets kant är däremot
   * inte trängd ut - den ligger där den ligger, och syns bara delvis.
   */
  overflow: boolean;
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
    let top = Math.max(timeTop, cursor);
    // Höjden följer händelsens längd, proportionellt mot tidsaxeln. Det finns
    // ingen övre gräns - en lång händelse ska se ut att vara lång. Den får
    // bara aldrig bli lägre än den lästa lägsta höjden.
    const rowHeight = Math.max(duration, minRow);
    // Bara etikettbandet reserverar plats. Titeln står i bandets översta
    // pixlar, så en rad som ligger längre ned i tiden får ligga ovanpå en
    // lång händelses tomma yta.
    const band = minRow;
    // En rad vid fönstrets kant flyttas upp en aning, så att titelbandet
    // syns. Den flyttas inte före en tidigare rad.
    if (top < height && top + band > height) {
      top = Math.max(cursor, height - band);
    }
    // Radens titelband måste rymmas i vyn, annars är raden inte läsbar.
    const clipped = top + band > height;
    // Raden har trängts ut ur vyn om en tidigare rad tog all plats. Det är
    // det som gör att kolumnen är full - inte att raden ligger vid
    // fönstrets kant.
    const overflow = top >= height;

    rows[index] = {
      top,
      height: Math.min(rowHeight, Math.max(band, height - top)),
      clipped,
      overflow,
    };

    cursor = top + band + gap;
  }

  return rows;
}

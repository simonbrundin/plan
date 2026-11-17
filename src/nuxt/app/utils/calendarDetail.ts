/**
 * Detaljnivåer för den vertikala kalendern.
 *
 * Tidslinjen kan visa allt från timmar till flera år, och samma händelse kan
 * inte visas likadant i alla de fallen. Därför väljs en detaljnivå utifrån hur
 * mycket tid som syns:
 *
 * - full    - timmar. Hela kortet visas med ikon, titel och klockslag.
 * - compact - dagar. Kortet visas med ikon och titel. Klockslaget flyttas till
 *             tooltipen eftersom kortet annars blir för tunt för att kunna
 *             visa både ikon och text.
 * - daily   - veckor och månader. Händelserna grupperas per dag och visas
 *             som prickar med antal. Klick på en prick hoppar in till dagen.
 * - monthly - år. Händelserna grupperas per månad och visas som punkter vars
 *             storlek speglar antalet. Klick hoppar in till månaden.
 * - yearly  - flera år. Händelserna grupperas per år.
 *
 * Grupperingen görs i lokal tid, eftersom det är användaren som ska känna
 * igen en dag ("måndag 12 november"), inte UTC.
 *
 * Gränserna är satta så att antalet markörer som ska rymmas i vyhöjden
 * förblir hanterbart. En prick per dag i ett årfönster skulle kräva 365
 * prickar, vilket inte går att läsa eller klicka på.
 */

export type ScaleLevel = "hour" | "day" | "month" | "year";

export type DetailLevel = "full" | "compact" | "daily" | "monthly" | "yearly";

/** Gränsen mellan timmar och dagar. */
export const HOUR_LEVEL_MAX = 48 * 60 * 60 * 1000;

/** Gränsen mellan kompakta kort och prickar per dag. */
export const COMPACT_LEVEL_MAX = 7 * 24 * 60 * 60 * 1000;

/** Gränsen mellan prickar per dag och prickar per månad. */
export const DAILY_LEVEL_MAX = 45 * 24 * 60 * 60 * 1000;

/** Gränsen mellan prickar per månad och prickar per år. */
export const MONTHLY_LEVEL_MAX = 3 * 365 * 24 * 60 * 60 * 1000;

/**
 * Marginal som måste överstigas innan detaljnivån får lämnas mot mer tid.
 *
 * Utan marginalen fladdrar nivån när användaren scrollar långsamt över en
 * gräns: nivån byts fram och tillbaka i stället för att stanna. Med
 * marginalen måste tiden passera gränsen med 8 procent extra innan nivån
 * byts tillbaka.
 */
export const DETAIL_HYSTERESIS = 0.08;

/** Detaljnivåerna i ordning från minst till mest tid. */
const DETAIL_ORDER: DetailLevel[] = [
  "full",
  "compact",
  "daily",
  "monthly",
  "yearly",
];

/** Minsta tidsfönster som varje nivå gäller för. */
const DETAIL_MIN: number[] = [
  0,
  HOUR_LEVEL_MAX,
  COMPACT_LEVEL_MAX,
  DAILY_LEVEL_MAX,
  MONTHLY_LEVEL_MAX,
];

export function resolveScaleLevel(duration: number): ScaleLevel {
  if (duration < HOUR_LEVEL_MAX) return "hour";
  if (duration < 45 * 24 * 60 * 60 * 1000) return "day";
  if (duration < 2 * 365 * 24 * 60 * 60 * 1000) return "month";
  return "year";
}

/**
 * Den detaljnivå som gäller för ett tidsfönster, utan hänsyn till vad som
 * visades tidigare.
 */
export function resolveDetailLevel(duration: number): DetailLevel {
  let index = 0;

  while (
    index < DETAIL_ORDER.length - 1 &&
    duration >= DETAIL_MIN[index + 1]
  ) {
    index++;
  }

  return DETAIL_ORDER[index];
}

/**
 * Detaljnivån med hysteresis.
 *
 * Mot mindre tid byts nivån direkt, eftersom det är entydigt att tiden
 * blivit kortare. Mot mer tid krävs att gränsen överstigs med marginal,
 * så att en gräns som passeras i båda riktningar inte fladdrar.
 */
export function resolveDetailLevelWithHysteresis(
  duration: number,
  previous: DetailLevel,
): DetailLevel {
  let index = DETAIL_ORDER.indexOf(previous);

  if (index === -1) index = 0;

  // Mer tid: gränsen måste överstigas med marginal.
  while (
    index < DETAIL_ORDER.length - 1 &&
    duration >= DETAIL_MIN[index + 1] * (1 + DETAIL_HYSTERESIS)
  ) {
    index++;
  }

  // Mindre tid: den vanliga gränsen räcker.
  while (index > 0 && duration < DETAIL_MIN[index]) {
    index--;
  }

  return DETAIL_ORDER[index];
}

export interface CalendarEventLike {
  id: number;
  title: string;
  icon?: string;
  color?: string;
  start_at: string;
  end_at: string;
  status?: string;
  recurrence_rule?: string | null;
  parent_event_id?: number | null;
}

/**
 * Serien som en händelse hör till.
 *
 * En expanderad förekomst bär fortfarande seriens `recurrence_rule`, så
 * serien identifieras av master-id. En override-rad hör till den serie den
 * ersätter, via `parent_event_id`.
 */
export function recurrenceSeriesKey(event: CalendarEventLike): string | null {
  if (!event.recurrence_rule && event.parent_event_id == null) return null;

  return event.parent_event_id == null
    ? `series:${event.id}`
    : `series:${event.parent_event_id}`;
}

/** Etiketter för återupprepningsregler, på svenska. */
const RECURRENCE_LABELS: Record<string, string> = {
  HOURLY: "varje timme",
  DAILY: "varje dag",
  WEEKLY: "varje vecka",
  MONTHLY: "varje månad",
  YEARLY: "varje år",
};

/**
 * Beskriver en återupprepningsregel på svenska.
 *
 * Kalendern visar en kort etikett om vad som upprepas, så att en förekomst
 * går att känna igen som en del av en serie utan att regeln behöver läsas.
 */
export function describeRecurrence(rule: string | null | undefined): string {
  if (!rule) return "";

  const parts: Record<string, string> = {};

  for (const part of rule.split(";")) {
    const [key, value] = part.split("=");
    if (key && value) parts[key.trim().toUpperCase()] = value.trim().toUpperCase();
  }

  const freq = parts.FREQ;
  if (!freq) return "";

  const base = RECURRENCE_LABELS[freq];
  if (!base) return "";

  const interval = Number(parts.INTERVAL ?? 1);
  if (!Number.isFinite(interval) || interval <= 1) return base;

  // "varje vecka" -> "var 2:e vecka"
  return `var ${interval}:e ${base.replace("varje ", "")}`;
}

export interface CalendarMarker {
  /** Stabil nyckel för rendering. */
  key: string;
  /** Tooltip-text som beskriver vad som finns i gruppen. */
  title: string;
  /** Representativ ikon, hämtad från den första händelsen med ikon. */
  icon?: string;
  /** Representativ färg från den första händelsen med färg. */
  color?: string;
  /** Början på den grupperade perioden. */
  startAt: number;
  /** Slutet på den grupperade perioden. */
  endAt: number;
  /** Antal händelser i gruppen. */
  count: number;
  /** Sant om någon händelse i gruppen är preliminär. */
  tentative: boolean;
  /** Sant om någon händelse i gruppen hör till en återkommande serie. */
  recurring: boolean;
}

const MONTH_NAMES = [
  "januari",
  "februari",
  "mars",
  "april",
  "maj",
  "juni",
  "juli",
  "augusti",
  "september",
  "oktober",
  "november",
  "december",
];

const WEEKDAY_NAMES = [
  "söndag",
  "måndag",
  "tisdag",
  "onsdag",
  "torsdag",
  "fredag",
  "lördag",
];

const formatDayLabel = (timestamp: number) => {
  const date = new Date(timestamp);
  return `${WEEKDAY_NAMES[date.getDay()]} ${date.getDate()} ${MONTH_NAMES[date.getMonth()]}`;
};

const formatMonthLabel = (timestamp: number) => {
  const date = new Date(timestamp);
  return `${MONTH_NAMES[date.getMonth()]} ${date.getFullYear()}`;
};

/**
 * Beskriver innehållet i en grupperad period.
 *
 * Återkommande förekomster visas en gång per serie med sitt antal. Annars
 * skulle en serie som upprepas 52 gånger skriva ut samma titel 52 gånger,
 * vilket gör etiketten oläslig.
 */
const describeBucket = (bucket: CalendarEventLike[], label: string) => {
  const word = bucket.length === 1 ? "händelse" : "händelser";
  const bySeries = new Map<string, { title: string; count: number }>();

  for (const event of bucket) {
    // Enstaka händelser räknas var för sig, så de behöver egen nyckel.
    const key = recurrenceSeriesKey(event) ?? `event:${event.id}`;
    const existing = bySeries.get(key);

    if (existing) {
      existing.count += 1;
    } else {
      bySeries.set(key, { title: event.title, count: 1 });
    }
  }

  const entries = [...bySeries.values()];
  const shown = entries
    .slice(0, 3)
    .map((entry) => (entry.count > 1 ? `${entry.title} ×${entry.count}` : entry.title))
    .join(", ");
  const rest = entries.length > 3 ? ` +${entries.length - 3} till` : "";

  return `${bucket.length} ${word} ${label}: ${shown}${rest}`;
};

const collect = (
  events: CalendarEventLike[],
  keyOf: (date: Date) => string,
  periodOf: (date: Date) => { startAt: number; endAt: number; label: string },
) => {
  const groups = new Map<string, CalendarEventLike[]>();

  for (const event of events) {
    const date = new Date(event.start_at);
    const key = keyOf(date);
    const bucket = groups.get(key);

    if (bucket) {
      bucket.push(event);
    } else {
      groups.set(key, [event]);
    }
  }

  return [...groups.entries()]
    .map(([key, bucket]) => {
      const period = periodOf(new Date(bucket[0].start_at));

      return {
        key,
        title: describeBucket(bucket, period.label),
        icon: bucket.find((event) => event.icon)?.icon,
        color: bucket.find((event) => event.color)?.color,
        startAt: period.startAt,
        endAt: period.endAt,
        count: bucket.length,
        tentative: bucket.some((event) => event.status === "tentative"),
        recurring: bucket.some((event) => recurrenceSeriesKey(event) !== null),
      } satisfies CalendarMarker;
    })
    .sort((a, b) => a.startAt - b.startAt);
};

/** Grupperar händelser per lokal dag. */
export function groupEventsByDay(events: CalendarEventLike[]): CalendarMarker[] {
  return collect(
    events,
    (date) => `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`,
    (date) => {
      const startAt = new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
      const endAt = new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1).getTime();

      return { startAt, endAt, label: formatDayLabel(startAt) };
    },
  );
}

/** Grupperar händelser per lokal månad. */
export function groupEventsByMonth(events: CalendarEventLike[]): CalendarMarker[] {
  return collect(
    events,
    (date) => `${date.getFullYear()}-${date.getMonth()}`,
    (date) => {
      const startAt = new Date(date.getFullYear(), date.getMonth(), 1).getTime();
      const endAt = new Date(date.getFullYear(), date.getMonth() + 1, 1).getTime();

      return { startAt, endAt, label: formatMonthLabel(startAt) };
    },
  );
}

/** Grupperar händelser per lokal år. */
export function groupEventsByYear(events: CalendarEventLike[]): CalendarMarker[] {
  return collect(
    events,
    (date) => `${date.getFullYear()}`,
    (date) => {
      const startAt = new Date(date.getFullYear(), 0, 1).getTime();
      const endAt = new Date(date.getFullYear() + 1, 0, 1).getTime();

      return { startAt, endAt, label: `${date.getFullYear()}` };
    },
  );
}

/**
 * Fördelar intervall på kolumner inom grupper av överlappande intervall.
 *
 * Två händelser som ligger långt ifrån varandra behöver inte dela bredd
 * med varandra. Därför grupperas de: en grupp bildas av händelser som
 * överlappar i tid, och gruppen får så många kolumner som den behöver.
 *
 * Returnerar både kolumnen och hur många kolumner gruppen totalt har, så
 * att varje händelse kan ges en bredd som passar just sin grupp.
 */
export function assignClusteredLanes(
  intervals: Array<[number, number]>,
  minGap: number,
  maxLanes = 6,
): Array<{ lane: number; lanes: number }> {
  const clusters: Array<{ laneEnds: number[] }> = [];
  const clusterOf: number[] = [];
  const laneOf: number[] = [];

  for (const [start, end] of intervals) {
    let placed = false;

    for (let index = 0; index < clusters.length; index++) {
      const cluster = clusters[index];
      const latest = Math.max(...cluster.laneEnds);

      // Klumpen är inte längre aktiv, så den här händelsen hör inte ihop
      // med den.
      if (latest + minGap <= start) continue;

      let lane = cluster.laneEnds.findIndex((laneEnd) => laneEnd + minGap <= start);

      if (lane === -1) {
        lane = cluster.laneEnds.length;

        if (lane < maxLanes) cluster.laneEnds.push(end);
      } else {
        cluster.laneEnds[lane] = Math.max(cluster.laneEnds[lane], end);
      }

      clusterOf.push(index);
      laneOf.push(lane);
      placed = true;
      break;
    }

    if (!placed) {
      clusters.push({ laneEnds: [end] });
      clusterOf.push(clusters.length - 1);
      laneOf.push(0);
    }
  }

  // Först när alla intervall är placerade vet vi hur många kolumner varje
  // klump faktiskt använder. Händelser som hamnade i samma klump måste
  // dela bredden mellan sig, även om de lades in före varandra.
  const lanesPerCluster = clusters.map((cluster) => cluster.laneEnds.length);

  return clusterOf.map((cluster, index) => ({
    lane: laneOf[index],
    lanes: lanesPerCluster[cluster],
  }));
}

/**
 * Fördelar markörer på kolumner ("lanes") så att de aldrig överlappar.
 *
 * Tidslinjen är vertikal, så två markörer som ligger nära varandra i tid
 * hamnar ovanpå varandra. Det gör dem oläsbara och oklickbara. Genom att
 * packa dem i olika kolumner bredvid tidslinjen kan de samexistera.
 *
 * `positions` är markörernas vertikala läge i procent av tidsfönstret och
 * `minGap` är minsta tillåtna avstånd i samma enhet.
 */
export function assignLanes(positions: number[], minGap: number, maxLanes = 6): number[] {
  const laneEnds: number[] = [];

  return positions.map((position) => {
    let lane = laneEnds.findIndex((end) => end + minGap < position);

    if (lane === -1) {
      lane = Math.min(laneEnds.length, maxLanes - 1);
      if (lane === laneEnds.length) laneEnds.push(position);
    }

    laneEnds[lane] = position;

    return lane;
  });
}

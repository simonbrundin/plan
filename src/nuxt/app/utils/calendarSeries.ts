/**
 * Återkommande serier i kalendern.
 *
 * Serier identifieras av master-id, och deras upprepningsregel beskrivs med
 * en kort etikett på svenska. De här hjälparna används både i tidslinjen och
 * i sammanfattningen av serierna ovanför den.
 */

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
  /**
   * Hur många verkliga förekomster den här förekomsten står för.
   *
   * En serie som tunnas ut för att den blir för stor representeras av ett
   * urval av förekomster. Vikten gör att antalen ändå blir rätta.
   */
  occurrence_weight?: number;
  /** Det sanna antalet förekomster i serien, när den har tunnats ut. */
  occurrence_total?: number;
  /**
   * Den period serien tunnats ut i, när den har tunnats ut.
   *
   * Varje visad förekomst står för alla förekomster i sin period. Det
   * behövs för att kunna beskriva urvalet rätt.
   */
  occurrence_period?: "day" | "month";
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

/**
 * Sant om regeln beskriver en serie som återkommer varje dag.
 *
 * Dagliga serier är den vanligaste orsaken till att en tidslinje blir
 * full. De döljs därför först när det blir för mycket att visa.
 */
export function isDailyRecurrence(rule: string | null | undefined): boolean {
  if (!rule) return false;

  const parts: Record<string, string> = {};

  for (const part of rule.split(";")) {
    const [key, value] = part.split("=");
    if (key && value) parts[key.trim().toUpperCase()] = value.trim().toUpperCase();
  }

  if (parts.FREQ !== "DAILY") return false;

  const interval = Number(parts.INTERVAL ?? 1);

  return Number.isFinite(interval) && interval <= 1;
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

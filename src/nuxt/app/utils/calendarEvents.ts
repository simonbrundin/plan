import type { CalendarEvent } from "~/composables/useCalendarApi";

/**
 * Hjälpare för att agera på kalenderhändelser.
 *
 * Kalendern visar en serie som enskilda förekomster, och varje förekomst
 * kan ha en egen rad (en override). Därför behöver klienten veta vilken rad
 * som egentligen ska ändras eller tas bort - seriens masterrad eller
 * förekomstens egen rad.
 */

/**
 * Sant om händelsen är en förekomst i en serie.
 *
 * En förekomst har alltid ett occurrence_id. En förekomst med en egen rad
 * (en override) har dessutom ett parent_event_id.
 */
export function isSeriesOccurrence(event: CalendarEvent): boolean {
  return Boolean(event.occurrence_id) || Boolean(event.parent_event_id);
}

/**
 * Id:t för raden som händelsen hör till i serien.
 *
 * För en förekomst utan egen rad är det seriens masterrad. För en förekomst
 * med egen rad är masterraden föräldern.
 */
export function parentEventId(event: CalendarEvent): number {
  return event.parent_event_id ?? event.id;
}

/**
 * Payload som skapar en kopia av en händelse.
 *
 * Kopian får samma titel, beskrivning, ikon, kalender och mål. En förekomst
 * kopplas loss från sin serie - kopian blir en egen händelse med samma
 * tid. En hel serie kopplas däremot med samma upprepningsregel, så att
 * kopien upprepas precis som originalet.
 */
export function duplicateEventPayload(event: CalendarEvent) {
  const payload = {
    title: event.title,
    description: event.description || "",
    icon: event.icon || "",
    start_at: event.start_at,
    end_at: event.end_at,
    time_zone: event.time_zone,
    ...(event.calendar_id ? { calendar_id: event.calendar_id } : {}),
    ...(event.goal_id ? { goal_id: event.goal_id } : {}),
  };

  if (isSeriesOccurrence(event)) return payload;

  return {
    ...payload,
    ...(event.recurrence_rule
      ? { recurrence_rule: event.recurrence_rule }
      : {}),
  };
}

/**
 * Klipper en upprepningsregel vid en tidpunkt.
 *
 * Regeln ska sluta strax före den valda förekomsten, så att förekomsten
 * och de följande försvinner ur serien. En befintlig UNTIL eller COUNT tas
 * bort först - annars skulle serien sluta där den redan slutade, och den
 * nya gränsen skulle ignoreras.
 */
export function truncateRecurrenceRule(rule: string, until: Date): string {
  const parts = rule
    .split(";")
    .map((part) => part.trim())
    .filter((part) => part.length > 0)
    .filter((part) => !/^(UNTIL|COUNT)=/i.test(part));

  return [...parts, `UNTIL=${toRRuleDate(until)}`].join(";");
}

/**
 * Skriver en tid i det format som RRULE använder: 20261118T120000Z.
 */
export function toRRuleDate(date: Date): string {
  return date.toISOString().replace(/[-:]/g, "").replace(/\.\d{3}Z$/, "Z");
}

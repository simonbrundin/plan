import type { Calendar } from '~/types/calendar'

/**
 * Användarens standardkalender.
 *
 * Standardkalendern är den kalender som nya händelser hamnar i när ingen
 * kalender väljs uttryckligen. Det finns högst en per användare - servern
 * ser till att det inte kan bli flera.
 */
export const findDefaultCalendar = (calendars: Calendar[]): Calendar | null =>
  calendars.find((calendar) => calendar.is_default) ?? null

/**
 * Id:t för standardkalendern, eller null om ingen kalender är satt som
 * standard.
 */
export const findDefaultCalendarId = (calendars: Calendar[]): number | null =>
  findDefaultCalendar(calendars)?.id ?? null

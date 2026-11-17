import type { Calendar } from '~/types/calendar'
import { resolveEventColor, type EventColorSource } from '~/utils/calendarColors'

/**
 * Delat tillstånd för kalendrarna.
 *
 * Kalendrarna och deras synlighet ligger i global state, eftersom flera
 * ställen behöver samma information: tidslinjen visar händelserna, och
 * formulären och kalenderhanteraren visar vilka kalendrar som finns.
 */
export function useCalendars() {
  const calendars = useState<Calendar[]>('calendars:list', () => []);
  const hiddenCalendarIds = useState<number[]>('calendars:hidden', () => []);
  const isLoaded = useState('calendars:loaded', () => false);

  const { loadCalendars, createCalendar, updateCalendar, deleteCalendar } = useCalendarApi();

  /** Kalendrarna som ett uppslag, så att färgen kan slås upp per händelse. */
  const calendarMap = computed(
    () => new Map(calendars.value.map((calendar) => [calendar.id, calendar])),
  );

  const load = async (force = false) => {
    if (isLoaded.value && !force) return;

    try {
      calendars.value = await loadCalendars();
      isLoaded.value = true;
    } catch (error) {
      // Kalendern ska gå att använda även om listan inte kan läsas - då
      // saknas bara kalenderväljaren.
      console.warn('Kunde inte läsa kalendrar:', error);
    }
  };

  const add = async (name: string, color: string) => {
    const created = await createCalendar({ name: name.trim(), color });
    calendars.value = [...calendars.value, created];
    return created;
  };

  const save = async (calendarId: number, changes: { name?: string; color?: string }) => {
    const updated = await updateCalendar(calendarId, changes);
    calendars.value = calendars.value.map((calendar) =>
      calendar.id === calendarId ? updated : calendar,
    );
    return updated;
  };

  const remove = async (calendarId: number) => {
    await deleteCalendar(calendarId);
    calendars.value = calendars.value.filter((calendar) => calendar.id !== calendarId);
    // En borttagen kalender ska inte ligga kvar som dold.
    hiddenCalendarIds.value = hiddenCalendarIds.value.filter((id) => id !== calendarId);
  };

  const isHidden = (calendarId: number) => hiddenCalendarIds.value.includes(calendarId);

  /** Vänder en kalenders synlighet. */
  const toggleVisibility = (calendarId: number) => {
    if (isHidden(calendarId)) {
      hiddenCalendarIds.value = hiddenCalendarIds.value.filter((id) => id !== calendarId);
      return;
    }
    hiddenCalendarIds.value = [...hiddenCalendarIds.value, calendarId];
  };

  /** Den färg en händelse ska visas med. */
  const colorFor = (event: EventColorSource) => resolveEventColor(event, calendarMap.value);

  return {
    calendars,
    hiddenCalendarIds,
    isLoaded,
    calendarMap,
    load,
    add,
    save,
    remove,
    isHidden,
    toggleVisibility,
    colorFor,
  };
}

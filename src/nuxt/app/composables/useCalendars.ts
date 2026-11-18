import type { Calendar } from '~/types/calendar'
import { resolveEventColor, type EventColorSource } from '~/utils/calendarColors'
import { findDefaultCalendar, findDefaultCalendarId } from '~/utils/calendars'

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

  /** Användarens standardkalender, om någon är satt. */
  const defaultCalendar = computed(() => findDefaultCalendar(calendars.value));

  /** Id:t för standardkalendern, eller null om ingen är satt. */
  const defaultCalendarId = computed(() => findDefaultCalendarId(calendars.value));

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

  const add = async (name: string, color: string, isDefault = false) => {
    const created = await createCalendar({ name: name.trim(), color, is_default: isDefault });
    calendars.value = [...calendars.value, created];
    return created;
  };

  const save = async (
    calendarId: number,
    changes: { name?: string; color?: string; is_default?: boolean },
  ) => {
    const updated = await updateCalendar(calendarId, changes);
    calendars.value = calendars.value.map((calendar) => {
      if (calendar.id === calendarId) return updated;
      // Servern avmarkerar den tidigare standardkalendern när en ny sätts,
      // så lokal state måste följa med.
      if (changes.is_default === true && calendar.is_default) {
        return { ...calendar, is_default: false };
      }
      return calendar;
    });
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
    defaultCalendar,
    defaultCalendarId,
    load,
    add,
    save,
    remove,
    isHidden,
    toggleVisibility,
    colorFor,
  };
}

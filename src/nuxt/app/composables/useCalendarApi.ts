import type { Calendar } from '~/types/calendar'

export interface CalendarEvent {
  id: number
  occurrence_id?: string
  goal_id?: number
  /** Vilken kalender händelsen ligger i. */
  calendar_id?: number
  title: string
  description: string
  /**
   * Valfri ikon: en emoji (t.ex. "🏃") eller ett Iconify-namn
   * (t.ex. "heroicons:star").
   */
  icon?: string
  start_at: string
  end_at: string
  time_zone: string
  all_day: boolean
  status: 'confirmed' | 'tentative' | 'cancelled'
  recurrence_rule?: string
  parent_event_id?: number
  original_start_at?: string
  /**
   * Hur många verkliga förekomster den här förekomsten representerar.
   *
   * En serie med fler förekomster än expanderaren klarar tunnas ut, så att
   * varje visad förekomst står för flera verkliga. Fältet gör det möjligt
   * att räkna rätt i stället för att visa ett godtyckligt antal.
   */
  occurrence_weight?: number
  /**
   * Det sanna antalet förekomster i serien, när den har tunnats ut.
   */
  occurrence_total?: number
  /**
   * Den period serien tunnats ut i, när den har tunnats ut.
   *
   * Varje visad förekomst står för alla förekomster i sin period. Det
   * behövs för att kunna beskriva urvalet rätt.
   */
  occurrence_period?: "day" | "month"
}

export function useCalendarApi() {
  const loadCalendarEvents = async (from: Date, to: Date): Promise<CalendarEvent[]> => {
    return await $fetch<CalendarEvent[]>('/api/proxy/calendar-events', {
      query: {
        from: from.toISOString(),
        to: to.toISOString(),
      },
    })
  }

  /**
   * Hämtar alla händelser för användaren på en gång.
   *
   * Återkommande serier kommer oexpanderade, så att klienten kan expandera
   * dem lokalt och slippa fråga servern varje gång tidsfönstret flyttas.
   */
  const loadAllCalendarEvents = async (): Promise<CalendarEvent[]> => {
    return await $fetch<CalendarEvent[]>('/api/proxy/calendar-events', {
      query: {
        scope: 'all',
      },
    })
  }

  const updateCalendarEvent = async (eventId: number, event: {
    title?: string
    description?: string
    icon?: string
    start_at?: string
    end_at?: string
    time_zone?: string
    all_day?: boolean
    status?: 'confirmed' | 'tentative' | 'cancelled'
    goal_id?: number
    calendar_id?: number
    recurrence_rule?: string
  }) => {
    return await $fetch<CalendarEvent>(`/api/proxy/calendar-events/${eventId}`, {
      method: 'PATCH',
      body: event,
    })
  }

  const createCalendarOccurrenceOverride = async (eventId: number, event: {
    original_start_at: string
    start_at: string
    end_at: string
    title?: string
    description?: string
    icon?: string
    time_zone?: string
    all_day?: boolean
    status?: 'confirmed' | 'tentative' | 'cancelled'
    goal_id?: number
    calendar_id?: number
  }) => {
    return await $fetch<CalendarEvent>(`/api/proxy/calendar-events/${eventId}/occurrences`, {
      method: 'POST',
      body: event,
    })
  }

  const splitCalendarSeries = async (eventId: number, event: {
    original_start_at: string
    start_at: string
    end_at: string
    recurrence_rule: string
    title?: string
    description?: string
    icon?: string
    time_zone?: string
    all_day?: boolean
    goal_id?: number
    calendar_id?: number
  }) => {
    return await $fetch<CalendarEvent>(`/api/proxy/calendar-events/${eventId}/split`, {
      method: 'POST',
      body: event,
    })
  }

  const createCalendarEvent = async (event: {
    title: string
    description?: string
    icon?: string
    start_at: string
    end_at: string
    time_zone?: string
    all_day?: boolean
    goal_id?: number
    calendar_id?: number
    recurrence_rule?: string
  }) => {
    return await $fetch<CalendarEvent>('/api/proxy/calendar-events', {
      method: 'POST',
      body: event,
    })
  }

  /**
   * Tar bort en kalenderhändelse.
   *
   * För en återkommande serie tas hela serien bort, inklusive dess
   * förekomster. Enstaka förekomster tas bort med en avstängd override,
   * se `createCalendarOccurrenceOverride`.
   */
  const deleteCalendarEvent = async (eventId: number): Promise<void> => {
    await $fetch(`/api/proxy/calendar-events/${eventId}`, { method: "DELETE" })
  }

  /** Hämtar användarens kalendrar. */
  const loadCalendars = async (): Promise<Calendar[]> => {
    return await $fetch<Calendar[]>('/api/proxy/calendars')
  }

  /** Skapar en ny kalender med eget namn och färg. */
  const createCalendar = async (calendar: {
    name: string
    color?: string
    icon?: string
    is_default?: boolean
  }): Promise<Calendar> => {
    return await $fetch<Calendar>('/api/proxy/calendars', {
      method: 'POST',
      body: calendar,
    })
  }

  /** Ändrar en kalenders namn, färg, ikon eller om den är standardkalender. */
  const updateCalendar = async (
    calendarId: number,
    calendar: { name?: string; color?: string; icon?: string; is_default?: boolean },
  ): Promise<Calendar> => {
    return await $fetch<Calendar>(`/api/proxy/calendars/${calendarId}`, {
      method: 'PATCH',
      body: calendar,
    })
  }

  /** Tar bort en kalender. Händelserna i den finns kvar utan kalender. */
  const deleteCalendar = async (calendarId: number): Promise<void> => {
    await $fetch(`/api/proxy/calendars/${calendarId}`, { method: 'DELETE' })
  }

  return {
    loadCalendarEvents,
    loadAllCalendarEvents,
    createCalendarEvent,
    updateCalendarEvent,
    createCalendarOccurrenceOverride,
    splitCalendarSeries,
    deleteCalendarEvent,
    loadCalendars,
    createCalendar,
    updateCalendar,
    deleteCalendar,
  }
}

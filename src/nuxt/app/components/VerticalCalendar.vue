<script setup lang="ts">
import type { CalendarEvent } from '~/composables/useCalendarApi'
import type { Goal } from '~/types/goal'
import { CalendarExpander } from '~/utils/calendarRecurrence'
import { DEFAULT_EVENT_COLOR, isEventHidden } from '~/utils/calendarColors'
import {
  duplicateEventPayload,
  isSeriesOccurrence,
  parentEventId,
  truncateRecurrenceRule,
} from '~/utils/calendarEvents'
import { describeRecurrence, recurrenceSeriesKey } from '~/utils/calendarSeries'
import { MIN_ROW_PX, layoutEventRows } from '~/utils/calendarLayout'
import { planVisibility } from '~/utils/calendarVisibility'
import {
  carriesGoal,
  GOAL_DRAG_MIME,
  millisecondsPerPixel,
  parseGoalDrag,
  shiftEventTimes,
  snapStepMs,
  snapTimestamp,
  timeAtOffset,
  type GoalDragPayload,
} from '~/utils/calendarScheduling'

const MIN_DURATION = 2 * 60 * 60 * 1000
// Date stöder ett mycket stort, men tekniskt ändligt, intervall.
const MAX_DATE_DURATION = 8.64e15 - Date.now()
const DEFAULT_DURATION = 24 * 60 * 60 * 1000
const ZOOM_SENSITIVITY = 650

const viewport = ref<HTMLElement | null>(null)
const shell = ref<HTMLElement | null>(null)
const startTime = ref(Date.now())
const visibleDuration = ref(DEFAULT_DURATION)
const targetDuration = ref(DEFAULT_DURATION)
const isDragging = ref(false)
const dragMode = ref<'pending' | 'zoom' | null>(null)

const endTime = computed(() => startTime.value + visibleDuration.value)
const now = ref(Date.now())
// Alla händelser hämtas en gång när kalendern öppnas. Serier expanderas
// lokalt till det tidsfönster som syns, så att scrollning och zoom inte
// kräver nya anrop till servern.
const allEvents = ref<CalendarEvent[]>([])
const goalsStore = useGoalsStore()
const expander = new CalendarExpander()

// Ett event som är kopplat till ett färdigt mål ska inte längre synas i
// tidslinjen. Målgrafen ligger i Pinia så att kalendern reagerar direkt när
// målet markeras som klart på en annan vy.
const completedGoalIds = computed(() => {
  // Använd den kompletta listan från /goals som bas. Graph-endpointen kan
  // filtrera bort färdiga mål, och får därför inte ersätta den listan.
  const completed = new Set(
    goals.value.filter((goal) => !!goal.finished).map((goal) => goal.id),
  )

  // Store-statusen gör att kalendern även reagerar direkt när ett mål
  // färdigmarkeras medan kalendern redan är öppen.
  for (const goal of goalsStore.goals) {
    if (goal.finished) completed.add(goal.id)
  }

  return completed
})

// Händelser i dolda kalendrar tas bort ur vyn, så att det går att fokusera
// en enda kalender - till exempel bara jobbet.
const visibleEvents = computed(() =>
  expander
    .visible(allEvents.value, startTime.value, endTime.value)
    .filter((event) => !isEventHidden(event, hiddenCalendarIds.value))
    .filter((event) => !event.goal_id || !completedGoalIds.value.has(event.goal_id)),
)
// Färgen löses upp här, så att kort och prickar alltid visar samma färg:
// händelsens egen färg, annars kalenderns färg.
const coloredEvents = computed<CalendarEvent[]>(() =>
  visibleEvents.value.map((event) => ({ ...event, color: resolveColor(event) })),
)

// En återkommande serie i vyn. Serierna räknas bara - de syns inte i en egen
// sammanfattning, utan används av döljningsreglerna för att se vilken serie
// som är för brusan för att få plats.
interface RecurringSeriesStats {
  key: string
  title: string
  /** Antalet förekomster i vyn. */
  count: number
}

// Regeln för varje serie i vyn. Override-rader bär inte själva regeln, så
// den slås upp via seriens nyckel.
const seriesRules = computed(() => {
  const rules = new Map<string, string>()

  for (const event of visibleEvents.value) {
    const key = recurrenceSeriesKey(event)
    if (!key || !event.recurrence_rule) continue

    rules.set(key, event.recurrence_rule)
  }

  return rules
})

const recurringSeries = computed<RecurringSeriesStats[]>(() => {
  const series = new Map<string, RecurringSeriesStats>()

  for (const event of visibleEvents.value) {
    const key = recurrenceSeriesKey(event)
    if (!key) continue

    // Vikten räknas in, så att antalet stämmer även när serien visas som
    // ett urval av förekomster.
    const weight = event.occurrence_weight ?? 1
    const existing = series.get(key)

    if (existing) {
      existing.count += weight
      continue
    }

    series.set(key, { key, title: event.title, count: weight })
  }

  return [...series.values()]
})

// Etiketten som visas på kortet för en återkommande förekomst.
const recurrenceLabel = (event: CalendarEvent) => {
  const key = recurrenceSeriesKey(event)
  if (!key) return ''

  return describeRecurrence(seriesRules.value.get(key)) || 'återkommande'
}
const showCreateEvent = ref(false)
const editingEvent = ref<CalendarEvent | null>(null)
const editScope = ref<'occurrence' | 'series' | 'following'>('occurrence')
const savingEvent = ref(false)
const deletingEvent = ref(false)
const duplicatingEvent = ref(false)
const errorMessage = ref('')
const managerOpen = ref(false)
const draftEvent = reactive({
  title: '',
  description: '',
  startAt: '',
  endAt: '',
  goalId: '',
  recurrenceRule: '',
  icon: '',
  // Vilken kalender händelsen ligger i. Tomt betyder "ingen kalender".
  calendarId: '',
  // Platser där händelsen sker. Flera platser har OR-semantik.
  placeIds: [] as number[],
})
const iconPickerOpen = ref(false)
const goals = ref<Goal[]>([])
const {
  loadAllCalendarEvents,
  createCalendarEvent,
  updateCalendarEvent,
  createCalendarOccurrenceOverride,
  splitCalendarSeries,
  deleteCalendarEvent,
} = useCalendarApi()
const { loadAllGoals } = useGoalApi()
// Platser: händelser kan kopplas till platser, ungefär som mål.
const { places: allPlaces, fetchPlaces, setEventPlaces } = usePlaces()
const {
  calendars,
  hiddenCalendarIds,
  load: loadCalendars,
  colorFor: resolveColor,
  defaultCalendarId,
} = useCalendars()
const { revision: quickAddRevision } = useQuickAdd()
let eventRequest = 0

let nowTimer: ReturnType<typeof setInterval> | undefined
let zoomFrame: number | undefined
let dragStartX = 0
let dragStartY = 0
let dragStartDuration = 0
let dragStartTime = 0

const tickIntervals = [
  5 * 60 * 1000,
  15 * 60 * 1000,
  30 * 60 * 1000,
  60 * 60 * 1000,
  2 * 60 * 60 * 1000,
  3 * 60 * 60 * 1000,
  6 * 60 * 60 * 1000,
  12 * 60 * 60 * 1000,
  24 * 60 * 60 * 1000,
  2 * 24 * 60 * 60 * 1000,
  7 * 24 * 60 * 60 * 1000,
  14 * 24 * 60 * 60 * 1000,
  30 * 24 * 60 * 60 * 1000,
  60 * 24 * 60 * 60 * 1000,
  365 * 24 * 60 * 60 * 1000,
  2 * 365 * 24 * 60 * 60 * 1000,
  5 * 365 * 24 * 60 * 60 * 1000,
  10 * 365 * 24 * 60 * 60 * 1000,
]

// Detaljnivåer, prickar och dolda förekomster finns inte längre. Varje
// händelse visas som en rad med ikon, titel och starttid, och rader som
// överlappar i tid packas i kolumner. Se utils/calendarLayout.ts.

// Hur snabbt tiden komprimeras avgör hur snabbt innehållet får försvinna.
// Vid en långsam scroll hinner en kort animation vara behaglig, vid en snabb
// scroll måste den hinna med - annars ligger element kvar och visar
// information som inte längre stämmer.
const zoomSpeed = ref(0)
let lastDuration = visibleDuration.value
let lastDurationAt = performance.now()
let speedTimer: ReturnType<typeof setTimeout> | undefined

const BASE_DURATION_MS = 200
const MIN_DURATION_MS = 60
const MAX_DURATION_MS = 240

const detailDuration = computed(() => {
  const speed = zoomSpeed.value

  if (speed <= 0) return BASE_DURATION_MS

  return Math.round(
    Math.min(
      MAX_DURATION_MS,
      Math.max(MIN_DURATION_MS, BASE_DURATION_MS / (1 + speed / 2)),
    ),
  )
})

watch(visibleDuration, (duration) => {
  const now = performance.now()
  const elapsed = Math.max(1, now - lastDurationAt) / 1000

  // Logaritmen gör att en dubblering räknas lika stor oavsett om den
  // sker från timmar till dagar eller från år till årtionden.
  zoomSpeed.value = Math.abs(Math.log(duration) - Math.log(lastDuration)) / elapsed
  lastDuration = duration
  lastDurationAt = now

  if (speedTimer) clearTimeout(speedTimer)
  speedTimer = setTimeout(() => {
    zoomSpeed.value = 0
  }, 160)
})

const tickInterval = computed(() => {
  const targetTickCount = 10
  const target = visibleDuration.value / targetTickCount
  const interval = tickIntervals.find((candidate) => candidate >= target)
  if (interval) return interval

  // Fortsätt skala utan ett sista fast zoomsteg. För mycket stora intervall
  // används hela år som tick-intervall för att hålla renderingstakten stabil.
  const year = 365 * 24 * 60 * 60 * 1000
  return Math.ceil(target / year) * year
})

const formatTime = (timestamp: number) =>
  new Intl.DateTimeFormat('sv-SE', {
    hour: '2-digit',
    minute: '2-digit',
  }).format(timestamp)

// Veckodagen står till vänster om datumet på dagskalan, så att det går att
// se direkt vilken dag en rad gäller. Samma gäller på timskalan, där dagsnamnet
// står på den rad som inleder det nya dygnet.
const formatWeekday = (timestamp: number) =>
  new Intl.DateTimeFormat('sv-SE', { weekday: 'short' }).format(timestamp)

const formatTick = (timestamp: number, scale: string) => {
  if (scale === 'hour') return formatTime(timestamp)
  if (scale === 'day') {
    return new Intl.DateTimeFormat('sv-SE', { day: 'numeric', month: 'short' }).format(timestamp)
  }
  if (scale === 'month') {
    return new Intl.DateTimeFormat('sv-SE', { month: 'short', year: 'numeric' }).format(timestamp)
  }
  return new Intl.DateTimeFormat('sv-SE', { year: 'numeric' }).format(timestamp)
}

const formatDuration = (duration: number) => {
  const hours = duration / (60 * 60 * 1000)
  if (hours < 24) return `${hours < 10 ? hours.toFixed(1) : Math.round(hours)} h`

  const days = hours / 24
  if (days < 365) return `${days.toFixed(days % 1 === 0 ? 0 : 1)} dygn`

  const years = days / 365
  return `${years.toFixed(years < 10 ? 1 : 0)} år`
}

const timeToPosition = (timestamp: number) =>
  ((timestamp - startTime.value) / visibleDuration.value) * 100

const ticks = computed(() => {
  const interval = tickInterval.value
  const firstTick = Math.floor(startTime.value / interval) * interval
  const result: Array<{
    timestamp: number
    position: number
    major: boolean
    /** Sant för raden som inleder ett nytt dygn på timskalan. */
    dayStart: boolean
  }> = []

  const scale = interval < 24 * 60 * 60 * 1000
    ? 'hour'
    : interval < 45 * 24 * 60 * 60 * 1000
      ? 'day'
      : interval < 365 * 24 * 60 * 60 * 1000
        ? 'month'
        : 'year'

  for (let timestamp = firstTick; timestamp <= endTime.value; timestamp += interval) {
    const position = timeToPosition(timestamp)
    if (position < -2 || position > 102) continue

    const date = new Date(timestamp)
    result.push({
      timestamp,
      position,
      major: interval >= 24 * 60 * 60 * 1000 || date.getHours() === 0,
      dayStart: scale === 'hour' && date.getHours() === 0,
      scale,
    })
  }

  // Intervallen som två eller tre timmar kan missa lokal midnatt eftersom
  // deras millisekundsraster är förankrat i UTC. Lägg därför alltid till en
  // tydlig linje vid varje nytt dygn när timskalan visas.
  if (scale === 'hour') {
    const dayCursor = new Date(startTime.value)
    dayCursor.setHours(0, 0, 0, 0)

    while (dayCursor.getTime() <= endTime.value) {
      const timestamp = dayCursor.getTime()
      if (
        !result.some((tick) => tick.timestamp === timestamp) &&
        timeToPosition(timestamp) >= -2 &&
        timeToPosition(timestamp) <= 102
      ) {
        result.push({
          timestamp,
          position: timeToPosition(timestamp),
          major: true,
          dayStart: true,
          scale,
        })
      }
      dayCursor.setDate(dayCursor.getDate() + 1)
    }
  }

  return result.sort((left, right) => left.timestamp - right.timestamp)
})

const YEAR_MS = 365 * 24 * 60 * 60 * 1000

// Egen ikon per månad gör att månadsgränserna går att känna igen direkt,
// utan att behöva läsa etiketten.
const MONTH_ICONS = [
  '❄', // januari
  '♥', // februari
  '✿', // mars
  '☀', // april
  '✿', // maj
  '☀', // juni
  '☀', // juli
  '✿', // augusti
  '🍂', // september
  '🍂', // oktober
  '🌧', // november
  '❄', // december
]

const monthIcon = (timestamp: number) => MONTH_ICONS[new Date(timestamp).getMonth()]

// Under ett års omfång visas månadsgränserna tydligt, annars blir det
// svårt att orientera sig i tidslinjen.
const monthMarkers = computed(() => {
  if (visibleDuration.value >= YEAR_MS) return []

  const result: Array<{ timestamp: number; position: number; label: string; icon: string }> = []
  const cursor = new Date(startTime.value)
  cursor.setDate(1)
  cursor.setHours(0, 0, 0, 0)

  while (cursor.getTime() <= endTime.value) {
    const timestamp = cursor.getTime()
    const position = timeToPosition(timestamp)
    if (position >= -2 && position <= 102) {
      result.push({
        timestamp,
        position,
        label: new Intl.DateTimeFormat('sv-SE', { month: 'long', year: 'numeric' }).format(timestamp),
        icon: monthIcon(timestamp),
      })
    }
    cursor.setMonth(cursor.getMonth() + 1)
  }

  return result
})

// En tick som ligger för nära en månadsgräns skulle överlappas av
// månadsmarkörens etikett. Då döljs tickens egen etikett i stället.
const MONTH_LABEL_COLLISION = 1.9

const calendarTicks = computed(() => {
  const markers = monthMarkers.value
  if (markers.length === 0) return ticks.value

  return ticks.value.map((tick) => ({
    ...tick,
    hidden: markers.some((marker) => Math.abs(marker.position - tick.position) < MONTH_LABEL_COLLISION),
  }))
})

const isNowVisible = computed(() => now.value >= startTime.value && now.value <= endTime.value)
const nowPosition = computed(() => timeToPosition(now.value))

// Hoppar till en händelse som inte fick plats i tidslinjen. Vyn zoomas in
// till dygnet och händelsen öppnas för redigering.
const focusEvent = (event: CalendarEvent) => {
  openClipped.value = false
  startTime.value = Date.parse(event.start_at) - 60 * 60 * 1000
  zoomTo(24 * 60 * 60 * 1000)
  openEditEvent(event)
}

const loadEvents = async () => {
  const request = ++eventRequest
  try {
    const loaded = await loadAllCalendarEvents()
    if (request === eventRequest) {
      allEvents.value = loaded
    }
  } catch (error) {
    // Kalendern ska fortfarande kunna användas när API:t eller eventtabellen
    // inte är tillgänglig ännu.
    console.warn('Kunde inte läsa kalenderhändelser:', error)
  }
}

const updateNow = () => {
  now.value = Date.now()
}

const animateZoom = () => {
  const current = visibleDuration.value
  const target = targetDuration.value
  const next = Math.exp(Math.log(current) + (Math.log(target) - Math.log(current)) * 0.18)

  if (Math.abs(Math.log(target) - Math.log(next)) < 0.0005) {
    visibleDuration.value = target
    zoomFrame = undefined
    return
  }

  visibleDuration.value = next
  zoomFrame = requestAnimationFrame(animateZoom)
}

const zoomTo = (duration: number) => {
  targetDuration.value = Math.max(
    MIN_DURATION,
    Math.min(MAX_DATE_DURATION, duration),
  )
  if (zoomFrame === undefined) zoomFrame = requestAnimationFrame(animateZoom)
}

const zoomBy = (distance: number) => {
  zoomTo(targetDuration.value * Math.exp(distance / ZOOM_SENSITIVITY))
}

const panBy = (distance: number) => {
  const height = viewport.value?.clientHeight || 600
  startTime.value += distance * visibleDuration.value / height
}

const panTo = (distance: number, originStart: number, originDuration: number) => {
  const height = viewport.value?.clientHeight || 600
  startTime.value = originStart + distance * originDuration / height
}

// Vyns hörd i pixlar. Den styr hur mycket utrymme en rad får, så den måste
// hållas uppdaterad när fönstret ändras.
const viewportHeight = ref(0)
let viewportObserver: ResizeObserver | undefined

// Händelserna packas i rader. Radens höjd följer händelsens längd, men en
// rad får aldrig bli lägre än den lägsta höjd som går att läsa - annars
// skulle en kort händelse försvinna när tiden komprimeras.
//
// Alla rader ligger i en enda kolumn. En rad som krockar med en tidigare rad
// skjuts ned under den, så att ingen rad döljer en annan.
//
// När kolumnen inte ryms döljs återkommande serier, först den som har flest
// förekomster. Se utils/calendarVisibility.ts.
const hiddenSeries = ref<string[]>([])

const { settings } = useCalendarSettings()

// Planen räknas om när vyn, serierna eller inställningarna ändras. Den
// läser den tidigare döljningen, så att samma serie förblir dold när antalet
// förekomster ändras - annars fladdrar de fram och tillbaka när man scrollar.
watch(
  [
    () => startTime.value,
    () => endTime.value,
    () => viewportHeight.value,
    () => settings.value.maxInstances,
    () => settings.value.hideRecurring,
    recurringSeries,
  ],
  () => {
    const plan = planVisibility({
      events: coloredEvents.value.map((event) => ({
        start: Date.parse(event.start_at),
        end: Date.parse(event.end_at),
        seriesKey: recurrenceSeriesKey(event),
      })),
      series: recurringSeries.value,
      maxInstances: settings.value.maxInstances,
      hidingEnabled: settings.value.hideRecurring,
      hiddenBefore: hiddenSeries.value,
      from: startTime.value,
      to: endTime.value,
      height: viewportHeight.value || 600,
    })

    const next = [...plan.hidden].sort()
    const current = [...hiddenSeries.value].sort()

    if (next.join('|') !== current.join('|')) hiddenSeries.value = next
  },
  { immediate: true },
)

const eventRows = computed(() => {
  const height = viewportHeight.value || 600
  const hidden = new Set(hiddenSeries.value)

  // Dolda serier tas bort ur tidslinjen helt. Deras förekomster listas i
  // sammanfattningen ovanför tidslinjen i stället.
  const events = coloredEvents.value
    .filter((event) => {
      const key = recurrenceSeriesKey(event)
      return key === null || !hidden.has(key)
    })
    // Sorteras efter tid, så att en rad som ligger längre ned i vyn ritas
    // ovanpå en rad som börjar tidigare. Då syns titeln på båda.
    .sort(
      (left, right) =>
        Date.parse(left.start_at) - Date.parse(right.start_at),
    )

  const rows = layoutEventRows(
    events.map((event) => ({
      start: Date.parse(event.start_at),
      end: Date.parse(event.end_at),
    })),
    {
      from: startTime.value,
      to: endTime.value,
      height,
      minRow: MIN_ROW_PX,
    },
  ).map((row, index) => ({ ...row, event: events[index]! }))

  return {
    rows: rows.filter((row) => !row.clipped),
    // Enstaka händelser som inte hinner med i kolumnen. De får inte
    // försvinna tyst, så de visas som ett antal i en knapp.
    clipped: rows.filter((row) => row.clipped),
  }
})

// De händelser som inte ryms i kolumnen. De visas i en panel när knappen
// öppnas.
const openClipped = ref(false)

const clippedEvents = computed(() => eventRows.value.clipped.map((row) => row.event))

// --- Schemaläggning av mål genom att släppa dem i kalendern ---

// Tiden som drop-markören visar just nu, eller null om inget dras över.
const dropPreviewTime = ref<number | null>(null)

// Ett mål som har släppts och väntar på att en tid väljs.
const pendingGoalDrop = ref<{
  goal: GoalDragPayload
  startAt: number
  left: number
  top: number
} | null>(null)

/**
 * Tiden som en pekarposition i vyn motsvarar.
 *
 * Pekaren räknas relativt vyn, så att tiden följer tidsaxeln oavsett
 * var i sidan vyn ligger.
 */
const timeAtClientY = (clientY: number): number => {
  const rect = viewport.value?.getBoundingClientRect()
  if (!rect || rect.height <= 0) return startTime.value

  return timeAtOffset({
    offsetY: clientY - rect.top,
    height: rect.height,
    startTime: startTime.value,
    visibleDuration: visibleDuration.value,
  })
}

/**
 * Tiden som pekaren pekar på, avrundad till en rimlig upplösning.
 */
const snappedTimeAtClientY = (clientY: number): number =>
  snapTimestamp(timeAtClientY(clientY), snapStepMs(visibleDuration.value))

/**
 * Handlar det om att ett mål dras över kalendern?
 *
 * Kalendern ska inte reagera på annat som dras i sidan.
 */
const handleDragOver = (event: DragEvent) => {
  if (!carriesGoal(event)) return

  // Utan preventDefault släpps ingenting - webbläsaren vägrar då
  // att markera platsen som ett giltig släppmål.
  event.preventDefault()

  if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy'

  dropPreviewTime.value = snappedTimeAtClientY(event.clientY)
}

const handleDragLeave = (event: DragEvent) => {
  // Markören ska bara försvinna när pekaren lämnar vyn helt. När den
  // går över ett barnelement brinner "dragleave", utan att pekaren
  // faktiskt har lämnat vyn.
  const related = event.relatedTarget as Node | null
  if (related && viewport.value?.contains(related)) return

  dropPreviewTime.value = null
}

const handleDrop = (event: DragEvent) => {
  if (!carriesGoal(event)) return

  event.preventDefault()
  dropPreviewTime.value = null

  const goal = parseGoalDrag(event.dataTransfer?.getData(GOAL_DRAG_MIME))
  if (!goal) return

  const shellRect = shell.value?.getBoundingClientRect()

  // Menyn ska rymmas inom kalendern, så den placeras inte utanför
  // kanterna.
  const menuWidth = 220
  const menuHeight = 320
  const left = shellRect
    ? Math.min(
        Math.max(event.clientX - shellRect.left - menuWidth / 2, 0),
        Math.max(shellRect.width - menuWidth, 0),
      )
    : 0
  const top = shellRect
    ? Math.min(
        Math.max(event.clientY - shellRect.top, 0),
        Math.max(shellRect.height - menuHeight, 0),
      )
    : 0

  pendingGoalDrop.value = {
    goal,
    startAt: snappedTimeAtClientY(event.clientY),
    left,
    top,
  }
}

const cancelGoalDrop = () => {
  pendingGoalDrop.value = null
}

// Escape stänger menyn. Menyn kan visas på alla sidor - målet dras
// från en lista, medan kalendern ligger i sidopanelen - så lyssnaren
// sitter på fönstret.
const handleDropKeydown = (event: KeyboardEvent) => {
  if (event.key !== 'Escape') return
  if (!pendingGoalDrop.value) return

  event.preventDefault()
  cancelGoalDrop()
}

/**
 * Skapar händelsen för ett släppt mål.
 *
 * Händelsen får målets titel, ikon och koppling. Tiden är den valda
 * längden, räknad från den tidpunkten målet släppetes på.
 */
const scheduleGoal = async (minutes: number) => {
  const drop = pendingGoalDrop.value
  if (!drop || minutes <= 0) return

  savingEvent.value = true
  errorMessage.value = ''

  try {
    await createCalendarEvent({
      title: drop.goal.title,
      icon: drop.goal.icon || '',
      goal_id: drop.goal.id,
      start_at: new Date(drop.startAt).toISOString(),
      end_at: new Date(drop.startAt + minutes * 60 * 1000).toISOString(),
      time_zone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      ...(defaultCalendarId.value ? { calendar_id: defaultCalendarId.value } : {}),
    })
    cancelGoalDrop()
    await loadEvents()
  } catch (error) {
    console.error('Kunde inte schemalägga målet:', error)
    errorMessage.value = 'Kunde inte schemalägga målet. Försök igen.'
  } finally {
    savingEvent.value = false
  }
}

// --- Flytta en händelse genom att dra den ---

// En händelse som dras just nu.
const eventDrag = ref<{
  key: string
  event: CalendarEvent
  startY: number
  deltaMs: number
  moved: boolean
} | null>(null)

// Ett klick strax efter ett drag ska inte öppna formuläret - raden har
// redan flyttats. Tiden sparas, så att ett klick som kommer en stund
// senare ändå öppnar händelsen.
let suppressClickUntil = 0

/**
 * Startar en dragning av en händelse.
 *
 * Inget händer förrän pekaren rör sig - ett klick ska fortfarande
 * öppna händelsen i formuläret.
 */
const startEventDrag = (event: PointerEvent, calendarEvent: CalendarEvent) => {
  if (event.button !== 0) return

  eventDrag.value = {
    key: calendarEvent.occurrence_id || `${calendarEvent.id}-${calendarEvent.start_at}`,
    event: calendarEvent,
    startY: event.clientY,
    deltaMs: 0,
    moved: false,
  }
}

/**
 * Flyttar den händelse som dras.
 *
 * Förflyttningen räknas i pixlar och omvandlas till tid, så att
 * händelsen flyttas lika långt i tiden som pekaren rör sig.
 */
const moveEventDrag = (event: PointerEvent) => {
  const drag = eventDrag.value
  if (!drag) return

  const deltaY = event.clientY - drag.startY

  if (!drag.moved && Math.abs(deltaY) < 4) return

  drag.moved = true
  drag.deltaMs = snapTimestamp(
    deltaY * millisecondsPerPixel(visibleDuration.value, viewport.value?.clientHeight || 600),
    snapStepMs(visibleDuration.value),
  )
}

/**
 * Slutför dragningen, om händelsen faktiskt flyttades.
 */
const endEventDrag = async () => {
  const drag = eventDrag.value
  eventDrag.value = null

  if (!drag || !drag.moved) return

  // Klicket som följer efteråt ska inte öppna formuläret - händelsen
  // har redan flyttats.
  suppressClickUntil = Date.now() + 250

  const start = Date.parse(drag.event.start_at)
  const end = Date.parse(drag.event.end_at)
  const moved = shiftEventTimes({ start, end, deltaMs: drag.deltaMs })

  if (moved.start === start) return

  try {
    await persistEventMove(drag.event, moved.start, moved.end)
    await loadEvents()
  } catch (error) {
    console.error('Kunde inte flytta händelsen:', error)
    errorMessage.value = 'Kunde inte flytta händelsen. Försök igen.'
  }
}

/**
 * Sparar en flyttad händelse.
 *
 * En förekomst i en serie flyttas med en override-rad, så att bara
 * just den förekomsten flyttas. En händelse med egen rad, eller en
 * vanlig händelse, uppdateras direkt.
 */
const persistEventMove = async (
  calendarEvent: CalendarEvent,
  start: number,
  end: number,
) => {
  const startAt = new Date(start).toISOString()
  const endAt = new Date(end).toISOString()

  if (isSeriesOccurrence(calendarEvent) && !calendarEvent.parent_event_id) {
    // Förekomsten har ingen egen rad - servern lägger till en.
    await createCalendarOccurrenceOverride(parentEventId(calendarEvent), {
      original_start_at: calendarEvent.original_start_at || calendarEvent.start_at,
      start_at: startAt,
      end_at: endAt,
      title: calendarEvent.title,
      icon: calendarEvent.icon || '',
      time_zone: calendarEvent.time_zone,
    })
    return
  }

  await updateCalendarEvent(calendarEvent.id, {
    start_at: startAt,
    end_at: endAt,
  })
}

const handleEventClick = (calendarEvent: CalendarEvent) => {
  if (Date.now() < suppressClickUntil) {
    suppressClickUntil = 0
    return
  }

  openEditEvent(calendarEvent)
}

/**
 * Sant om just den händelsen dras just nu.
 */
const isDraggedEvent = (calendarEvent: CalendarEvent): boolean => {
  const drag = eventDrag.value
  if (!drag) return false

  return (
    drag.key === (calendarEvent.occurrence_id || `${calendarEvent.id}-${calendarEvent.start_at}`)
  )
}

/**
 * Förflyttningen av en händelse som dras, i pixlar.
 *
 * Raden flyttas med pekaren, så att det går att se vart den hamnar.
 */
const eventDragOffset = (calendarEvent: CalendarEvent): string | undefined => {
  const drag = eventDrag.value
  if (!drag || !drag.moved || !isDraggedEvent(calendarEvent)) return undefined

  const pixels = drag.deltaMs / millisecondsPerPixel(
    visibleDuration.value,
    viewport.value?.clientHeight || 600,
  )

  return `translateY(${pixels}px)`
}

const handleWheel = (event: WheelEvent) => {
  event.preventDefault()
  // Trackpad och mushjul skickar olika delta-enheter. Normalisera dem först.
  const unit = event.deltaMode === 1
    ? 16
    : event.deltaMode === 2
      ? (viewport.value?.clientHeight || 600)
      : 1

  // Lodrät skrollning zoomar tidsfönstret. Sidled flyttar det framåt eller
  // bakåt.
  zoomBy(event.deltaY * unit)
  panBy(event.deltaX * unit)
}

const handlePointerDown = (event: PointerEvent) => {
  if (event.button !== 0) return

  const element = viewport.value
  if (!element) return

  dragStartX = event.clientX
  dragStartY = event.clientY
  dragStartDuration = visibleDuration.value
  dragStartTime = startTime.value
  isDragging.value = true
  dragMode.value = 'pending'
  element.setPointerCapture(event.pointerId)
}

const handlePointerMove = (event: PointerEvent) => {
  if (!isDragging.value || !viewport.value) return

  const deltaX = event.clientX - dragStartX
  const deltaY = event.clientY - dragStartY

  if (dragMode.value === 'pending' && Math.max(Math.abs(deltaX), Math.abs(deltaY)) > 5) {
    dragMode.value = 'zoom'
  }

  if (dragMode.value === 'zoom') {
    // Samma konvention som scrollningen: lodrätt drag zoomar, vågrätt drag
    // flyttar tidsfönstret.
    zoomTo(dragStartDuration * Math.exp(deltaY / ZOOM_SENSITIVITY))
    panTo(deltaX, dragStartTime, dragStartDuration)
  }
}

const handlePointerUp = (event: PointerEvent) => {
  if (viewport.value?.hasPointerCapture(event.pointerId)) {
    viewport.value.releasePointerCapture(event.pointerId)
  }
  isDragging.value = false
  dragMode.value = null
}

const toLocalInput = (date: Date) => {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000)
  return local.toISOString().slice(0, 16)
}

/**
 * Färgen som visas i formuläret. En händelse har alltid kalenderns färg,
 * så det finns inget eget färgval att göra.
 */

const openCreateEvent = () => {
  const start = new Date(Date.now() + 60 * 60 * 1000)
  const end = new Date(start.getTime() + 60 * 60 * 1000)
  draftEvent.title = ''
  draftEvent.description = ''
  draftEvent.goalId = ''
  draftEvent.recurrenceRule = ''
  draftEvent.icon = ''
  // Standardkalendern är förvald, så att det bara är att skriva. Händelsen
  // ärver kalenderns färg.
  draftEvent.calendarId = defaultCalendarId.value ? String(defaultCalendarId.value) : ''
  draftEvent.placeIds = []
  draftEvent.startAt = toLocalInput(start)
  draftEvent.endAt = toLocalInput(end)
  showCreateEvent.value = true
}

const openEditEvent = (event: CalendarEvent) => {
  editingEvent.value = event
  // En förekomst i en serie kan ändras för sig, tillsammans med de
  // följande eller för hela serien. En vanlig händelse har bara en rad.
  editScope.value = isSeriesOccurrence(event) ? 'occurrence' : 'series'
  draftEvent.title = event.title
  draftEvent.description = event.description || ''
  draftEvent.goalId = event.goal_id ? String(event.goal_id) : ''
  draftEvent.recurrenceRule = event.recurrence_rule || ''
  draftEvent.icon = event.icon || ''
  draftEvent.calendarId = event.calendar_id ? String(event.calendar_id) : ''
  draftEvent.placeIds = (event.places ?? []).map((place) => place.id)
  draftEvent.startAt = toLocalInput(new Date(event.start_at))
  draftEvent.endAt = toLocalInput(new Date(event.end_at))
}

const closeEditEvent = () => {
  editingEvent.value = null
}

/**
 * Sant om händelsen som redigeras är en förekomst i en serie.
 *
 * Förekomsten bär inte alltid seriens upprepningsregel - en förekomst med
 * egen rad (en override) har bara seriens id. Därför avgörs det av
 * förekomstens identitet, inte av regeln.
 */
const editingIsSeriesOccurrence = computed(() =>
  editingEvent.value ? isSeriesOccurrence(editingEvent.value) : false,
)

/**
 * Seriens masterrad, som har upprepningsregeln.
 *
 * Masterraden finns bland de hämtade händelserna, så regeln kan slås upp
 * där även när den synda förekomsten inte bär med sig den.
 */
const seriesMaster = (id: number) =>
  allEvents.value.find((event) => event.id === id)

const saveEditedEvent = async () => {
  const event = editingEvent.value
  if (!event || !draftEvent.title.trim()) return

  savingEvent.value = true
  const payload = {
    title: draftEvent.title.trim(),
    description: draftEvent.description.trim(),
    // En tom sträng tömmer ikonen, så den skickas alltid med.
    icon: draftEvent.icon,
    // 0 betyder "ingen kalender".
    calendar_id: draftEvent.calendarId ? Number(draftEvent.calendarId) : 0,
    start_at: new Date(draftEvent.startAt).toISOString(),
    end_at: new Date(draftEvent.endAt).toISOString(),
    time_zone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    ...(draftEvent.goalId ? { goal_id: Number(draftEvent.goalId) } : {}),
  }

  try {
    if (editScope.value === 'occurrence' && isSeriesOccurrence(event)) {
      // En förekomst med egen rad ändras direkt. En förekomst utan egen
      // rad får en override, som servern lägger till.
      if (event.parent_event_id) {
        await updateCalendarEvent(event.id, payload)
      } else {
        await createCalendarOccurrenceOverride(parentEventId(event), {
          original_start_at: event.original_start_at || event.start_at,
          ...payload,
        })
      }
    } else if (editScope.value === 'following' && isSeriesOccurrence(event)) {
      const master = seriesMaster(parentEventId(event))
      await splitCalendarSeries(parentEventId(event), {
        original_start_at: event.original_start_at || event.start_at,
        recurrence_rule:
          draftEvent.recurrenceRule || master?.recurrence_rule || event.recurrence_rule || '',
        ...payload,
      })
    } else {
      await updateCalendarEvent(parentEventId(event), {
        ...payload,
        ...(draftEvent.recurrenceRule ? { recurrence_rule: draftEvent.recurrenceRule } : {}),
      })
    }
    // Platserna sätts separat från händelsen - de ändras inte när
    // förekomsten eller serien sparas.
    await setEventPlaces(parentEventId(event), draftEvent.placeIds)
    closeEditEvent()
    await loadEvents()
  } finally {
    savingEvent.value = false
  }
}

/**
 * Tar bort en händelse.
 *
 * Omfattningen beror på vilken rad som egentligen ska försvinna: en vanlig
 * händelse tas bort som den är, en förekomst i en serie kan tas bort för
 * sig, tillsammans med de följande eller tillsammans med hela serien.
 */
const removeEvent = async (
  event: CalendarEvent,
  scope: 'occurrence' | 'series' | 'following',
) => {
  if (!isSeriesOccurrence(event)) {
    await deleteCalendarEvent(event.id)
    return
  }

  const parentId = parentEventId(event)
  const occurrenceStart = Date.parse(event.original_start_at || event.start_at)

  if (scope === 'series') {
    await deleteCalendarEvent(parentId)
    return
  }

  if (scope === 'following') {
    // Serien slutar strax före förekomsten, så att den och de följande
    // försvinner ur serien.
    const master = seriesMaster(parentId)
    if (master?.recurrence_rule) {
      await updateCalendarEvent(parentId, {
        recurrence_rule: truncateRecurrenceRule(
          master.recurrence_rule,
          new Date(occurrenceStart - 1000),
        ),
      })
    }

    // Rader för förekomster som nu ligger utanför serien tas bort.
    // Annars skulle de ligga kvar som egna händelser i vyn.
    for (const override of allEvents.value) {
      if (override.parent_event_id !== parentId) continue
      if (!override.original_start_at) continue
      if (Date.parse(override.original_start_at) < occurrenceStart) continue

      await deleteCalendarEvent(override.id)
    }
    return
  }

  // Bara den valda förekomsten: den stängs av.
  if (event.parent_event_id) {
    // Förekomsten har en egen rad - raden stängs av med sin status.
    await updateCalendarEvent(event.id, { status: 'cancelled' })
    return
  }

  await createCalendarOccurrenceOverride(parentId, {
    original_start_at: event.original_start_at || event.start_at,
    start_at: event.start_at,
    end_at: event.end_at,
    status: 'cancelled',
  })
}

const deleteEvent = async () => {
  const event = editingEvent.value
  if (!event) return

  const scope = editScope.value
  const label = !isSeriesOccurrence(event)
    ? 'händelsen'
    : scope === 'series'
      ? 'hela serien'
      : scope === 'following'
        ? 'denna och de följande'
        : 'denna förekomst'

  if (!window.confirm(`Ta bort ${label}? Det går inte att ångra.`)) return

  deletingEvent.value = true
  errorMessage.value = ''
  try {
    await removeEvent(event, scope)
    closeEditEvent()
    await loadEvents()
  } catch (error) {
    console.error('Kunde inte ta bort händelsen:', error)
    errorMessage.value = 'Kunde inte ta bort händelsen. Försök igen.'
  } finally {
    deletingEvent.value = false
  }
}

/**
 * Duplicerar en händelse.
 *
 * Kopian får samma innehåll och samma tid, och öppnas direkt i formuläret
 * så att tiden kan justeras med en gång. En förekomst kopplas loss från
 * sin serie - kopian blir en egen händelse.
 */
const duplicateEvent = async () => {
  const event = editingEvent.value
  if (!event) return

  duplicatingEvent.value = true
  errorMessage.value = ''
  try {
    const created = await createCalendarEvent(duplicateEventPayload(event))
    closeEditEvent()
    await loadEvents()
    openEditEvent(created)
  } catch (error) {
    console.error('Kunde inte duplicera händelsen:', error)
    errorMessage.value = 'Kunde inte duplicera händelsen. Försök igen.'
  } finally {
    duplicatingEvent.value = false
  }
}

const saveEvent = async () => {
  if (!draftEvent.title.trim() || !draftEvent.startAt || !draftEvent.endAt) return

  savingEvent.value = true
  try {
    const created = await createCalendarEvent({
      title: draftEvent.title.trim(),
      description: draftEvent.description.trim(),
      icon: draftEvent.icon,
      ...(draftEvent.calendarId ? { calendar_id: Number(draftEvent.calendarId) } : {}),
      start_at: new Date(draftEvent.startAt).toISOString(),
      end_at: new Date(draftEvent.endAt).toISOString(),
      time_zone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      ...(draftEvent.goalId ? { goal_id: Number(draftEvent.goalId) } : {}),
      ...(draftEvent.recurrenceRule ? { recurrence_rule: draftEvent.recurrenceRule } : {}),
    })
    await setEventPlaces(created.id, draftEvent.placeIds)
    showCreateEvent.value = false
    await loadEvents()
  } finally {
    savingEvent.value = false
  }
}

// "Hoppa till nu" ska visa de närmaste 24 timmarna. Därför nollställs även
// zoomnivån - annars kan vyn hamna i ett mycket stort eller litet
// tidsfönster som inte matchar det användaren förväntar sig efter "t".
const resetToNow = () => {
  startTime.value = Date.now()
  now.value = startTime.value
  zoomTo(DEFAULT_DURATION)
}

const resetCalendarView = () => {
  resetToNow()
}

const route = useRoute()

const handleCalendarKeydown = (event: KeyboardEvent) => {
  // Kalendern ligger kvar i layouten även när användaren arbetar med mål.
  // Dess globala tangentlyssnare får därför bara styra tidslinjen på själva
  // kalendersidan; annars blir j/k/h/l dubbeltolkade av mål-navigeringen.
  if (route.path !== '/calendar') return

  // Escape stänger en öppen händelsedialog. Det hanteras före kollen
  // nedan, eftersom fokus ligger i ett fält i dialogen när den öppnas -
  // annars skulle tangenten aldrig nå fram till kalendern.
  if (event.key === 'Escape') {
    if (editingEvent.value) {
      event.preventDefault()
      closeEditEvent()
      return
    }
    if (showCreateEvent.value) {
      event.preventDefault()
      showCreateEvent.value = false
      return
    }
  }

  const target = event.target as HTMLElement | null
  if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT' || target.isContentEditable)) return

  if (event.key === 'n') {
    // Snabbt läggande till hanteras globalt i layouten, så att det finns
    // tillgängligt från alla sidor.
    return
  }
  if (event.key === 't') {
    event.preventDefault()
    resetToNow()
    return
  }
  if (event.key === 'r') {
    event.preventDefault()
    resetCalendarView()
    return
  }
  if (event.key === 'j' || event.key === 'ArrowDown') {
    event.preventDefault()
    // j/k följer den lodräta scrollningen: ner zoomar ut, upp zoomar in.
    zoomBy(260)
    return
  }
  if (event.key === 'k' || event.key === 'ArrowUp') {
    event.preventDefault()
    zoomBy(-260)
    return
  }
  if (event.key === 'h' || event.key === 'l') {
    event.preventDefault()
    // h/l följer den horisontella scrollningen: vänster flyttar tiden bakåt,
    // höger framåt.
    if (event.key === 'h') panBy(-150)
    else panBy(150)
    return
  }
  if (event.key === '-' || event.key === '+' || event.key === '=') {
    event.preventDefault()
    zoomBy(event.key === '-' ? 260 : -260)
  }
}

onMounted(async () => {
  nowTimer = setInterval(updateNow, 30_000)
  loadEvents()
  loadCalendars()
  try {
    goals.value = await loadAllGoals()
    // Håll den delade grafen aktuell så att färdigmarkeringar på andra
    // målsidor omedelbart kan dölja kopplade kalenderhändelser.
    await goalsStore.loadGraph(false)
  } catch (error) {
    console.warn('Kunde inte läsa mål för kalenderhändelser:', error)
  }
  try {
    await fetchPlaces()
  } catch (error) {
    console.warn('Kunde inte läsa platser för kalenderhändelser:', error)
  }
  window.addEventListener('keydown', handleCalendarKeydown)
  window.addEventListener('keydown', handleDropKeydown)
  // Dragningen av en händelse följer pekaren även när den lämnar raden,
  // så lyssnarna sitter på fönstret.
  window.addEventListener('pointermove', moveEventDrag)
  window.addEventListener('pointerup', endEventDrag)
  window.addEventListener('pointercancel', endEventDrag)

  // Tidsfönstrets höjd styr hur mycket utrymme prickarna får, så den måste
  // hållas uppdaterad när fönstret ändras.
  if (viewport.value) {
    viewportHeight.value = viewport.value.clientHeight
    viewportObserver = new ResizeObserver(() => {
      viewportHeight.value = viewport.value?.clientHeight ?? 0
    })
    viewportObserver.observe(viewport.value)
  }
})

// Något som skapats via det globala snabbfältet ska synas direkt.
watch(quickAddRevision, loadEvents)

onUnmounted(() => {
  if (nowTimer) clearInterval(nowTimer)
  if (zoomFrame !== undefined) cancelAnimationFrame(zoomFrame)
  if (speedTimer) clearTimeout(speedTimer)
  viewportObserver?.disconnect()
  window.removeEventListener('keydown', handleCalendarKeydown)
  window.removeEventListener('keydown', handleDropKeydown)
  window.removeEventListener('pointermove', moveEventDrag)
  window.removeEventListener('pointerup', endEventDrag)
  window.removeEventListener('pointercancel', endEventDrag)
})
</script>

<template>
  <section ref="shell" class="calendar-shell">
    <!-- Kalendrarna ligger i en lista som öppnas med en knapp, i stället för
         att ligga som etiketter över tidslinjen. -->
    <div class="calendar-toolbar">
      <button type="button" class="toolbar-button" @click="openCreateEvent">
        + Händelse
      </button>
      <button
        type="button"
        class="toolbar-button"
        :aria-expanded="managerOpen"
        @click="managerOpen = !managerOpen"
      >
        Kalendrar
        <span class="toolbar-count">{{ calendars.length }}</span>
      </button>
    </div>

    <Teleport to="body">
      <div
        v-if="editingEvent"
        class="event-modal-backdrop"
        role="presentation"
        @click.self="closeEditEvent"
      >
        <form
          class="event-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="event-modal-title"
          @submit.prevent="saveEditedEvent"
        >
          <header class="event-modal-header">
            <div>
              <span class="event-modal-kicker">KALENDERHÄNDELSE</span>
              <h2 id="event-modal-title">Redigera händelse</h2>
            </div>
            <button
              type="button"
              class="event-modal-close"
              aria-label="Stäng"
              @click="closeEditEvent"
            >
              <Icon name="lucide:x" class="h-5 w-5" />
            </button>
          </header>

          <div class="event-modal-body">
            <label class="event-field event-field-wide">
              <span>Titel</span>
              <input v-model="draftEvent.title" required maxlength="255" aria-label="Titel" autofocus />
            </label>

            <div class="event-field">
              <span>Ikon</span>
              <div class="icon-field">
                <button type="button" class="icon-button" @click="iconPickerOpen = true">
                  <EventIcon :icon="draftEvent.icon" class="icon-preview" />
                  <span>{{ draftEvent.icon ? 'Byt ikon' : 'Lägg till ikon' }}</span>
                </button>
                <EventIconPicker v-model="draftEvent.icon" :open="iconPickerOpen" @update:open="iconPickerOpen = $event" />
              </div>
            </div>

            <label class="event-field">
              <span>Kalender</span>
              <select v-model="draftEvent.calendarId" aria-label="Kalender">
                <option value="">Utan kalender</option>
                <option v-for="calendar in calendars" :key="calendar.id" :value="String(calendar.id)">
                  {{ calendar.name }}
                </option>
              </select>
            </label>

            <label class="event-field event-field-wide">
              <span>Kopplat mål</span>
              <select v-model="draftEvent.goalId" aria-label="Kopplat mål">
                <option value="">Inget kopplat mål</option>
                <option v-for="goal in goals" :key="goal.id" :value="String(goal.id)">
                  {{ goal.title }}
                </option>
              </select>
            </label>

            <label class="event-field">
              <span>Start</span>
              <input v-model="draftEvent.startAt" required type="datetime-local" aria-label="Starttid" />
            </label>
            <label class="event-field">
              <span>Slut</span>
              <input v-model="draftEvent.endAt" required type="datetime-local" aria-label="Sluttid" />
            </label>

            <label class="event-field event-field-wide">
              <span>Beskrivning</span>
              <textarea v-model="draftEvent.description" aria-label="Beskrivning" rows="3" />
            </label>

            <label class="event-field event-field-wide">
              <span>Platser</span>
              <select v-model="draftEvent.placeIds" multiple aria-label="Platser" class="event-places">
                <option v-for="place in allPlaces" :key="place.id" :value="place.id">
                  {{ place.name }}
                </option>
              </select>
            </label>

            <label v-if="editingIsSeriesOccurrence" class="event-field event-field-wide">
              <span>Ändra serie</span>
              <select v-model="editScope" aria-label="Ändra serie">
                <option value="occurrence">Bara denna förekomst</option>
                <option value="following">Denna och följande</option>
                <option value="series">Hela serien</option>
              </select>
            </label>
          </div>

          <footer class="event-modal-footer">
            <button type="button" class="event-action event-action-danger" :disabled="deletingEvent" @click="deleteEvent">
              {{ deletingEvent ? 'Tar bort…' : 'Ta bort' }}
            </button>
            <button type="button" class="event-action" :disabled="duplicatingEvent" @click="duplicateEvent">
              {{ duplicatingEvent ? 'Duplicerar…' : 'Duplicera' }}
            </button>
            <span class="event-modal-spacer" />
            <button type="button" class="event-action" @click="closeEditEvent">Avbryt</button>
            <button type="submit" class="event-action event-action-primary" :disabled="savingEvent">
              {{ savingEvent ? 'Sparar…' : 'Spara ändringar' }}
            </button>
          </footer>
          <p v-if="errorMessage" class="form-error" role="status">{{ errorMessage }}</p>
        </form>
      </div>
    </Teleport>

    <form v-if="showCreateEvent" class="event-form" @submit.prevent="saveEvent">
      <input v-model="draftEvent.title" required maxlength="255" placeholder="Titel" aria-label="Titel" class="field-wide" />
      <div class="icon-field">
        <button type="button" class="icon-button" @click="iconPickerOpen = true">
          <EventIcon :icon="draftEvent.icon" class="icon-preview" />
          <span>{{ draftEvent.icon ? 'Byt ikon' : 'Lägg till ikon' }}</span>
        </button>
        <EventIconPicker v-model="draftEvent.icon" :open="iconPickerOpen" @update:open="iconPickerOpen = $event" />
      </div>
      <select v-model="draftEvent.calendarId" aria-label="Kalender">
        <option value="">Utan kalender</option>
        <option v-for="calendar in calendars" :key="calendar.id" :value="String(calendar.id)">
          {{ calendar.name }}
        </option>
      </select>
      <input v-model="draftEvent.startAt" required type="datetime-local" aria-label="Starttid" />
      <input v-model="draftEvent.endAt" required type="datetime-local" aria-label="Sluttid" />
      <input v-model="draftEvent.description" placeholder="Beskrivning (valfritt)" aria-label="Beskrivning" class="field-wide" />
      <select v-model="draftEvent.goalId" aria-label="Kopplat mål">
        <option value="">Inget kopplat mål</option>
        <option v-for="goal in goals" :key="goal.id" :value="String(goal.id)">
          {{ goal.title }}
        </option>
      </select>
      <select v-model="draftEvent.placeIds" multiple aria-label="Platser" class="field-wide">
        <option v-for="place in allPlaces" :key="place.id" :value="place.id">
          {{ place.name }}
        </option>
      </select>
      <select v-model="draftEvent.recurrenceRule" aria-label="Upprepning">
        <option value="">Upprepa inte</option>
        <option value="FREQ=DAILY">Varje dag</option>
        <option value="FREQ=WEEKLY">Varje vecka</option>
        <option value="FREQ=MONTHLY">Varje månad</option>
      </select>
      <button type="submit" class="now-button primary-button" :disabled="savingEvent">
        {{ savingEvent ? 'Sparar…' : 'Spara' }}
      </button>
      <button type="button" class="now-button" @click="showCreateEvent = false">Avbryt</button>
    </form>

    <!-- De händelser som inte hinner med i kolumnen. De visas som en
         "+N"-knapp i vyn, och panelen listar dem när knappen öppnas. -->
    <div v-if="openClipped && clippedEvents.length > 0" class="overflow-panel">
      <div class="overflow-panel-header">
        <strong>Dolda händelser</strong>
        <span>{{ clippedEvents.length }} st</span>
        <button type="button" class="overflow-panel-close" title="Stäng" @click="openClipped = false">×</button>
      </div>
      <p class="overflow-panel-note">
        De här händelserna ryms inte i kolumnen just nu. Klicka på en händelse
        för att zooma in på den.
      </p>
      <ul class="overflow-panel-list">
        <li v-for="event in clippedEvents" :key="event.occurrence_id || `${event.id}-${event.start_at}`">
          <button type="button" class="overflow-panel-entry" @click="focusEvent(event)">
            <EventIcon v-if="event.icon" :icon="event.icon" class="event-icon" />
            <Icon
              v-if="event.places?.length"
              name="lucide:map-pin"
              class="event-icon"
              :title="`Platser: ${event.places.map((p) => p.name).join(', ')}`"
            />
            <span class="overflow-panel-title">{{ event.title }}</span>
            <span class="overflow-panel-count">{{ formatTime(Date.parse(event.start_at)) }}</span>
          </button>
        </li>
      </ul>
    </div>

    <div
      ref="viewport"
      class="calendar-viewport"
      :class="{ dragging: isDragging, zooming: dragMode === 'zoom' }"
      @wheel="handleWheel"
      @pointerdown="handlePointerDown"
      @pointermove="handlePointerMove"
      @pointerup="handlePointerUp"
      @pointercancel="handlePointerUp"
      @dragover="handleDragOver"
      @dragleave="handleDragLeave"
      @drop="handleDrop"
    >
      <div
        class="calendar-grid"
        :style="{ '--detail-duration': `${detailDuration}ms` }"
        aria-label="Vertikal kalender"
      >
        <TransitionGroup name="timeline-move" tag="div" class="timeline-layer">
          <div
            v-for="tick in calendarTicks"
            :key="tick.timestamp"
            class="time-row"
            :class="{ major: tick.major }"
            :style="{ top: `${tick.position}%` }"
          >
            <!-- På timskalan står dagsnamnet till vänster på den rad som
                 inleder det nya dygnet - på samma sätt som på dagskalan. -->
            <time
              :datetime="new Date(tick.timestamp).toISOString()"
              :class="[`tick-${tick.scale}`, { 'tick-daystart': tick.dayStart, hidden: tick.hidden }]"
            >
              <span v-if="tick.scale === 'day' || tick.dayStart" class="tick-weekday">{{ formatWeekday(tick.timestamp) }}</span>
              <span class="tick-value">{{ formatTick(tick.timestamp, tick.dayStart ? 'day' : tick.scale) }}</span>
            </time>
            <span class="time-line" />
            <span v-if="tick.major && tick.scale !== 'hour'" class="date-label">{{ tick.scale }}</span>
          </div>
        </TransitionGroup>

        <TransitionGroup name="timeline-move" tag="div" class="timeline-layer">
          <div
            v-for="marker in monthMarkers"
            :key="`month-${marker.timestamp}`"
            class="month-marker"
            :style="{ top: `${marker.position}%` }"
          >
            <span class="month-marker-icon" aria-hidden="true">{{ marker.icon }}</span>
            <span class="month-marker-label">{{ marker.label }}</span>
            <span class="month-marker-line" />
          </div>
        </TransitionGroup>

        <!-- Varje händelse visas som en rad med ikon, titel och starttid.
             Alla rader ligger i en enda kolumn, och en rad som krockar med
             en tidigare skjuts ned under den. -->
        <div class="detail-layer">
          <TransitionGroup name="detail-move" tag="div" class="event-layer">
            <article
              v-for="(row, rowIndex) in eventRows.rows"
              :key="row.event.occurrence_id || `${row.event.id}-${row.event.start_at}`"
              class="calendar-event"
              :class="{
                tentative: row.event.status === 'tentative',
                dragging: isDraggedEvent(row.event),
              }"
              :style="{
                top: `${row.top}px`,
                height: `${row.height}px`,
                // Rader som ligger längre ned i vyn ritas ovanpå. Då kan en
                // kort händelse ligga ovanpå en lång utan att dölja dess
                // titel.
                // En rad som dras lyfts över alla andra, så att den
                // går att se när den flyttas.
                'z-index': isDraggedEvent(row.event) ? 9999 : Math.round(row.top),
                '--event-color': row.event.color || DEFAULT_EVENT_COLOR,
                '--enter-delay': `${Math.min(rowIndex, 8) * 18}ms`,
                transform: eventDragOffset(row.event),
              }"
              @pointerdown.stop="startEventDrag($event, row.event)"
              @click.stop="handleEventClick(row.event)"
              :title="row.event.description || row.event.title"
            >
              <EventIcon v-if="row.event.icon" :icon="row.event.icon" class="event-icon" />
              <Icon
                v-if="row.event.places?.length"
                name="lucide:map-pin"
                class="event-icon"
                :title="`Platser: ${row.event.places.map((p) => p.name).join(', ')}`"
              />
              <strong class="event-title">{{ row.event.title }}</strong>
              <!-- Återkommande förekomster märks, så att det går att se
                   att de hör till en serie. -->
              <span
                v-if="recurrenceLabel(row.event)"
                class="event-recurrence"
                :title="`Återkommer: ${recurrenceLabel(row.event)}`"
              >↻</span>
              <small class="event-time">{{ formatTime(Date.parse(row.event.start_at)) }}</small>
            </article>
          </TransitionGroup>

          <TransitionGroup name="detail-move" tag="div" class="event-layer">
            <button
              v-if="eventRows.clipped.length > 0"
              type="button"
              class="overflow-chip"
              :class="{ open: openClipped }"
              :title="`${eventRows.clipped.length} händelser ryms inte`"
              @pointerdown.stop
              @click.stop="openClipped = !openClipped"
            >
              +{{ eventRows.clipped.length }} till
            </button>
          </TransitionGroup>
        </div>

        <div v-if="isNowVisible" class="now-marker" :style="{ top: `${nowPosition}%` }">
          <span>NU</span>
          <i />
        </div>

        <!-- Markören som visar var ett släppt mål skulle hamna. -->
        <div
          v-if="dropPreviewTime !== null"
          class="drop-marker"
          :style="{ top: `${timeToPosition(dropPreviewTime)}%` }"
          aria-hidden="true"
        >
          <span>släpp här</span>
        </div>

        <div v-if="coloredEvents.length === 0" class="empty-state" aria-hidden="true">
          <span class="empty-icon">◷</span>
          <p>Din tidslinje</p>
          <small>Kalenderhändelser visas här</small>
        </div>
      </div>
    </div>

    <!-- Menyn med tidsalternativ visas när ett mål släppts i kalendern.
         Den ligger i kalenderns egen koordinatsystem, så att den hamnar
         där pekaren släppte. -->
    <GoalScheduleMenu
      v-if="pendingGoalDrop"
      :style="{
        left: `${pendingGoalDrop.left}px`,
        top: `${pendingGoalDrop.top}px`,
      }"
      @select="scheduleGoal"
      @cancel="cancelGoalDrop"
    />

    <CalendarManager v-model:open="managerOpen" />
  </section>
</template>

<style scoped>
.calendar-shell {
  display: flex;
  flex-direction: column;
  position: relative;
  /* Egen staplingskontext: händelsernas z-index (upp till 9999, eftersom
     de staplas efter sin placering i tiden) ska inte kunna ritas över
     appens modaler, som ligger på z-index 100. */
  isolation: isolate;
  height: calc(100vh - 1rem);
  min-height: 38rem;
  padding: 0 1.5rem 1rem;
}

/* Panel som visar innehållet i en prick. Flera serier kan ligga under samma
   prick, och då går de inte att skilja åt i tooltipen. */
.overflow-panel {
  flex: 0 0 auto;
  margin: 0 0 0.45rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 0.6rem;
  background: var(--color-surface);
  color: var(--color-ink-muted);
  font-size: 0.74rem;
}

.overflow-panel-header {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.overflow-panel-header strong {
  color: var(--color-ink);
}

.overflow-panel-close {
  margin-left: auto;
  border: none;
  background: none;
  color: var(--color-ink-muted);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
}

.overflow-panel-note {
  margin: 0.3rem 0 0;
  opacity: 0.8;
}

.overflow-panel-list {
  margin: 0.35rem 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.overflow-panel-entry {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.15rem 0.5rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 999px;
  background: var(--color-surface);
  color: var(--color-ink-muted);
  font-size: 0.72rem;
  cursor: pointer;
}

.overflow-panel-entry:hover {
  color: var(--color-ink);
  border-color: var(--color-accent);
}

.overflow-panel-title {
  color: var(--color-ink);
}

.overflow-panel-count {
  opacity: 0.75;
  font-variant-numeric: tabular-nums;
}

/* Verktygsrad ovanför tidslinjen. Kalendrarna ligger inte här som
   etiketter - de finns i en lista som öppnas med knappen bredvid. */
.calendar-toolbar {
  flex: 0 0 auto;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.45rem;
}

.toolbar-button {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 999px;
  background: var(--color-surface);
  color: var(--color-ink);
  font-size: 0.72rem;
  cursor: pointer;
}

.toolbar-button:hover {
  border-color: var(--color-accent);
}

/* Antalet kalendrar i listan. */
.toolbar-count {
  padding: 0 0.35rem;
  border-radius: 999px;
  background: var(--color-surface-2);
  font-size: 0.66rem;
  font-variant-numeric: tabular-nums;
}

.now-button {
  border: 1px solid var(--color-line-strong);
  border-radius: 999px;
  padding: 0.3rem 0.6rem;
  font-size: 0.72rem;
}

.now-button {
  color: var(--color-ink);
  background: var(--color-surface);
  cursor: pointer;
}

.now-button:hover {
  border-color: var(--color-accent);
  background: var(--color-surface-2);
}

.primary-button {
  border-color: var(--color-accent-strong);
  background: var(--color-accent-deep);
  color: white;
}

/* Knappen som tar bort en händelse. */
.danger-button {
  border-color: #ef4444;
  color: #f87171;
}

.danger-button:hover {
  border-color: #ef4444;
  background: rgb(239 68 68 / 0.12);
}

.form-error {
  margin: 0;
  color: #f87171;
  font-size: 0.75rem;
}

.event-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  padding: 0 0 1rem;
}

.event-form input,
.event-form select {
  min-width: 0;
  flex: 1 1 10rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 0.55rem;
  padding: 0.55rem 0.7rem;
  color: var(--color-ink);
  background: var(--color-surface);
  font: inherit;
  font-size: 0.8rem;
}

.event-form input:focus,
.event-form select:focus {
  border-color: var(--color-accent);
  outline: none;
}

.event-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 100001;
  display: grid;
  place-items: center;
  padding: clamp(1rem, 4vw, 3rem);
  background: rgb(3 6 18 / 0.72);
  backdrop-filter: blur(12px) saturate(120%);
  animation: event-modal-fade-in 160ms ease-out;
}

.event-modal {
  display: flex;
  flex-direction: column;
  width: min(44rem, 100%);
  max-height: min(45rem, calc(100vh - 2rem));
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.2);
  border-radius: 1.25rem;
  background: linear-gradient(145deg, rgb(24 31 56 / 0.98), rgb(13 18 36 / 0.98));
  box-shadow: 0 2rem 6rem rgb(0 0 0 / 0.5), 0 0 0 1px rgb(99 102 241 / 0.08);
  color: var(--color-ink);
  animation: event-modal-rise-in 180ms cubic-bezier(0.22, 1, 0.36, 1);
}

.event-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.35rem 1.5rem 1.1rem;
  border-bottom: 1px solid rgb(148 163 184 / 0.12);
}

.event-modal-kicker {
  display: block;
  margin-bottom: 0.35rem;
  color: var(--color-accent);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.14em;
}

.event-modal h2 {
  margin: 0;
  color: var(--color-ink);
  font-size: 1.25rem;
  font-weight: 750;
  letter-spacing: -0.02em;
}

.event-modal-close {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  width: 2rem;
  height: 2rem;
  border: 1px solid transparent;
  border-radius: 0.55rem;
  color: var(--color-ink-dim);
  background: transparent;
  cursor: pointer;
}

.event-modal-close:hover {
  border-color: rgb(148 163 184 / 0.2);
  color: var(--color-ink);
  background: rgb(148 163 184 / 0.1);
}

.event-modal-body {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  overflow-y: auto;
  padding: 1.35rem 1.5rem;
}

.event-field {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.4rem;
}

.event-field-wide {
  grid-column: 1 / -1;
}

.event-field > span {
  color: var(--color-ink-dim);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.event-modal input,
.event-modal select,
.event-modal textarea {
  min-width: 0;
  width: 100%;
  border: 1px solid rgb(148 163 184 / 0.2);
  border-radius: 0.65rem;
  padding: 0.7rem 0.8rem;
  color: var(--color-ink);
  background: rgb(15 23 42 / 0.72);
  font: inherit;
  font-size: 0.84rem;
  transition: border-color 140ms ease, box-shadow 140ms ease, background 140ms ease;
}

.event-modal textarea {
  resize: vertical;
}

.event-modal input:focus,
.event-modal select:focus,
.event-modal textarea:focus {
  border-color: var(--color-accent);
  outline: none;
  background: rgb(15 23 42 / 0.95);
  box-shadow: 0 0 0 3px rgb(99 102 241 / 0.16);
}

.event-modal select[multiple] {
  min-height: 5.5rem;
}

.event-modal .icon-field {
  min-height: 2.55rem;
}

.event-modal .icon-button {
  width: 100%;
  justify-content: flex-start;
  border: 1px solid rgb(148 163 184 / 0.2);
  border-radius: 0.65rem;
  padding: 0.55rem 0.7rem;
  color: var(--color-ink-muted);
  background: rgb(15 23 42 / 0.72);
}

.event-modal-footer {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 1rem 1.5rem 1.25rem;
  border-top: 1px solid rgb(148 163 184 / 0.12);
}

.event-modal-spacer {
  flex: 1;
}

.event-action {
  border: 1px solid rgb(148 163 184 / 0.2);
  border-radius: 0.6rem;
  padding: 0.55rem 0.8rem;
  color: var(--color-ink-muted);
  background: rgb(148 163 184 / 0.08);
  font-size: 0.76rem;
  font-weight: 650;
  cursor: pointer;
  transition: background 140ms ease, border-color 140ms ease, color 140ms ease, transform 140ms ease;
}

.event-action:hover:not(:disabled) {
  border-color: rgb(148 163 184 / 0.35);
  color: var(--color-ink);
  background: rgb(148 163 184 / 0.15);
  transform: translateY(-1px);
}

.event-action-primary {
  border-color: rgb(129 140 248 / 0.65);
  color: white;
  background: linear-gradient(135deg, #6366f1, #7c3aed);
  box-shadow: 0 0.35rem 1rem rgb(99 102 241 / 0.24);
}

.event-action-primary:hover:not(:disabled) {
  border-color: rgb(165 180 252 / 0.9);
  background: linear-gradient(135deg, #818cf8, #8b5cf6);
}

.event-action-danger {
  color: #fca5a5;
}

.event-action-danger:hover:not(:disabled) {
  border-color: rgb(248 113 113 / 0.45);
  color: #fecaca;
  background: rgb(127 29 29 / 0.3);
}

.event-action:disabled {
  cursor: wait;
  opacity: 0.5;
}

.event-modal .form-error {
  margin: 0 1.5rem 1.25rem;
}

@keyframes event-modal-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes event-modal-rise-in {
  from { opacity: 0; transform: translateY(0.75rem) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.icon-field {
  position: relative;
  display: inline-flex;
}

.icon-button {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  min-height: 2rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 0.55rem;
  padding: 0.25rem 0.6rem;
  color: var(--color-ink-muted);
  background: var(--color-surface);
  font-size: 0.78rem;
  cursor: pointer;
}

.icon-button:hover {
  color: var(--color-ink);
  border-color: var(--color-accent);
}

.icon-preview {
  font-size: 1rem;
  line-height: 1;
}

/* Kalendern ligger direkt på sidans bakgrund - ingen ram, ingen egen
   yta. Kanten på tidslinjen dras av tidshalarna, inte av en kortyta. */
.calendar-viewport {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: transparent;
  cursor: ns-resize;
  touch-action: none;
  user-select: none;
}

.calendar-viewport.dragging {
  cursor: grabbing;
}

.calendar-viewport.zooming {
  cursor: ns-resize;
}

.calendar-grid {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background-image: linear-gradient(to right, rgb(129 140 248 / 0.025) 1px, transparent 1px);
  background-size: 25% 100%;
}

/* Lager som håller händelseraderna. Det ligger över samma område som
   tidsaxeln, så att elementen kan positioneras som vanligt. */
.detail-layer {
  position: absolute;
  inset: 0;
}

/* Lagren som håller raderna och tidsaxeln. Elementen är absolut
   positionerade, så lagren behöver ingen egen storlek - men de får
   uttryckligen position: static, så att barnen positionerar sig mot
   .detail-layer och inte mot någon annan förälder. */
.event-layer,
.timeline-layer {
  position: static;
}

/* Ticks och månadsgränser ändras när zoomnivån byter intervall. De ska
   inte dyka upp eller försvinna tvärt, utan följa samma hastighetsstyrda
   animation som kalenderinnehållet. */
.timeline-move-enter-active,
.timeline-move-leave-active {
  transition-property: opacity, scale;
  transition-duration: var(--detail-duration, 200ms);
  transition-timing-function: cubic-bezier(0.4, 0, 1, 1);
}

.timeline-move-enter-active {
  transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1);
}

.timeline-move-enter-from,
.timeline-move-leave-to {
  opacity: 0;
  scale: 0.98;
}

.timeline-move-move {
  transition-property: transform;
  transition-duration: var(--detail-duration, 200ms);
  transition-timing-function: cubic-bezier(0.22, 1, 0.36, 1);
}

/* Varje element får sin egen in- och ut-animation. Vue:s enter-klasser
   läggs på lagret ovanför, så barnen kan inte ärva dem - därför animeras
   de här.

   Längden styrs av --detail-duration, som följer scrollhastigheten: ju
   snabbare tiden komprimeras, desto kortare får elementen vara kvar. */
.detail-move-enter-active {
  transition-property: opacity, scale;
  transition-duration: var(--detail-duration, 200ms);
  transition-delay: var(--enter-delay, 0ms);
  transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1);
}

.detail-move-leave-active {
  transition-property: opacity, scale;
  transition-duration: var(--detail-duration, 200ms);
  transition-timing-function: cubic-bezier(0.4, 0, 1, 1);
}

.detail-move-enter-from {
  opacity: 0;
  scale: 0.35;
}

.detail-move-leave-to {
  opacity: 0;
  scale: 0.35;
}

/* När en prick byter kolumn, eller en händelse flyttas, glider den dit
   i stället för att hoppa. */
.detail-move-move {
  transition-property: transform;
  transition-duration: var(--detail-duration, 200ms);
  transition-timing-function: cubic-bezier(0.22, 1, 0.36, 1);
}

.time-row {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  height: 1px;
  color: var(--color-ink-dim);
  font-variant-numeric: tabular-nums;
  font-size: 0.72rem;
  translate: 0 -50%;
}

.time-row.major {
  color: var(--color-ink-muted);
}

.time-row time {
  width: 4.7rem;
  padding-left: 0.9rem;
}

.time-row time.tick-day,
.time-row time.tick-month,
.time-row time.tick-year {
  width: 6.2rem;
}

/* På dagskalan står veckodagen till vänster om datumet. Den får en egen
   smal plats så att datumet inte glider in över händelserna. */
.time-row time.tick-day {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.time-row time .tick-value {
  white-space: nowrap;
}

.tick-weekday {
  width: 1.15rem;
  color: #93c5fd;
  opacity: 0.85;
  font-size: 0.62rem;
  font-weight: 700;
  text-transform: lowercase;
}

.time-line {
  flex: 1;
  height: 1px;
  margin-left: 0.5rem;
  background: var(--color-line);
}

.time-row.major .time-line {
  background: var(--color-line-strong);
}

.time-row:has(.tick-hour) .time-line {
  background: rgb(148 163 184 / 0.28);
}

.time-row:has(.tick-day) .time-line {
  background: rgb(96 165 250 / 0.38);
}

.time-row:has(.tick-month) .time-line {
  height: 2px;
  background: rgb(167 139 250 / 0.5);
}

.time-row:has(.tick-year) .time-line {
  height: 2px;
  background: rgb(244 114 182 / 0.62);
}

.tick-day,
.tick-month,
.tick-year {
  font-weight: 700;
}

.tick-hour { color: #94a3b8; }
.tick-day { color: #60a5fa; }
.tick-month { color: #a78bfa; }
.tick-year { color: #f472b6; }

/* På timskalan inleder en rad det nya dygnet. Då står dagsnamnet till
   vänster om datumet, precis som på dagskalan. */
.time-row time.tick-daystart {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
  color: #60a5fa;
  font-weight: 700;
}

.time-row:has(.tick-daystart) .time-line {
  background: rgb(96 165 250 / 0.38);
}

.time-row time.hidden {
  visibility: hidden;
}

.time-row:has(time.hidden) .date-label {
  opacity: 0.3;
}

.month-marker {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  height: 0;
  translate: 0 -50%;
  pointer-events: none;
}

.month-marker-icon {
  width: 1.1rem;
  text-align: center;
  color: #f0abfc;
  font-size: 0.85rem;
  line-height: 1;
  text-shadow: 0 0 0.7rem rgb(240 171 252 / 0.55);
}

.month-marker-label {
  width: 6.6rem;
  padding-left: 0.25rem;
  color: #f0abfc;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.01em;
  text-transform: capitalize;
  white-space: nowrap;
  background: linear-gradient(to right, rgb(8 11 19 / 0.96) 55%, rgb(8 11 19 / 0));
  text-shadow: 0 0 0.6rem rgb(240 171 252 / 0.45);
}

.month-marker-line {
  flex: 1;
  height: 2px;
  margin-left: 0.35rem;
  border-radius: 2px;
  background: linear-gradient(
    to right,
    rgb(240 171 252 / 0.85),
    rgb(240 171 252 / 0.28) 35%,
    rgb(240 171 252 / 0.12)
  );
}

.date-label {
  width: 7rem;
  padding-left: 0.75rem;
  color: var(--color-ink-dim);
  font-size: 0.68rem;
  text-transform: capitalize;
}

/* Varje händelse är en rad: ikon, titel och starttid. Radens bredd delas
   med de andra raderna i samma klump, så att en rad aldrig döljer en
   annan. */
.calendar-event {
  position: absolute;
  --timeline-left: 5.2rem;
  --timeline-right: 8rem;
  left: var(--timeline-left);
  right: var(--timeline-right);
  display: flex;
  /* Titeln står i radens översta band, så att den ligger på händelsens
     starttid. En kort rad som ligger ovanpå en lång träffar då inte
     titeln. */
  align-items: flex-start;
  gap: 0.35rem;
  min-height: 16px;
  overflow: hidden;
  border-left: 3px solid var(--event-color, var(--color-accent));
  border-radius: 0.35rem;
  padding: 0 0.45rem;
  color: var(--color-ink);
  /* Bakgrunden är sluten, så att en rad ovanpå en annan rad går att läsa. */
  background: color-mix(in srgb, var(--event-color, #6366f1) 18%, var(--color-surface));
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--event-color, #818cf8) 30%, transparent);
  pointer-events: auto;
}

.calendar-event.tentative {
  border-left-color: #f59e0b;
  background: rgb(245 158 11 / 0.14);
}

/* En händelse som dras lyfts över de andra raderna, så att det går att
   se vilken rad som flyttas. Animationen stängs av - raden ska följa
   pekaren direkt, inte hinna ikapp. */
.calendar-event.dragging {
  cursor: grabbing;
  box-shadow: 0 8px 20px rgb(0 0 0 / 0.45);
  opacity: 0.9;
  transition: none;
}

/* Markören som visar var i tiden ett släppt mål skulle hamna. */
.drop-marker {
  position: absolute;
  left: 0.7rem;
  right: 0.7rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  height: 0;
  translate: 0 -50%;
  pointer-events: none;
  z-index: 9998;
}

.drop-marker span {
  z-index: 1;
  border-radius: 999px;
  padding: 0.14rem 0.4rem;
  color: white;
  background: var(--color-accent-strong);
  font-size: 0.6rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.drop-marker::after {
  content: '';
  flex: 1;
  height: 2px;
  background: var(--color-accent);
  box-shadow: 0 0 8px rgb(129 140 248 / 0.8);
}

.calendar-event strong,
.calendar-event small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Titeln fyller raden, och klipps med ellips när den inte ryms. */
.calendar-event strong {
  flex: 1 1 auto;
  min-width: 0;
  font-size: 0.72rem;
  font-weight: 650;
  line-height: 1.3;
}

.event-icon {
  flex: 0 0 auto;
  font-size: 0.8rem;
  line-height: 1;
  opacity: 0.85;
}

/* Märket på raden visar att händelsen är en del av en återkommande serie. */
.event-recurrence {
  flex: 0 0 auto;
  color: var(--color-accent);
  font-size: 0.68rem;
  line-height: 1;
  opacity: 0.9;
}

.calendar-event .event-time {
  flex: 0 0 auto;
  color: var(--color-ink-muted);
  font-size: 0.62rem;
  font-variant-numeric: tabular-nums;
}

/* Knappen för de händelser som inte fick plats i klumpen. Den står bredvid
   klumpens sista kolumn, i marginalen höger om tidslinjen, så att den inte
   täcker någon rad. */
.overflow-chip {
  position: absolute;
  --timeline-left: 5.2rem;
  --timeline-right: 8rem;
  right: var(--timeline-right);
  bottom: 0;
  width: 4.5rem;
  display: flex;
  align-items: center;
  min-height: 16px;
  border: 1px dashed var(--color-line-strong);
  border-radius: 0.35rem;
  padding: 0 0.4rem;
  color: var(--color-ink-muted);
  background: var(--color-surface);
  font-size: 0.66rem;
  line-height: 1.3;
  text-align: left;
  white-space: nowrap;
  cursor: pointer;
}

.overflow-chip:hover,
.overflow-chip.open {
  color: var(--color-ink);
  border-color: var(--color-accent);
}

.now-marker {
  position: absolute;
  left: 0.7rem;
  right: 0.7rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transform: translateY(-50%);
  pointer-events: none;
}

.now-marker span {
  z-index: 1;
  border-radius: 999px;
  padding: 0.17rem 0.38rem;
  color: white;
  background: var(--color-accent-strong);
  font-size: 0.6rem;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.now-marker i {
  flex: 1;
  height: 2px;
  background: var(--color-accent);
  box-shadow: 0 0 0.7rem rgb(129 140 248 / 0.75);
}

.empty-state {
  position: absolute;
  top: 50%;
  left: 50%;
  display: grid;
  justify-items: center;
  gap: 0.25rem;
  color: var(--color-ink-dim);
  text-align: center;
  transform: translate(-50%, -50%);
  pointer-events: none;
}

.empty-icon {
  color: var(--color-accent);
  font-size: 2.5rem;
  opacity: 0.65;
}

.empty-state p,
.empty-state small {
  margin: 0;
}

.empty-state p {
  color: var(--color-ink-muted);
  font-size: 0.9rem;
}

.empty-state small {
  font-size: 0.75rem;
}

@media (max-width: 640px) {
  .calendar-shell {
    height: calc(100dvh - 5.5rem);
    min-height: 0;
    padding: 0 0.45rem 0.45rem;
  }

  /* På mobil ska nästan hela ytan vara tidslinje. Titel, hjälptråd,
     skala-legenden och tangentbordsinformation hör hemma på större skärmar. */
  .now-button {
    padding: 0.32rem 0.55rem;
    font-size: 0.7rem;
    white-space: nowrap;
  }

  .event-form {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .event-modal-backdrop {
    padding: 0.6rem;
  }

  .event-modal {
    max-height: calc(100dvh - 1.2rem);
    border-radius: 1rem;
  }

  .event-modal-header,
  .event-modal-body,
  .event-modal-footer {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .event-modal-body {
    grid-template-columns: 1fr;
    gap: 0.8rem;
  }

  .event-field-wide {
    grid-column: auto;
  }

  .event-modal-footer {
    flex-wrap: wrap;
  }

  .event-modal-spacer {
    display: none;
  }

  .event-action-primary {
    order: -1;
    width: 100%;
  }

  .event-form .field-wide,
  .event-form select {
    grid-column: 1 / -1;
  }

  .month-marker-label {
    width: 4.8rem;
    padding-left: 0.2rem;
    font-size: 0.66rem;
  }

  .month-marker-icon {
    width: 0.9rem;
    font-size: 0.75rem;
  }

  .calendar-event,
  .overflow-chip {
    --timeline-left: 4.4rem;
    /* Marginalen höger om tidslinjen rymmer "+N"-knappen. */
    --timeline-right: 4rem;
  }

  .overflow-chip {
    width: 3.5rem;
    font-size: 0.6rem;
  }

  .date-label {
    width: 5.5rem;
    padding-left: 0.35rem;
    font-size: 0.6rem;
  }
}
</style>

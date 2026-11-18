<script setup lang="ts">
import type { CalendarEvent } from '~/composables/useCalendarApi'
import type { Goal } from '~/types/goal'
import { CalendarExpander } from '~/utils/calendarRecurrence'
import { DEFAULT_EVENT_COLOR, isEventHidden } from '~/utils/calendarColors'
import {
  assignClusteredLanes,
  assignLanes,
  describeRecurrence,
  groupEventsByDay,
  groupEventsByMonth,
  groupEventsByYear,
  recurrenceSeriesKey,
  resolveDetailLevel,
  resolveDetailLevelWithHysteresis,
  type CalendarMarker,
  type CalendarMarkerEntry,
  type DetailLevel,
} from '~/utils/calendarDetail'

const MIN_DURATION = 2 * 60 * 60 * 1000
// Date stöder ett mycket stort, men tekniskt ändligt, intervall.
const MAX_DATE_DURATION = 8.64e15 - Date.now()
const DEFAULT_DURATION = 24 * 60 * 60 * 1000
const ZOOM_SENSITIVITY = 650

const viewport = ref<HTMLElement | null>(null)
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
const expander = new CalendarExpander()
// Händelser i dolda kalendrar tas bort ur vyn, så att det går att fokusera
// en enda kalender - till exempel bara jobbet.
const visibleEvents = computed(() =>
  expander
    .visible(allEvents.value, startTime.value, endTime.value)
    .filter((event) => !isEventHidden(event, hiddenCalendarIds.value)),
)
// Färgen löses upp här, så att kort och prickar alltid visar samma färg:
// händelsens egen färg, annars kalenderns färg.
const coloredEvents = computed<CalendarEvent[]>(() =>
  visibleEvents.value.map((event) => ({ ...event, color: resolveColor(event) })),
)

// Återkommande serier döljs aldrig för att zoomnivån är låg. I stället
// sammanfattas de ovanför tidslinjen, med antalet förekomster i den synliga
// vyn och en knapp som hoppar till nästa förekomst. En tät rutin ska synas
// på tidslinjen, inte försvinna ur den.
interface RecurringSeriesSummary {
  key: string
  title: string
  label: string
  count: number
  nextAt: number | null
  /**
   * Sant om serien visas som ett urval av förekomster.
   *
   * Då står varje visad förekomst för flera verkliga, och det ska framgå i
   * sammanfattningen.
   */
  sampled: boolean
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

const recurringSeries = computed<RecurringSeriesSummary[]>(() => {
  const series = new Map<string, { summary: RecurringSeriesSummary; starts: number[] }>()

  for (const event of visibleEvents.value) {
    const key = recurrenceSeriesKey(event)
    if (!key) continue

    const start = Date.parse(event.start_at)
    const weight = event.occurrence_weight ?? 1
    const existing = series.get(key)

    if (existing) {
      // Vikten räknas in, så att antalet stämmer även när serien visas som
      // ett urval av förekomster.
      existing.summary.count += weight
      existing.summary.sampled = existing.summary.sampled || weight > 1
      existing.starts.push(start)
      continue
    }

    series.set(key, {
      summary: {
        key,
        title: event.title,
        label: describeRecurrence(seriesRules.value.get(key)) || 'återkommande',
        count: weight,
        nextAt: null,
        sampled: weight > 1,
      },
      starts: [start],
    })
  }

  return [...series.values()].map(({ summary, starts }) => {
    // Nästa förekomst är den första som ligger vid eller efter toppen av
    // vyn. Finns ingen tränas kalendern tillbaka till seriens första
    // förekomst i vyn.
    const upcoming = starts.filter((start) => start >= startTime.value)
    const next = upcoming.length > 0
      ? Math.min(...upcoming)
      : starts.length > 0
        ? Math.min(...starts)
        : null

    return { ...summary, nextAt: next }
  })
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
})
const iconPickerOpen = ref(false)
const goals = ref<Goal[]>([])
const {
  loadAllCalendarEvents,
  createCalendarEvent,
  updateCalendarEvent,
  createCalendarOccurrenceOverride,
  splitCalendarSeries,
} = useCalendarApi()
const { loadAllGoals } = useGoalApi()
const {
  calendars,
  hiddenCalendarIds,
  load: loadCalendars,
  isHidden: isCalendarHidden,
  toggleVisibility: toggleCalendarVisibility,
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

// Detaljnivån avgör hur mycket information varje händelse får visa. Se
// utils/calendarDetail.ts för resonemanget bakom nivåerna.
//
// Nivån hålls i en ref i stället för i en computed, eftersom bytet ska
// följa hysteresis: en gräns som passeras i båda riktningar får inte
// fladdra fram och tillbaka när användaren scrollar över den.
const detailLevel = ref<DetailLevel>(
  resolveDetailLevel(visibleDuration.value),
)

watch(visibleDuration, (duration) => {
  detailLevel.value = resolveDetailLevelWithHysteresis(
    duration,
    detailLevel.value,
  )
})

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

const formatDate = (timestamp: number) =>
  new Intl.DateTimeFormat('sv-SE', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
  }).format(timestamp)

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
  const result: Array<{ timestamp: number; position: number; major: boolean }> = []

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

const eventStyle = (event: CalendarEvent, index: number) => {
  const eventStart = Math.max(Date.parse(event.start_at), startTime.value)
  const eventEnd = Math.min(Date.parse(event.end_at), endTime.value)
  const top = timeToPosition(eventStart)
  const height = Math.max(2, timeToPosition(eventEnd) - top)
  const placement = eventLanes.value[index] ?? { lane: 0, lanes: 1 }

  return {
    top: `${top}%`,
    height: `${height}%`,
    '--lane': `${placement.lane}`,
    '--lanes': `${placement.lanes}`,
    '--event-color': event.color || DEFAULT_EVENT_COLOR,
  }
}

// Händelser som ligger nära varandra i tid läggs ovanpå varandra, med en
// liten förskjutning åt sidan så att alla går att läsa och klicka på.
// Förskjutningen delas bara mellan de händelser som faktiskt överlappar, så
// att en ensam händelse får hela bredden.
const eventLanes = computed<Array<{ lane: number; lanes: number }>>(() => {
  if (detailLevel.value !== 'full' && detailLevel.value !== 'compact') {
    return coloredEvents.value.map(() => ({ lane: 0, lanes: 1 }))
  }

  const intervals = coloredEvents.value.map((event): [number, number] => {
    const start = Math.max(Date.parse(event.start_at), startTime.value)
    const end = Math.min(Date.parse(event.end_at), endTime.value)
    const from = timeToPosition(start)
    // Korten har en minsta höjd, så även ett kort händelse ska rymmas.
    const height = Math.max(timeToPosition(end) - from, markerGap.value)

    return [from, from + height]
  })

  return assignClusteredLanes(intervals, cardMargin.value)
})

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

// Hoppar till nästa förekomst av en serie och zoomar in till det dygn
// förekomsten ligger i.
const jumpToSeries = (series: RecurringSeriesSummary) => {
  if (series.nextAt === null) return

  startTime.value = series.nextAt - 60 * 60 * 1000
  zoomTo(24 * 60 * 60 * 1000)
}

const panBy = (distance: number) => {
  const height = viewport.value?.clientHeight || 600
  startTime.value += distance * visibleDuration.value / height
}

const panTo = (distance: number, originStart: number, originDuration: number) => {
  const height = viewport.value?.clientHeight || 600
  startTime.value = originStart + distance * originDuration / height
}

// På låg zoom räcker det inte att visa varje händelse som ett eget kort -
// då blir de för tunna att läsa eller ens klicka på. Därför grupperas de
// per dag, månad eller år och visas som prickar.
type RenderedMarker = CalendarMarker & { intensity: number; lane: number }

// Minsta avståndet mellan två prickar, i pixlar. Utan det skulle prickar som
// ligger nära varandra i tid ligga ovanpå varandra. Värdet räknas om till
// procent av tidsfönstret eftersom positionerna är relativt.
const MARKER_MIN_GAP_PX = 22

const viewportHeight = ref(0)
let viewportObserver: ResizeObserver | undefined

const markerGap = computed(() => {
  if (viewportHeight.value <= 0) return MARKER_MIN_GAP_PX

  return (MARKER_MIN_GAP_PX / viewportHeight.value) * 100
})

// Liten marginal mellan två kort som ligger bredvid varandra i tid.
const CARD_MARGIN_PX = 2

const cardMargin = computed(() => {
  if (viewportHeight.value <= 0) return CARD_MARGIN_PX

  return (CARD_MARGIN_PX / viewportHeight.value) * 100
})

const markers = computed<RenderedMarker[]>(() => {
  const groups = detailLevel.value === 'daily'
    ? groupEventsByDay(coloredEvents.value)
    : detailLevel.value === 'monthly'
      ? groupEventsByMonth(coloredEvents.value)
      : detailLevel.value === 'yearly'
        ? groupEventsByYear(coloredEvents.value)
        : []

  if (groups.length === 0) return groups

  // Endast grupper som ligger i det synliga fönstret ska renderas.
  const visible = groups.filter((group) => {
    const position = timeToPosition(group.startAt)
    return position >= -5 && position <= 105
  })

  if (visible.length === 0) return visible

  const positions = visible.map((group) =>
    timeToPosition(group.startAt + (group.endAt - group.startAt) / 2),
  )
  const lanes = assignLanes(positions, markerGap.value)

  return visible.map((group, index) => ({
    ...group,
    lane: lanes[index],
    // Tätheten ska kännas igen utan att antalet ska läsas. Från tre
    // händelser och uppåt är det full styrka.
    intensity: 0.45 + 0.55 * Math.min(1, group.count / 3),
  }))
})

const markerStyle = (marker: RenderedMarker) => {
  const center = marker.startAt + (marker.endAt - marker.startAt) / 2

  return {
    top: `${timeToPosition(center)}%`,
    '--lane': `${marker.lane}`,
    '--event-color': marker.color || DEFAULT_EVENT_COLOR,
    // Intensiteten visar hur många händelser som ligger i perioden. Den
    // appliceras som ett filter så att opaciteten är fri för in-animationen,
    // annan hoppar pricken tillbaka till sitt gamla utseende efteråt.
    filter: `opacity(${marker.intensity})`,
  }
}

// Klick på en prick hoppar in i tiden. En dag blir ett dygn, en månad blir
// ungefär en månad och ett år blir ett år, så att man landar på rätt
// detaljnivå.
//
// Innehållet i pricken sparas samtidigt, så att det går att se: en prick kan
// samla flera serier och enstaka händelser, och då ska de gå att skilja åt
// utan att zooma in.
const selectedMarker = ref<RenderedMarker | null>(null)

// Pricken som pekas på just nu. Den visas direkt, och den ersätter den
// markerade pricken i panelen så länge pekaren är kvar.
const hoveredMarker = ref<RenderedMarker | null>(null)

const visibleMarker = computed(() => hoveredMarker.value ?? selectedMarker.value)

const jumpToMarker = (marker: RenderedMarker) => {
  selectedMarker.value = marker
  startTime.value = marker.startAt

  if (detailLevel.value === 'daily') {
    zoomTo(24 * 60 * 60 * 1000)
    return
  }

  if (detailLevel.value === 'monthly') {
    zoomTo(30 * 24 * 60 * 60 * 1000)
    return
  }

  zoomTo(365 * 24 * 60 * 60 * 1000)
}

// Hoppar till den förekomst som en post i markören representerar.
const jumpToEntry = (entry: CalendarMarkerEntry) => {
  const start = Date.parse(entry.event.start_at)

  startTime.value = start - 60 * 60 * 1000
  zoomTo(24 * 60 * 60 * 1000)
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
  draftEvent.startAt = toLocalInput(start)
  draftEvent.endAt = toLocalInput(end)
  showCreateEvent.value = true
}

const openEditEvent = (event: CalendarEvent) => {
  editingEvent.value = event
  editScope.value = event.recurrence_rule ? 'occurrence' : 'series'
  draftEvent.title = event.title
  draftEvent.description = event.description || ''
  draftEvent.goalId = event.goal_id ? String(event.goal_id) : ''
  draftEvent.recurrenceRule = event.recurrence_rule || ''
  draftEvent.icon = event.icon || ''
  draftEvent.calendarId = event.calendar_id ? String(event.calendar_id) : ''
  draftEvent.startAt = toLocalInput(new Date(event.start_at))
  draftEvent.endAt = toLocalInput(new Date(event.end_at))
}

const closeEditEvent = () => {
  editingEvent.value = null
}

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
    if (editScope.value === 'occurrence' && event.recurrence_rule && event.original_start_at) {
      await createCalendarOccurrenceOverride(event.id, {
        original_start_at: event.original_start_at,
        ...payload,
      })
    } else if (editScope.value === 'following' && event.recurrence_rule && event.original_start_at) {
      await splitCalendarSeries(event.id, {
        original_start_at: event.original_start_at,
        recurrence_rule: draftEvent.recurrenceRule || event.recurrence_rule,
        ...payload,
      })
    } else {
      await updateCalendarEvent(event.id, {
        ...payload,
        ...(draftEvent.recurrenceRule ? { recurrence_rule: draftEvent.recurrenceRule } : {}),
      })
    }
    closeEditEvent()
    await loadEvents()
  } finally {
    savingEvent.value = false
  }
}

const saveEvent = async () => {
  if (!draftEvent.title.trim() || !draftEvent.startAt || !draftEvent.endAt) return

  savingEvent.value = true
  try {
    await createCalendarEvent({
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

const handleCalendarKeydown = (event: KeyboardEvent) => {
  const target = event.target as HTMLElement | null
  if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT' || target.isContentEditable)) return

  if (event.key === 'Escape') {
    if (editingEvent.value) closeEditEvent()
    else if (showCreateEvent.value) showCreateEvent.value = false
    return
  }
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
  } catch (error) {
    console.warn('Kunde inte läsa mål för kalenderhändelser:', error)
  }
  window.addEventListener('keydown', handleCalendarKeydown)

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
})
</script>

<template>
  <section class="calendar-shell">
    <!-- Kalendrarna syns som en rad över tidslinjen. Klick på en kalender
         döljer eller visar dess händelser. -->
    <div class="calendar-legend">
      <button
        v-for="calendar in calendars"
        :key="calendar.id"
        type="button"
        class="legend-chip"
        :class="{ hidden: isCalendarHidden(calendar.id) }"
        :aria-pressed="!isCalendarHidden(calendar.id)"
        @click="toggleCalendarVisibility(calendar.id)"
      >
        <span class="legend-dot" :style="{ background: calendar.color }" />
        {{ calendar.name }}
        <span v-if="calendar.is_default" class="legend-default" title="Standardkalender">★</span>
      </button>
      <button type="button" class="legend-add" @click="openCreateEvent">
        + Händelse
      </button>
      <button type="button" class="legend-add" @click="managerOpen = true">
        + Kalender
      </button>
    </div>

    <form v-if="editingEvent" class="event-form edit-form" @submit.prevent="saveEditedEvent">
      <strong>Redigera händelse</strong>
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
      <input v-model="draftEvent.description" placeholder="Beskrivning" aria-label="Beskrivning" class="field-wide" />
      <select v-if="editingEvent.recurrence_rule" v-model="editScope" aria-label="Ändra serie">
        <option value="occurrence">Bara denna förekomst</option>
        <option value="following">Denna och följande</option>
        <option value="series">Hela serien</option>
      </select>
      <button type="submit" class="now-button primary-button" :disabled="savingEvent">Spara</button>
      <button type="button" class="now-button" @click="closeEditEvent">Avbryt</button>
    </form>

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

    <!-- Återkommande serier visas alltid, oavsett zoomnivå. Sammanfattningen
         här ersätter dold utrymme: antalet förekomster i vyn syns direkt, och
         knappen hoppar till nästa förekomst. -->
    <div v-if="recurringSeries.length > 0" class="recurrence-summary">
      <button
        v-for="series in recurringSeries"
        :key="series.key"
        type="button"
        class="recurrence-series"
        :title="`Hoppa till nästa förekomst av ${series.title}`"
        @click="jumpToSeries(series)"
      >
        <span class="recurrence-icon" aria-hidden="true">↻</span>
        <strong>{{ series.title }}</strong>
        <span>{{ series.label }}</span>
        <span class="recurrence-count">{{ series.count }} i vyn</span>
        <!-- En serie som tunnas ut visas som ett urval. Det ska stå i
             sammanfattningen, annars ser antalet ut att vara en bild. -->
        <span v-if="series.sampled" class="recurrence-sampled">urval</span>
      </button>
    </div>

    <!-- Innehållet i en vald prick. Flera serier kan ligga under samma
         prick, och då går de inte att skilja åt i tooltipen. Panelen listar
         dem var för sig, med sina antal. -->
    <div v-if="visibleMarker" class="marker-detail">
      <div class="marker-detail-header">
        <strong>{{ visibleMarker.label }}</strong>
        <span>{{ visibleMarker.count }} händelser</span>
        <button type="button" class="marker-detail-close" title="Stäng" @click="selectedMarker = null; hoveredMarker = null">×</button>
      </div>
      <!-- När en serie tunnas ut visas ett urval av förekomsterna. Det ska
           framgå, annars ser antalen ut att vara exakta. -->
      <p v-if="visibleMarker.approximate" class="marker-detail-note">
        Vissa serier visas som ett urval av förekomsterna. Antalet är ändå exakt.
      </p>
      <ul class="marker-detail-list">
        <li v-for="entry in visibleMarker.entries" :key="entry.key">
          <button type="button" class="marker-detail-entry" @click="jumpToEntry(entry)">
            <span v-if="entry.recurring" class="marker-detail-icon" aria-hidden="true">↻</span>
            <span class="marker-detail-title">{{ entry.title }}</span>
            <span class="marker-detail-count">{{ entry.count }}</span>
            <span v-if="entry.weight > 1" class="marker-detail-sampled">
              visar 1 av {{ entry.weight }} per {{ entry.period === 'month' ? 'månad' : 'dag' }}
            </span>
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
    >
      <div
        class="calendar-grid"
        :data-detail="detailLevel"
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
            <time
              :datetime="new Date(tick.timestamp).toISOString()"
              :class="[`tick-${tick.scale}`, { hidden: tick.hidden }]"
            >
              {{ formatTick(tick.timestamp, tick.scale) }}
            </time>
            <span class="time-line" />
            <span v-if="tick.major" class="date-label">{{ tick.scale === 'hour' ? formatDate(tick.timestamp) : tick.scale }}</span>
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

        <!-- Timmar och dagar: varje händelse visas som ett eget kort.
             Månader och år: prickar per dag respektive per månad.

             Nivåväxlingen animeras, så att innehållet tonar in och ut i
             stället för att bytas blixtsnabbt. -->
        <Transition name="detail-swap">
          <div :key="detailLevel" class="detail-layer">
            <template v-if="detailLevel === 'full' || detailLevel === 'compact'">
              <TransitionGroup name="detail-move" tag="div" class="event-layer">
                <article
                  v-for="(event, eventIndex) in coloredEvents"
                  :key="event.occurrence_id || `${event.id}-${event.start_at}`"
                  class="calendar-event"
                  :class="[detailLevel, { tentative: event.status === 'tentative', 'has-lanes': (eventLanes[eventIndex]?.lanes ?? 1) > 1 }]"
                  :style="{ ...eventStyle(event, eventIndex), '--enter-delay': `${Math.min(eventIndex, 8) * 18}ms` }"
                  @pointerdown.stop
                  @click.stop="openEditEvent(event)"
                  :title="event.description || event.title"
                >
                  <span class="event-header">
                    <span class="event-title">
                      <EventIcon v-if="event.icon" :icon="event.icon" class="event-icon" />
                      <strong>{{ event.title }}</strong>
                      <!-- Återkommande förekomster märks, så att det går att se
                           att de hör till en serie. -->
                      <span
                        v-if="recurrenceLabel(event)"
                        class="event-recurrence"
                        :title="`Återkommer: ${recurrenceLabel(event)}`"
                      >↻</span>
                    </span>
                    <small class="event-time">{{ formatTime(Date.parse(event.start_at)) }}</small>
                  </span>
                </article>
              </TransitionGroup>
            </template>

            <template v-else>
              <TransitionGroup name="detail-move" tag="div" class="marker-layer">
                <button
                  v-for="(marker, markerIndex) in markers"
                  :key="marker.key"
                  type="button"
                  class="calendar-marker"
                  :class="[detailLevel, { tentative: marker.tentative, selected: visibleMarker?.key === marker.key }]"
                  :style="{ ...markerStyle(marker), '--enter-delay': `${Math.min(markerIndex, 12) * 14}ms` }"
                  :title="marker.title"
                  :aria-label="marker.title"
                  @pointerdown.stop
                  @mouseenter="hoveredMarker = marker"
                  @mouseleave="hoveredMarker = null"
                  @focus="selectedMarker = marker"
                  @click.stop="jumpToMarker(marker)"
                >
                  <!-- Antalet visas bara när det är fler än ett. Ett "1" säger ingenting. -->
                  <span v-if="detailLevel === 'daily' && marker.count > 1" class="marker-count">{{ marker.count }}</span>
                </button>
              </TransitionGroup>
            </template>
          </div>
        </Transition>

        <div v-if="isNowVisible" class="now-marker" :style="{ top: `${nowPosition}%` }">
          <span>NU</span>
          <i />
        </div>

        <div v-if="coloredEvents.length === 0" class="empty-state" aria-hidden="true">
          <span class="empty-icon">◷</span>
          <p>Din tidslinje</p>
          <small>Kalenderhändelser visas här</small>
        </div>
      </div>
    </div>

    <CalendarManager v-model:open="managerOpen" />
  </section>
</template>

<style scoped>
.calendar-shell {
  display: flex;
  flex-direction: column;
  position: relative;
  height: calc(100vh - 1rem);
  min-height: 38rem;
  padding: 0 1.5rem 1rem;
}

/* Sammanfattning av de återkommande serierna i vyn. Den ersätter den
   gamla notisen om dolda förekomster: ingenting döljs, men serierna syns
   här med sitt antal och går att klicka för att hoppa till nästa
   förekomst. */
.recurrence-summary {
  flex: 0 0 auto;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  margin: 0 0 0.45rem;
}

.recurrence-series {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 999px;
  padding: 0.2rem 0.6rem;
  background: var(--color-surface);
  color: var(--color-ink-muted);
  font-size: 0.72rem;
  cursor: pointer;
}

.recurrence-series:hover {
  color: var(--color-ink);
  border-color: var(--color-accent);
}

.recurrence-icon {
  color: var(--color-accent);
  line-height: 1;
}

.recurrence-series strong {
  color: var(--color-ink);
  font-weight: 650;
}

.recurrence-count {
  opacity: 0.75;
  font-variant-numeric: tabular-nums;
}

/* En serie som tunnas ut visas som ett urval av förekomsterna. Det ska
   framgå i sammanfattningen, annars ser antalet ut att vara en bild. */
.recurrence-sampled {
  padding: 0 0.35rem;
  border-radius: 999px;
  background: var(--color-surface-raised, rgb(129 140 248 / 0.12));
  font-size: 0.66rem;
  letter-spacing: 0.02em;
  opacity: 0.9;
}

/* Panel som visar innehållet i en prick. Flera serier kan ligga under samma
   prick, och då går de inte att skilja åt i tooltipen. */
.marker-detail {
  flex: 0 0 auto;
  margin: 0 0 0.45rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--color-line-strong);
  border-radius: 0.6rem;
  background: var(--color-surface);
  color: var(--color-ink-muted);
  font-size: 0.74rem;
}

.marker-detail-header {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.marker-detail-header strong {
  color: var(--color-ink);
}

.marker-detail-close {
  margin-left: auto;
  border: none;
  background: none;
  color: var(--color-ink-muted);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
}

.marker-detail-note {
  margin: 0.3rem 0 0;
  opacity: 0.8;
}

.marker-detail-list {
  margin: 0.35rem 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.marker-detail-entry {
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

.marker-detail-entry:hover {
  color: var(--color-ink);
  border-color: var(--color-accent);
}

.marker-detail-icon {
  color: var(--color-accent);
  line-height: 1;
}

.marker-detail-title {
  color: var(--color-ink);
}

.marker-detail-count {
  opacity: 0.75;
  font-variant-numeric: tabular-nums;
}

/* En post som visas som ett urval av förekomsterna. */
.marker-detail-sampled {
  padding: 0 0.3rem;
  border-radius: 999px;
  background: rgb(129 140 248 / 0.14);
  font-size: 0.64rem;
  opacity: 0.9;
}

/* Rad med kalendrarna över tidslinjen. Klick på en kalender döljer den. */
.calendar-legend {
  flex: 0 0 auto;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.45rem;
}

.legend-chip,
.legend-add {
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

.legend-chip:hover,
.legend-add:hover {
  border-color: var(--color-accent);
}

.legend-chip.hidden {
  opacity: 0.45;
  text-decoration: line-through;
}

.legend-dot {
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 999px;
}

/* Stjärnan markerar standardkalendern, alltså kalendern som nya
   händelser hamnar i. */
.legend-default {
  color: #fbbf24;
  font-size: 0.7rem;
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

/* Lager som håller korten respektive prickarna. Det ligger över samma
   område som tidsaxeln, så att elementen kan positioneras som vanligt. */
.detail-layer {
  position: absolute;
  inset: 0;
}

/* Lagren som håller korten och prickarna. Elementen är absolut
   positionerade, så lagren behöver ingen egen storlek - men de får
   uttryckligen position: static, så att barnen positionerar sig mot
   .detail-layer och inte mot någon annan förälder. */
.event-layer,
.marker-layer,
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

/* Nivåväxlingen tonar in och ut i stället för att bytas blixtsnabbt. */
.detail-swap-enter-active {
  transition: opacity 0.2s ease, transform 0.24s cubic-bezier(0.22, 1, 0.36, 1);
}

.detail-swap-leave-active {
  transition: opacity 0.1s ease;
}

.detail-swap-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.985);
}

.detail-swap-leave-to {
  opacity: 0;
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

.calendar-event {
  position: absolute;
  --timeline-left: 5.2rem;
  --timeline-right: 8rem;
  left: var(--timeline-left);
  right: var(--timeline-right);
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  overflow: hidden;
  border-left: 3px solid var(--event-color, var(--color-accent));
  border-radius: 0.35rem;
  padding: 0.2rem 0.55rem;
  color: var(--color-ink);
  background: color-mix(in srgb, var(--event-color, #6366f1) 20%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--event-color, #818cf8) 30%, transparent);
  pointer-events: auto;
}

/* Händelser som krockar i tid läggs ovanpå varandra i stället för att
   delas upp i kolumner. Varje kort flyttas ett litet steg åt höger, så att
   även de underliggande kortens titlar syns. Högerkanten ligger kvar på
   samma ställe för alla kort i gruppen. */
.calendar-event.has-lanes {
  --lane-step: 0.9rem;
  left: calc(var(--timeline-left) + var(--lane, 0) * var(--lane-step));
}

.calendar-event.tentative {
  border-left-color: #f59e0b;
  background: rgb(245 158 11 / 0.14);
}

.calendar-event strong,
.calendar-event small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.calendar-event strong {
  font-size: 0.75rem;
  font-weight: 650;
}

/* Rubriken ligger alltid överst i kortet: titeln till vänster och
   starttiden alltid på samma ställe, i högra hörnet. */
.event-header {
  align-items: baseline;
  display: flex;
  gap: 0.5rem;
  min-width: 0;
}

.event-title {
  align-items: center;
  display: flex;
  flex: 1 1 auto;
  gap: 0.3rem;
  min-width: 0;
}

.event-icon {
  flex: 0 0 auto;
  font-size: 0.85rem;
  line-height: 1;
  opacity: 0.85;
}

/* Märket på kortet visar att händelsen är en del av en återkommande serie. */
.event-recurrence {
  flex: 0 0 auto;
  color: var(--color-accent);
  font-size: 0.7rem;
  line-height: 1;
  opacity: 0.9;
}

.calendar-event .event-time {
  flex: 0 0 auto;
  color: var(--color-ink-muted);
  font-size: 0.65rem;
  font-variant-numeric: tabular-nums;
}

/* Dagnivå: kortet ska gå att läsa och klicka på även när det är tunt.
   Klockslaget ligger alltid i högra hörnet, även här. */
.calendar-event.compact {
  min-height: 1.4rem;
  padding: 0.1rem 0.45rem;
}

.calendar-event.compact strong {
  font-size: 0.7rem;
}

.calendar-event.compact .event-time {
  font-size: 0.6rem;
}

/* Månads- och årsnivå: prickar i stället för kort. Prickar som ligger nära
   varandra i tid hamnar i olika kolumner (lanes) så att de inte överlappar. */
.calendar-marker {
  position: absolute;
  left: calc(var(--marker-left, 5.2rem) + var(--lane, 0) * var(--marker-lane-step));
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  padding: 0;
  cursor: pointer;
  pointer-events: auto;
  background: color-mix(in srgb, var(--event-color, #818cf8) 90%, transparent);
}

.calendar-marker.daily {
  --marker-lane-step: 2rem;
  width: 1.85rem;
  height: 1.15rem;
  translate: 0 -50%;
  border-radius: 0.3rem;
  color: #0b1020;
  font-size: 0.62rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.calendar-marker.monthly {
  --marker-lane-step: 0.95rem;
  width: 0.8rem;
  height: 0.8rem;
  translate: 0 -50%;
  border-radius: 999px;
  background: color-mix(in srgb, var(--event-color, #c4b5fd) 95%, transparent);
}

.calendar-marker.yearly {
  --marker-lane-step: 1.15rem;
  width: 0.95rem;
  height: 0.95rem;
  translate: 0 -50%;
  border-radius: 999px;
  background: color-mix(in srgb, var(--event-color, #f0abfc) 95%, transparent);
}

/* Osynlig yta runt pricken så att den går att klicka på även när den är
   liten och ligger nära en granne. */
.calendar-marker::before {
  content: '';
  position: absolute;
  inset: -0.3rem;
}

.calendar-marker.tentative {
  background: color-mix(in srgb, #f59e0b 90%, transparent);
}

.calendar-marker:hover {
  outline: 2px solid rgb(226 232 240 / 0.7);
  outline-offset: 1px;
}

/* Pricken som panelen visar innehållet i. */
.calendar-marker.selected {
  outline: 2px solid var(--color-accent);
  outline-offset: 1px;
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

  .calendar-event {
    --timeline-left: 4.4rem;
    --timeline-right: 0.5rem;
    left: var(--timeline-left);
    right: var(--timeline-right);
  }

  .calendar-marker {
    --marker-left: 4.4rem;
  }

  .calendar-marker.daily {
    --marker-lane-step: 1.75rem;
    width: 1.6rem;
  }

  .calendar-marker.monthly {
    --marker-lane-step: 0.85rem;
  }

  .calendar-marker.yearly {
    --marker-lane-step: 1rem;
  }

  .date-label {
    width: 5.5rem;
    padding-left: 0.35rem;
    font-size: 0.6rem;
  }
}
</style>

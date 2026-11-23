<script setup lang="ts">
import type { Goal, GoalParent } from '~/types/goal'
import { canBeFinished } from '~/utils/goalRules'
import { dueTone, formatDueDate, isFutureDate, todayDateString } from '~/utils/dueDate'
import { describeGoalRecurrence, isRecurringGoal, isWaitingForNextOccurrence, recurrenceCount } from '~/utils/recurrence'
import type { RecurrenceMode } from '~/utils/recurrence'
import { ETERNAL_ACCENT, ETERNAL_TEXT, ETERNAL_TEXT_HOVER } from '~/utils/eternalStyle'

definePageMeta({
  ssr: false,
})

const { user } = useUserSession()
const { toggleGoalFinished, toggleGoalStarted, setGoalEternal, setGoalAgent, setGoalNow, setGoalDue, setGoalStartDate, setGoalEstimate, updateGoalDescription, deleteGoal, completeGoalOccurrence, setGoalRecurrence } = useGoalApi()
const { revision: undoRevision } = useUndoStack()
const { undoFinished, undoOccurrence, undoStarted, undoEternal, undoAgent, undoNow, undoDue, undoStartDate, undoEstimate, undoRecurrence } = useGoalUndo()
const goalDates = useGoalDatesModal()
// Målet kan dras in i kalendern för att schemaläggas.
const { startDrag } = useGoalDrag()

function dueToneClass(due: string | null): string {
  const tone = dueTone(due)
  if (tone === 'overdue') return 'text-red-400'
  if (tone === 'today') return 'text-amber-400'
  return 'text-ink-dim'
}

type PrioritizedGoalRow = Goal & {
  weight: number
  parentTitle: string | null
  parents: GoalParent[]
}

const {
  isPriorityMode,
  selectedGoalId,
  prioritizedGoals,
  visibleGoals,
  isLoading,
  nowFilter,
  filterByCurrentPlace,
  loadPrioritizedGoals,
  selectGoal,
  setNowFilter,
  setFilterByCurrentPlace,
} = usePriorityMode()

// Platser: "jag är här"-platsen styr filtret "görbar här".
const { places: allPlaces, currentPlaceId, fetchPlaces, fetchCurrentPlace, setCurrentPlace, findNearbyPlaces } = usePlaces()
const gpsMessage = ref('')
const locating = ref(false)

onMounted(async () => {
  await Promise.all([
    fetchPlaces().catch(() => console.error('Failed to fetch places')),
    fetchCurrentPlace().catch(() => console.error('Failed to fetch current place')),
  ])
})

// Slår på GPS och matchar positionen mot platsernas radie.
async function locateMe() {
  gpsMessage.value = ''

  if (!('geolocation' in navigator)) {
    gpsMessage.value = 'GPS stöds inte i den här webbläsaren.'
    return
  }

  locating.value = true
  try {
    const position = await new Promise<GeolocationPosition>((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(resolve, reject, {
        enableHighAccuracy: true,
        timeout: 10000,
      })
    })

    const nearby = await findNearbyPlaces(position.coords.latitude, position.coords.longitude)
    if (nearby.length === 0) {
      gpsMessage.value = 'Ingen plats matchade din position.'
      return
    }

    await setCurrentPlace(nearby[0].id)
    gpsMessage.value = `Du är vid ${nearby[0].name}.`
  } catch (error: any) {
    gpsMessage.value = error?.code === 1
      ? 'Platsdelning nekades - välj platsen manuellt.'
      : 'Kunde inte hitta din position.'
  } finally {
    locating.value = false
  }
}

async function selectCurrentPlace(placeId: number | null) {
  gpsMessage.value = ''
  await setCurrentPlace(placeId)
}

// Scrolla till markerat mål när j/k ändrar urvalet, precis som på /goals
// och målsidorna. `nearest` gör att sidan bara flyttar sig när raden hamnar
// utanför den synliga delen av viewporten.
watch(selectedGoalId, async () => {
  await nextTick()
  if (selectedGoalId.value === null) return

  const selectedElement = document.querySelector(
    `[data-goal-id="${selectedGoalId.value}"]`,
  ) as HTMLElement | null

  selectedElement?.scrollIntoView({
    behavior: 'smooth',
    block: 'nearest',
  })
})

// När något har ångras kan listan ha ändrats - hämta om den.
watch(undoRevision, () => {
  loadPrioritizedGoals()
})
watch(goalDates.savedRevision, () => {
  loadPrioritizedGoals()
})

onMounted(() => {
  loadPrioritizedGoals()

  const onKeydown = (e: KeyboardEvent) => {
    const tag = (e.target as HTMLElement).tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || (e.target as HTMLElement).isContentEditable) return

    // Escape stänger delete-dialogen eller beskrivningseditorn
    if (e.key === 'Escape') {
      if (showDeleteConfirm.value) {
        e.preventDefault()
        showDeleteConfirm.value = false
        goalToDelete.value = null
        return
      }
      if (editingDescriptionGoalId.value !== null) {
        e.preventDefault()
        cancelEditDescription()
        return
      }
    }

    const goal = visibleGoals.value.find(g => g.id === selectedGoalId.value)

    if (e.key === 's' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) toggleStarted(goal)
      return
    }

    // Samma tangentkommandon som på målsidan och /goals:
    // d = klar, e = evigt, c = deadline, t = beskrivning, x = ta bort
    if (e.key === 'd' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) toggleFinished(goal)
      return
    }

    if (e.key === 'e' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) toggleEternal(goal)
      return
    }

    if (e.key === 'c' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) openDuePicker(goal)
      return
    }

    if (e.key === 'b' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) openStartPicker(goal)
      return
    }

    if (e.key === 'm' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) openEstimateEditor(goal)
      return
    }

    if (e.key === 't' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) startEditDescription(goal)
      return
    }

    // r = upprepning
    if (e.key === 'r' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) openRecurrenceEditor(goal)
      return
    }

    if (e.key === 'x' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      if (goal) {
        goalToDelete.value = goal
        showDeleteConfirm.value = true
      }
    }
  }
  window.addEventListener('keydown', onKeydown)
  onUnmounted(() => window.removeEventListener('keydown', onKeydown))
})

async function toggleFinished(goal: PrioritizedGoalRow) {
  // Eviga mål pågår hela livet och kan inte slutföras.
  if (!canBeFinished(goal)) {
    return
  }

  // Ett återkommande mål slutförs en förekomst i taget. Målet ligger kvar
  // i listan och väntar på nästa upprepning.
  if (isRecurringGoal(goal)) {
    // Förekomsten har inte kommit än - då finns inget att checka av.
    if (isWaitingForNextOccurrence(goal)) {
      return
    }
    try {
      await completeGoalOccurrence(goal.id, goal.recurrence_next ?? undefined, todayDateString())
      undoOccurrence(goal)
      await loadPrioritizedGoals()
    } catch (err) {
      console.error('Failed to complete occurrence:', err)
    }
    return
  }

  try {
    const previousFinished = goal.finished
    const newFinishedValue = previousFinished ? null : new Date().toISOString()
    await toggleGoalFinished(goal.id, newFinishedValue)
    undoFinished(goal, previousFinished)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal finished:', err)
  }
}

async function toggleStarted(goal: PrioritizedGoalRow) {
  try {
    const previousStarted = goal.started
    const newStartedValue = previousStarted ? null : new Date().toISOString()
    await toggleGoalStarted(goal.id, newStartedValue)
    undoStarted(goal, previousStarted)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal started:', err)
  }
}

// Växla evigt-märket på ett mål
async function toggleEternal(goal: PrioritizedGoalRow) {
  try {
    const previousEternal = goal.eternal
    await setGoalEternal(goal.id, !goal.eternal)
    undoEternal(goal, previousEternal)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal eternal:', err)
  }
}

// Växla agent-märket på ett mål
async function toggleAgent(goal: PrioritizedGoalRow) {
  try {
    const previousAgent = goal.agent ?? false
    await setGoalAgent(goal.id, !goal.agent)
    undoAgent(goal, previousAgent)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal agent:', err)
  }
}

// Växla nu-märket på ett mål. Målet ligger då i "Gör nu"-listan.
async function toggleNow(goal: PrioritizedGoalRow) {
  try {
    const previousNow = goal.now ?? false
    await setGoalNow(goal.id, !goal.now)
    undoNow(goal, previousNow)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal now:', err)
  }
}

// Leader + a växlar agent-märket på det markerade målet.
const { registerPageCommands } = useLeaderKey()
let unregisterLeaderCommands: (() => void) | null = null

onMounted(() => {
  unregisterLeaderCommands = registerPageCommands([
    {
      keys: 'a',
      label: 'A',
      title: 'Växla agent-märke',
      run: () => {
        const goal = visibleGoals.value.find(g => g.id === selectedGoalId.value)
        if (goal) {
          toggleAgent(goal)
        }
      },
    },
    {
      keys: 'n',
      label: 'N',
      title: 'Växla nu-märke',
      run: () => {
        const goal = visibleGoals.value.find(g => g.id === selectedGoalId.value)
        if (goal) {
          toggleNow(goal)
        }
      },
    },
    {
      keys: 'd',
      label: 'D',
      title: 'Sätt startdatum och deadline',
      run: () => {
        const goal = visibleGoals.value.find(g => g.id === selectedGoalId.value)
        if (goal) {
          goalDates.openFor(goal)
        }
      },
    },
  ])
})

onUnmounted(() => {
  unregisterLeaderCommands?.()
  unregisterLeaderCommands = null
})

// Upprepning (r). Editorn visas i raden för det mål den öppnades för.
const recurrenceTargetId = ref<number | null>(null)

function openRecurrenceEditor(goal: PrioritizedGoalRow) {
  recurrenceTargetId.value = goal.id
}

function closeRecurrenceEditor() {
  recurrenceTargetId.value = null
}

async function saveRecurrence(rule: string, mode: RecurrenceMode) {
  const goal = visibleGoals.value.find(g => g.id === recurrenceTargetId.value)
  if (!goal) return
  try {
    const previousRule = goal.recurrence_rule
    const previousMode = goal.recurrence_mode
    await setGoalRecurrence(goal.id, rule, mode)
    undoRecurrence(goal, previousRule, previousMode)
    await loadPrioritizedGoals()
    closeRecurrenceEditor()
  } catch (err) {
    console.error('Failed to save goal recurrence:', err)
  }
}

async function removeRecurrence() {
  const goal = visibleGoals.value.find(g => g.id === recurrenceTargetId.value)
  if (!goal) return
  try {
    const previousRule = goal.recurrence_rule
    const previousMode = goal.recurrence_mode
    await setGoalRecurrence(goal.id, null)
    undoRecurrence(goal, previousRule, previousMode)
    await loadPrioritizedGoals()
    closeRecurrenceEditor()
  } catch (err) {
    console.error('Failed to remove goal recurrence:', err)
  }
}

// Deadline-väljare inline i raden. Datumet sparas direkt vid val och
// tom sträng rensar det.
const dueGoalId = ref<number | null>(null)
const dueDraft = ref('')
const dueInputRef = ref<HTMLInputElement | null>(null)

function openDuePicker(goal: PrioritizedGoalRow) {
  dueDraft.value = goal.due ?? ''
  dueGoalId.value = goal.id
  nextTick(() => {
    const input = document.querySelector(
      `[data-due-input="${goal.id}"]`,
    ) as (HTMLInputElement & { showPicker?: () => void }) | null

    input?.focus()
    // Öppna kalender-popupen direkt när c eller kalenderknappen används.
    try {
      input?.showPicker?.()
    } catch {
      // Fokus fungerar som fallback i webbläsare som nekar showPicker().
    }
  })
}

async function saveDue(goal: PrioritizedGoalRow, value: string) {
  const due = value ? value : null
  try {
    const previousDue = goal.due
    await setGoalDue(goal.id, due)
    undoDue(goal, previousDue)
    dueGoalId.value = null
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to save due date:', err)
  }
}

// Startdatum-väljare inline i raden. Datumet anger från vilket datum
// målet får dyka upp i prioriteringslistan - tom sträng rensar det.
const startGoalId = ref<number | null>(null)
const startDraft = ref('')

function openStartPicker(goal: PrioritizedGoalRow) {
  startDraft.value = goal.start_date ?? ''
  startGoalId.value = goal.id
  nextTick(() => {
    const input = document.querySelector(
      `[data-start-input="${goal.id}"]`,
    ) as (HTMLInputElement & { showPicker?: () => void }) | null

    input?.focus()
    try {
      input?.showPicker?.()
    } catch {
      // Fokus fungerar som fallback i webbläsare som nekar showPicker().
    }
  })
}

async function saveStartDate(goal: PrioritizedGoalRow, value: string) {
  const startDate = value ? value : null
  try {
    const previousStartDate = goal.start_date ?? null
    await setGoalStartDate(goal.id, startDate)
    undoStartDate(goal, previousStartDate)
    startGoalId.value = null
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to save start date:', err)
  }
}

// Uppskattad tidsåtgång inline i raden.
const estimateGoalId = ref<number | null>(null)

function openEstimateEditor(goal: PrioritizedGoalRow) {
  estimateGoalId.value = goal.id
}

function closeEstimateEditor() {
  estimateGoalId.value = null
}

async function saveEstimate(goal: PrioritizedGoalRow, minutes: number | null) {
  try {
    const previous = goal.estimated_minutes ?? null
    await setGoalEstimate(goal.id, minutes)
    undoEstimate(goal, previous)
    estimateGoalId.value = null
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to save estimate:', err)
  }
}
// Beskrivningsredigering inline i raden (Ctrl+Enter sparar, Esc avbryter)
const editingDescriptionGoalId = ref<number | null>(null)
const descriptionDraft = ref('')

function startEditDescription(goal: PrioritizedGoalRow) {
  editingDescriptionGoalId.value = goal.id
  descriptionDraft.value = goal.description ?? ''
  nextTick(() => {
    const el = document.querySelector('#priority-description-input') as HTMLTextAreaElement | null
    el?.focus()
  })
}

function cancelEditDescription() {
  editingDescriptionGoalId.value = null
  descriptionDraft.value = ''
}

function handleDescriptionKeydown(event: KeyboardEvent, goal: PrioritizedGoalRow) {
  if (event.key !== 'Enter' || (!event.ctrlKey && !event.metaKey)) return

  event.preventDefault()
  event.stopPropagation()
  void saveDescription(goal)
}

async function saveDescription(goal: PrioritizedGoalRow) {
  const description = descriptionDraft.value
  try {
    await updateGoalDescription(goal.id, description)
    cancelEditDescription()
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to save goal description:', err)
  }
}

// Bekräftelse innan ett mål tas bort
const showDeleteConfirm = ref(false)
const goalToDelete = ref<PrioritizedGoalRow | null>(null)

async function handleDelete(goal: PrioritizedGoalRow) {
  try {
    await deleteGoal(goal.id)
    showDeleteConfirm.value = false
    goalToDelete.value = null
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to delete goal:', err)
  }
}

function getWeightColor(weight: number): string {
  if (weight <= 2) return 'text-ink-muted'
  if (weight <= 9) return 'text-ink-dim'
  if (weight <= 14) return 'text-ink-dim'
  return 'text-accent'
}

function getWeightBgColor(weight: number): string {
  if (weight <= 2) return 'bg-elevated/30'
  if (weight <= 9) return 'bg-elevated/50'
  if (weight <= 14) return 'bg-elevated/70'
  return 'bg-accent/30'
}

const filterOptions = [
  { value: 'now', label: 'Nu' },
  { value: 'all', label: 'Alla' },
  { value: 'started', label: 'Påbörjade' },
  { value: 'not_started', label: 'Ej påbörjade' },
] as const
</script>

<template>
  <div v-if="user" class="max-w-4xl mx-auto px-6 py-8">
    <div class="flex flex-wrap items-start justify-between gap-4 mb-6">
      <div class="flex items-center gap-3">
        <h1 class="text-3xl font-bold tracking-tight text-ink">Gör nu</h1>
        <span
          v-if="isPriorityMode"
          class="px-3 py-1 rounded-full text-xs bg-fuchsia-600/20 text-fuchsia-300 border border-fuchsia-500/30"
        >
          Ordna framsteg (håll P + K/J)
        </span>
      </div>
      <NuxtLink to="/" class="rounded-lg border border-line bg-surface-2 px-4 py-2 text-ink-muted transition-colors hover:border-line-strong hover:text-ink inline-flex items-center gap-2">
        <Icon name="lucide:arrow-left" class="w-4 h-4" />
        Tillbaka
      </NuxtLink>
    </div>

    <div class="text-xs text-ink-dim mb-6">
      <span class="text-ink-muted">j/k</span> navigera &nbsp;|&nbsp;
      <span class="text-ink-muted">↑/↓</span> sortera &nbsp;|&nbsp;
      <span class="text-ink-muted">l</span> öppna mål &nbsp;|&nbsp;
      <span class="text-ink-muted">Håll Shift+P</span> + <span class="text-ink-muted">K/J</span> ändra vikt &nbsp;|&nbsp;
      <span class="text-ink-muted">s</span> påbörjad &nbsp;|&nbsp;
      <span class="text-ink-muted">Space+N</span> nu &nbsp;|&nbsp;
      <span class="text-ink-muted">d</span> klar &nbsp;|&nbsp;
      <span class="text-ink-muted">e</span> evigt &nbsp;|&nbsp;
      <span class="text-ink-muted">c</span> deadline &nbsp;|&nbsp;
      <span class="text-ink-muted">b</span> startdatum &nbsp;|&nbsp;
      <span class="text-ink-muted">t</span> beskrivning &nbsp;|&nbsp;
      <span class="text-ink-muted">x</span> ta bort
    </div>

    <div class="flex flex-wrap items-center gap-2 mb-3">
      <span class="text-xs text-ink-dim">Jag är här:</span>
      <select
        :value="currentPlaceId ?? ''"
        class="px-2 py-1.5 rounded-lg bg-surface-2 border border-line text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
        @change="selectCurrentPlace(($event.target as HTMLSelectElement).value ? Number(($event.target as HTMLSelectElement).value) : null)"
      >
        <option value="">Ingen plats vald</option>
        <option v-for="place in allPlaces" :key="place.id" :value="place.id">
          {{ place.name }}
        </option>
      </select>
      <button
        class="px-3 py-1.5 rounded-full text-sm transition-colors border border-line bg-surface-2 text-ink-muted hover:text-ink flex items-center gap-1.5"
        :disabled="locating"
        @click="locateMe"
      >
        <Icon name="lucide:locate-fixed" class="w-4 h-4" />
        {{ locating ? 'Söker…' : 'Var jag befinner mig' }}
      </button>
      <button
        class="px-3 py-1.5 rounded-full text-sm transition-colors border"
        :class="filterByCurrentPlace
          ? 'bg-accent-strong text-white border-accent-strong'
          : 'bg-surface-2 text-ink-muted border-line hover:border-line-strong hover:text-ink'"
        @click="setFilterByCurrentPlace(!filterByCurrentPlace)"
      >
        Görbar här
      </button>
    </div>

    <p v-if="gpsMessage" class="text-xs text-ink-dim mb-3">{{ gpsMessage }}</p>

    <p
      v-if="filterByCurrentPlace && currentPlaceId === null"
      class="text-xs text-ink-dim mb-3"
    >
      Välj en plats för att se vilka mål som är görbara där.
    </p>

    <div class="flex flex-wrap items-center gap-2 mb-6">
      <button
        v-for="option in filterOptions"
        :key="option.value"
        type="button"
        @click="setNowFilter(option.value)"
        class="px-3 py-1.5 rounded-full text-sm transition-colors border"
        :class="
          nowFilter === option.value
            ? 'bg-accent-strong text-white border-accent-strong'
            : 'bg-surface-2 text-ink-muted border-line hover:border-line-strong hover:text-ink'
        "
      >
        {{ option.label }}
      </button>
      <span class="text-xs text-ink-dim ml-2">
        {{ visibleGoals.length }} av {{ prioritizedGoals.length }} mål visas
      </span>
    </div>

    <div v-if="isLoading" class="text-center py-12">
      <p class="text-ink-muted">Laddar...</p>
    </div>

    <div v-else-if="prioritizedGoals.length === 0" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-10 text-center">
      <p class="text-ink-muted">Inga mål är möjliga att göra nu</p>
    </div>

    <div v-else-if="visibleGoals.length === 0" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-10 text-center">
      <p class="text-ink-muted">
        Inga mål matchar filtret &quot;{{
          filterOptions.find((o) => o.value === nowFilter)?.label
        }}&quot;
      </p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="(goal, index) in visibleGoals"
        :key="goal.id"
        :data-index="index"
        :data-goal-id="goal.id"
      >
        <div
          data-goal-row
          draggable="true"
          class="rounded-2xl border border-line transition-colors hover:border-line-strong flex items-center gap-4 px-4 py-3.5 cursor-pointer opt-row"
          :class="[
            getWeightBgColor(goal.weight),
            selectedGoalId === goal.id ? 'ring-1 ring-accent' : '',
            isPriorityMode && selectedGoalId === goal.id ? 'ring-1 ring-fuchsia-500' : '',
            isWaitingForNextOccurrence(goal) ? 'opacity-50' : ''
          ]"
          @dragstart="startDrag($event, goal)"
          @click="selectGoal(goal.id)"
        >
          <div class="flex-shrink-0 w-8 text-center">
            <span class="text-lg font-bold tabular-nums" :class="getWeightColor(goal.weight)">
              {{ index + 1 }}
            </span>
          </div>

          <div class="flex items-center gap-2 flex-1 min-w-0">
            <Icon :name="goal.icon || 'heroicons:star'" class="w-5 h-5 flex-shrink-0 text-ink-muted"
              :class="goal.eternal ? ETERNAL_TEXT : ''" />
            <NuxtLink
              :to="`/goal/${goal.id}`"
              class="truncate transition-colors"
              :class="goal.eternal ? ETERNAL_TEXT + ' ' + ETERNAL_TEXT_HOVER : 'text-ink hover:text-white'"
              @click.stop
            >
              {{ goal.title }}
            </NuxtLink>
            <!-- Antalet undermål målet har. -->
            <ChildCountBadge :goal-id="goal.id" />
            <!-- Föräldrarna målet ligger under, som klickbara länkar. -->
            <div
              v-if="goal.parents?.length"
              class="flex items-center gap-1 min-w-0 flex-shrink-0"
              :title="`Förälder: ${goal.parents.map((p) => p.title).join(', ')}`"
            >
              <Icon name="lucide:corner-left-up" class="w-3.5 h-3.5 text-ink-dim flex-shrink-0" />
              <template v-for="(parent, parentIndex) in goal.parents" :key="parent.id">
                <span v-if="parentIndex > 0" class="text-xs text-ink-dim">·</span>
                <NuxtLink
                  :to="`/goal/${parent.id}`"
                  class="text-xs text-ink-dim hover:text-ink-muted transition-colors truncate max-w-[10rem]"
                  @click.stop
                >
                  {{ parent.title }}
                </NuxtLink>
              </template>
            </div>
            <Icon
              v-if="goal.places?.length"
              name="lucide:map-pin"
              class="w-4 h-4 flex-shrink-0 text-accent"
              :title="`Platser: ${goal.places.map((p) => p.name).join(', ')}`"
            />
          </div>
          <!-- Ikonerna grupperas tätt ihop så att raden inte ser utsträckt ut. -->
          <div class="flex items-center gap-1 flex-shrink-0">
          <span class="opt-icon" :class="{ 'is-on': !!goal.started }">
            <button
              @click.stop="toggleStarted(goal)"
              class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
              :class="goal.started ? 'text-yellow-400' : 'text-ink-dim hover:text-ink-muted'"
              :title="goal.started ? 'Avbörja (s)' : 'Påbörja (s)'"
            >
              <Icon
                name="lucide:circle-play"
                class="w-5 h-5"
                :style="{ opacity: goal.started ? 1 : 0.25 }"
              />
            </button>
          </span>

          <span class="opt-icon" :class="{ 'is-on': !!goal.agent }">
            <AgentBadge :active="!!goal.agent" @toggle="toggleAgent(goal)" />
          </span>

          <span class="opt-icon" :class="{ 'is-on': !!goal.now }">
            <NowBadge :active="!!goal.now" @toggle="toggleNow(goal)" />
          </span>

          <span class="opt-icon" :class="{ 'is-on': !!goal.due || dueGoalId === goal.id }">
            <template v-if="dueGoalId === goal.id">
              <input
                ref="dueInputRef"
                :data-due-input="goal.id"
                v-model="dueDraft"
                type="date"
                class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
                @click.stop
                @change="saveDue(goal, dueDraft)"
                @keydown.enter.prevent="saveDue(goal, dueDraft)"
                @keydown.esc.prevent="dueGoalId = null"
              />
              <button
                @click.stop="saveDue(goal, '')"
                class="rounded p-1 text-ink-dim transition-colors hover:text-ink-muted"
                title="Rensa deadline"
              >
                <Icon name="lucide:x" class="w-4 h-4" />
              </button>
            </template>
            <button
              v-else
              @click.stop="openDuePicker(goal)"
              class="rounded p-1 transition-colors"
              :class="goal.due ? dueToneClass(goal.due) : 'text-ink-dim hover:text-ink-muted'"
              :title="goal.due ? `Deadline: ${formatDueDate(goal.due)}` : 'Sätt deadline (c)'"
            >
              <Icon name="lucide:calendar-clock" class="w-5 h-5" :style="{ opacity: goal.due ? 1 : 0.25 }" />
            </button>
          </span>

          <span class="opt-icon" :class="{ 'is-on': !!goal.start_date || startGoalId === goal.id }">
            <template v-if="startGoalId === goal.id">
              <input
                :data-start-input="goal.id"
                v-model="startDraft"
                type="date"
                class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
                @click.stop
                @change="saveStartDate(goal, startDraft)"
                @keydown.enter.prevent="saveStartDate(goal, startDraft)"
                @keydown.esc.prevent="startGoalId = null"
              />
              <button @click.stop="saveStartDate(goal, '')"
                class="rounded p-1 text-ink-dim transition-colors hover:text-ink-muted"
                title="Rensa startdatum">
                <Icon name="lucide:x" class="w-4 h-4" />
              </button>
            </template>
            <button
              v-else
              @click.stop="openStartPicker(goal)"
              class="rounded p-1 transition-colors"
              :class="goal.start_date ? (isFutureDate(goal.start_date) ? 'text-sky-300' : 'text-ink-muted') : 'text-ink-dim hover:text-ink-muted'"
              :title="goal.start_date ? `Startdatum: ${formatDueDate(goal.start_date)} - målet syns i listan från detta datum` : 'Sätt startdatum (b)'"
            >
              <Icon name="lucide:calendar-days" class="w-5 h-5" :style="{ opacity: goal.start_date ? 1 : 0.25 }" />
            </button>
          </span>

          <span class="opt-icon" :class="{ 'is-on': !!goal.estimated_minutes || estimateGoalId === goal.id }">
            <EstimateBadge
              :goal="goal"
              size="sm"
              :editor-open="estimateGoalId === goal.id"
              @open="openEstimateEditor(goal)"
              @save="saveEstimate(goal, $event)"
              @cancel="closeEstimateEditor" />
          </span>
          <div v-if="isRecurringGoal(goal) || recurrenceTargetId === goal.id || goal.eternal || goal.due || goal.start_date"
            class="relative flex items-center gap-2 flex-shrink-0">
            <button v-if="isRecurringGoal(goal)" @click.stop="openRecurrenceEditor(goal)"
              class="flex items-center gap-1 text-xs tabular-nums text-ink-dim transition-colors hover:text-ink-muted"
              :title="`${describeGoalRecurrence(goal.recurrence_rule)} - ${recurrenceCount(goal)} slutförda (r)`">
              <Icon name="lucide:repeat" class="w-3.5 h-3.5" />
              {{ recurrenceCount(goal) }}
            </button>
            <div v-if="recurrenceTargetId === goal.id"
              class="absolute right-0 top-full z-30 mt-1" @click.stop @keydown.stop>
              <RecurrenceEditor :rule="goal.recurrence_rule ?? null"
                :mode="goal.recurrence_mode ?? 'schedule'"
                @save="saveRecurrence"
                @remove="removeRecurrence"
                @cancel="closeRecurrenceEditor" />
            </div>
            <span v-if="goal.eternal" :class="ETERNAL_ACCENT + ' text-xs flex-shrink-0'">∞</span>
            <span v-if="goal.due" class="text-xs flex-shrink-0 tabular-nums" :class="dueToneClass(goal.due)">
              📅 {{ formatDueDate(goal.due) }}
            </span>
            <span v-if="goal.start_date" class="text-xs flex-shrink-0 tabular-nums"
              :class="isFutureDate(goal.start_date) ? 'text-sky-300' : 'text-ink-dim'">
              🗓 {{ formatDueDate(goal.start_date) }}
            </span>
          </div>
          </div>

          <div class="flex-shrink-0 text-right">
            <div
              class="text-sm font-mono px-3 py-1 rounded-full min-w-[3rem] text-center tabular-nums"
              :class="getWeightBgColor(goal.weight) + ' ' + getWeightColor(goal.weight)"
            >
              {{ goal.weight }}
            </div>
          </div>
        </div>

        <!-- Beskrivningsredigering inline (t) -->
        <div
          v-if="editingDescriptionGoalId === goal.id"
          class="mt-1 ml-12 mr-12 mb-2"
        >
          <textarea
            id="priority-description-input"
            v-model="descriptionDraft"
            rows="3"
            placeholder="Beskrivning..."
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
            @keydown="handleDescriptionKeydown($event, goal)"
            @keydown.esc.prevent="cancelEditDescription"
          ></textarea>
          <div class="flex items-center gap-2 mt-1">
            <button
              @click="saveDescription(goal)"
              class="px-2 py-1 rounded text-xs bg-accent-strong text-white"
            >
              Spara
            </button>
            <button
              @click="cancelEditDescription"
              class="px-2 py-1 rounded text-xs text-ink-dim hover:text-ink-muted"
            >
              Avbryt
            </button>
            <span class="text-xs text-ink-dim">Ctrl+Enter sparar</span>
          </div>
        </div>
      </div>

    </div>
  </div>

  <div v-else class="flex items-center justify-center min-h-[50vh]">
    <div class="text-center">
      <h1 class="text-3xl font-bold tracking-tight text-ink mb-6">Gör nu</h1>
      <NuxtLink to="/login">
        <UButton size="lg" icon="material-symbols:login-rounded">
          Logga in
        </UButton>
      </NuxtLink>
    </div>
  </div>

  <!-- Bekräftelse innan ett mål tas bort -->
  <div
    v-if="showDeleteConfirm && goalToDelete"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm"
    @click.self="showDeleteConfirm = false"
  >
    <div class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-6 max-w-md w-full mx-4 shadow-2xl">
      <h3 class="text-lg font-semibold mb-2 text-ink">Ta bort mål?</h3>
      <p class="text-ink-muted mb-4">
        Är du säker på att du vill ta bort "{{ goalToDelete.title }}"?
        Detta kan inte ångras.
      </p>
      <div class="flex gap-3">
        <button
          @click="showDeleteConfirm = false"
          class="rounded-lg border border-line bg-surface-2 px-4 py-2 text-ink-muted transition-colors hover:border-line-strong hover:text-ink"
        >
          Avbryt
        </button>
        <button
          @click="handleDelete(goalToDelete)"
          class="px-4 py-2 bg-red-600 hover:bg-red-500 rounded-lg text-white transition-colors"
        >
          Ta bort
        </button>
      </div>
      <p class="text-xs text-ink-dim mt-4">
        Tryck Escape för att avbryta
      </p>
    </div>
  </div>
</template>

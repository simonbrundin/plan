<script setup lang="ts">
import type { Goal } from '~/types/goal'
import { AuthenticationError } from '~/composables/useGoalApi'
import { canBeFinished } from '~/utils/goalRules'
import { isRootGoal, isUnfinishedGoal, matchesStartedFilter } from '~/utils/goalFilters'
import { moveChildId } from '~/utils/childOrder'
import { sortGoalsByOrder } from '~/utils/goalOrder'
import { descriptionPreview } from '~/utils/goalDescription'
import { dueTone, formatDueDate, todayDateString } from '~/utils/dueDate'
import { describeGoalRecurrence, isRecurringGoal, isWaitingForNextOccurrence, recurrenceCount } from '~/utils/recurrence'
import type { RecurrenceMode } from '~/utils/recurrence'
import { ETERNAL_ACCENT, ETERNAL_ACCENT_HOVER, ETERNAL_TEXT } from '~/utils/eternalStyle'

const router = useRouter()
const { toggleGoalFinished, toggleGoalStarted, setGoalEternal, setGoalDue, setGoalRecurrence, deleteGoal, loadAllGoals, createGoal, reorderGoals, updateGoalDescription, completeGoalOccurrence } = useGoalApi()
const { revision: quickAddRevision } = useQuickAdd()

const goals = ref<Goal[]>([])
const pending = ref(false)
const error = ref<Error | null>(null)
const authError = ref(false)

onMounted(async () => {
  pending.value = true
  try {
    goals.value = await loadAllGoals()
    authError.value = false
  } catch (e) {
    if (e instanceof AuthenticationError) {
      authError.value = true
      // Redirect to login after a short delay
      setTimeout(() => {
        router.push('/login')
      }, 2000)
    } else {
      error.value = e as Error
    }
    console.error('Failed to load goals:', e)
  } finally {
    pending.value = false
  }
})

const selectedIndex = ref(0)
const mode = ref<'normal' | 'insert'>('normal')

// Lokal state för att uppdatera UI direkt
const localGoals = computed(() => goals.value)

// Grundmål (utan förälder) i sin sparade ordning. Målen som saknar order
// hamnar sist i gruppen, med nyast först - samma ordning som API:t gav
// innan orderfältet fanns.
const rootGoals = computed(() =>
  sortGoalsByOrder(localGoals.value.filter(isRootGoal))
)

// Endast grundmål (utan förälder) visas under /goals - undermål nås
// via sitt föräldermål. Slutförda mål döljs helt.
const displayedGoals = computed(() =>
  rootGoals.value.filter(isUnfinishedGoal)
)
const showStarted = ref(false)

const filteredGoals = computed(() => {
  let result = displayedGoals.value
  if (showStarted.value) {
    result = result.filter(g => matchesStartedFilter(g, 'started'))
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(g => g.title.toLowerCase().includes(query))
  }
  return result
})

// Flytta det markerade grundmålet upp eller ner i listan.
//
// Ordningen uppdateras lokalt först (optimistiskt) så att tangenttryck
// direkt ger respons, och sparas sedan till servern. Sparandet serialiseras
// och slås ihop, så att snabba tangenttryck inte skriver över varandra.
function moveGoal(direction: 'up' | 'down') {
  const visible = filteredGoals.value
  const index = selectedIndex.value
  if (index < 0 || index >= visible.length) return

  const moving = visible[index]
  if (!moving) return

  const reordered = moveChildId(
    rootGoals.value.map(g => g.id),
    visible.map(g => g.id),
    moving.id,
    direction,
  )
  if (!reordered) return

  applyLocalOrder(reordered)
  selectedIndex.value = direction === 'up' ? index - 1 : index + 1

  persistGoalOrder()
}

// Sätter den nya ordningen lokalt utan att vänta på servern
function applyLocalOrder(ids: number[]) {
  const byId = new Map(goals.value.map(g => [g.id, g]))
  ids.forEach((id, index) => {
    const goal = byId.get(id)
    if (goal) {
      goal.order = index
    }
  })
}

// Sparar ordningen till servern. Anropen körs i kedjan och senaste ordningen
// vinner - flera snabba tangenttryck slås ihop till ett anrop.
let orderSaveChain: Promise<void> = Promise.resolve()
let pendingOrder: number[] | null = null
let isSavingOrder = false

function persistGoalOrder() {
  pendingOrder = rootGoals.value.map(g => g.id)

  if (isSavingOrder) {
    return
  }

  isSavingOrder = true
  orderSaveChain = orderSaveChain.then(flushGoalOrder)
}

async function flushGoalOrder() {
  try {
    while (pendingOrder) {
      const ids = pendingOrder
      pendingOrder = null

      try {
        await reorderGoals(ids)
      } catch (err) {
        console.error('Failed to save goal order:', err)
        await refreshGoals()
      }
    }
  } finally {
    isSavingOrder = false
  }
}

async function refreshGoals() {
  try {
    goals.value = await loadAllGoals()
  } catch (err) {
    console.error('Failed to reload goals:', err)
  }
}

// Ett mål som skapats via det globala snabbfältet ska synas direkt.
watch(quickAddRevision, refreshGoals)

// Navigera till mål
function goToGoal(goalId: number) {
  router.push(`/goal/${goalId}`)
}

// Skapa nytt grundmål (mål utan förälder)
const showNewGoalInput = ref(false)
const newGoalTitle = ref('')
const newGoalInput = ref<HTMLInputElement | null>(null)

function startNewGoal() {
  showNewGoalInput.value = true
  newGoalTitle.value = ''
}

function cancelNewGoal() {
  showNewGoalInput.value = false
  newGoalTitle.value = ''
}

async function saveNewGoal() {
  const title = newGoalTitle.value.trim()
  if (!title) {
    cancelNewGoal()
    return
  }

  try {
    const newGoal = await createGoal(title)
    goals.value.push(newGoal)
    showNewGoalInput.value = false
    newGoalTitle.value = ''

    // Markera det ny skapade målet
    await nextTick()
    const idx = filteredGoals.value.findIndex(g => g.id === newGoal.id)
    if (idx !== -1) {
      selectedIndex.value = idx
    }
  } catch (err) {
    console.error('Failed to create goal:', err)
  }
}

watch(showNewGoalInput, async (val) => {
  if (val) {
    await nextTick()
    newGoalInput.value?.focus()
  }
})

// Öppna sök
const showSearch = ref(false)
const searchQuery = ref('')

// Delete confirmation
const showDeleteConfirm = ref(false)
const goalToDelete = ref<Goal | null>(null)

// Toggle finished
async function toggleFinished(goal: Goal) {
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
    await completeOccurrence(goal)
    return
  }

  try {
    const newFinishedValue = goal.finished ? null : new Date().toISOString()
    await toggleGoalFinished(goal.id, newFinishedValue)
    
    // Uppdatera lokal state
    const idx = localGoals.value.findIndex(g => g.id === goal.id)
    if (idx !== -1) {
      localGoals.value[idx] = { ...localGoals.value[idx], finished: newFinishedValue }
    }
  } catch (err) {
    console.error('Failed to toggle finished:', err)
  }
}

// Slutför den aktuella förekomsten av ett återkommande mål.
async function completeOccurrence(goal: Goal) {
  try {
    const updated = await completeGoalOccurrence(goal.id, goal.recurrence_next ?? undefined, todayDateString())
    const idx = localGoals.value.findIndex(g => g.id === goal.id)
    const current = localGoals.value[idx]
    if (idx !== -1 && current) {
      localGoals.value[idx] = {
        ...current,
        recurrence_rule: updated.recurrence_rule,
        recurrence_mode: updated.recurrence_mode,
        recurrence_next: updated.recurrence_next,
        recurrence_count: updated.recurrence_count,
        finished: updated.finished,
      }
    }
  } catch (err) {
    console.error('Failed to complete occurrence:', err)
  }
}

// Upprepning (r). Editorn visas i listan för det mål den öppnades för.
const recurrenceTargetId = ref<number | null>(null)

function openRecurrenceEditor(goal: Goal) {
  recurrenceTargetId.value = goal.id
}

function closeRecurrenceEditor() {
  recurrenceTargetId.value = null
}

function updateLocalRecurrence(goal: Goal) {
  const idx = localGoals.value.findIndex(g => g.id === goal.id)
  const current = localGoals.value[idx]
  if (idx === -1 || !current) return
  localGoals.value[idx] = {
    ...current,
    recurrence_rule: goal.recurrence_rule,
    recurrence_mode: goal.recurrence_mode,
    recurrence_next: goal.recurrence_next,
    recurrence_count: goal.recurrence_count,
    finished: goal.finished,
  }
}

async function saveRecurrence(rule: string, mode: RecurrenceMode) {
  const goal = localGoals.value.find(g => g.id === recurrenceTargetId.value)
  if (!goal) return
  try {
    updateLocalRecurrence(await setGoalRecurrence(goal.id, rule, mode))
    closeRecurrenceEditor()
  } catch (err) {
    console.error('Failed to save goal recurrence:', err)
  }
}

async function removeRecurrence() {
  const goal = localGoals.value.find(g => g.id === recurrenceTargetId.value)
  if (!goal) return
  try {
    updateLocalRecurrence(await setGoalRecurrence(goal.id, null))
    closeRecurrenceEditor()
  } catch (err) {
    console.error('Failed to remove goal recurrence:', err)
  }
}

// Toggle started
async function toggleStarted(goal: Goal) {
  try {
    const newStartedValue = goal.started ? null : new Date().toISOString()
    await toggleGoalStarted(goal.id, newStartedValue)
    const idx = localGoals.value.findIndex(g => g.id === goal.id)
    if (idx !== -1) {
      localGoals.value[idx] = { ...localGoals.value[idx], started: newStartedValue }
    }
  } catch (err) {
    console.error('Failed to toggle started:', err)
  }
}

// Due-datum (deadline) för ett grundmål. Datumet är ett kalenderdatum
// ("YYYY-MM-DD") - tom sträng rensar det.
const dueGoalId = ref<number | null>(null)
const dueDraft = ref('')
const dueInputRef = ref<HTMLInputElement | null>(null)

function dueToneClass(due: string | null): string {
  const tone = dueTone(due)
  if (tone === 'overdue') return 'text-red-400'
  if (tone === 'today') return 'text-amber-400'
  return 'text-ink-muted'
}

function openDuePicker(goal: Goal) {
  dueDraft.value = goal.due ?? ''
  dueGoalId.value = goal.id
  nextTick(() => dueInputRef.value?.focus())
}

async function saveDue(goal: Goal, value: string) {
  const due = value ? value : null
  try {
    await setGoalDue(goal.id, due)
    const idx = localGoals.value.findIndex(g => g.id === goal.id)
    const current = localGoals.value[idx]
    if (idx !== -1 && current) {
      localGoals.value[idx] = { ...current, due }
    }
    dueGoalId.value = null
  } catch (err) {
    console.error('Failed to save due date:', err)
  }
}

// Toggle evigt-märke
async function toggleEternal(goal: Goal) {
  const newEternal = !goal.eternal
  try {
    await setGoalEternal(goal.id, newEternal)
    const idx = localGoals.value.findIndex(g => g.id === goal.id)
    if (idx !== -1) {
      localGoals.value[idx] = { ...localGoals.value[idx], eternal: newEternal }
    }
  } catch (err) {
    console.error('Failed to toggle eternal:', err)
  }
}

// Delete goal
async function handleDelete(goal: Goal) {
  try {
    await deleteGoal(goal.id)
    // Ta bort från lokal state
    localGoals.value = localGoals.value.filter(g => g.id !== goal.id)
    // Justera selectedIndex om det behövs
    if (selectedIndex.value >= localGoals.value.length) {
      selectedIndex.value = Math.max(0, localGoals.value.length - 1)
    }
    showDeleteConfirm.value = false
    goalToDelete.value = null
  } catch (err) {
    console.error('Failed to delete goal:', err)
  }
}

// Redigera beskrivning för det markerade målet
const editingDescriptionGoalId = ref<number | null>(null)
const descriptionDraft = ref('')

function startEditDescription(goal: Goal) {
  editingDescriptionGoalId.value = goal.id
  descriptionDraft.value = goal.description ?? ''
  nextTick(() => {
    const el = document.querySelector('#goal-description-input') as HTMLTextAreaElement | null
    el?.focus()
  })
}

function cancelEditDescription() {
  editingDescriptionGoalId.value = null
  descriptionDraft.value = ''
}

function handleDescriptionKeydown(event: KeyboardEvent, goal: Goal) {
  if (event.key !== 'Enter' || (!event.ctrlKey && !event.metaKey)) return

  event.preventDefault()
  event.stopPropagation()
  void saveDescription(goal)
}

async function saveDescription(goal: Goal) {
  const description = descriptionDraft.value
  try {
    await updateGoalDescription(goal.id, description)
    const idx = localGoals.value.findIndex(g => g.id === goal.id)
    if (idx !== -1) {
      localGoals.value[idx] = { ...localGoals.value[idx], description }
    }
    cancelEditDescription()
  } catch (err) {
    console.error('Failed to save goal description:', err)
  }
}

// Vim tangentbordshantering
function handleKeydown(event: KeyboardEvent) {
  const target = event.target as HTMLElement
  const isInInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA'
  
  // ESC för att stänga delete confirmation
  if (event.key === 'Escape' && showDeleteConfirm.value) {
    showDeleteConfirm.value = false
    goalToDelete.value = null
    return
  }

  if (isInInput && !showSearch.value && !showDeleteConfirm.value) return
  if (isInInput) return

  // ESC för att stänga sök
  if (event.key === 'Escape' && showSearch.value) {
    showSearch.value = false
    searchQuery.value = ''
    return
  }

  // /
  if (event.key === '/') {
    event.preventDefault()
    showSearch.value = true
    return
  }

  // h - tillbaka till inbox
  if (event.key === 'h') {
    event.preventDefault()
    router.push('/')
    return
  }

  // d - toggle finished
  if (event.key === 'd') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      toggleFinished(filteredGoals.value[selectedIndex.value])
    }
    return
  }

  // e - växla evigt-märke
  if (event.key === 'e') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      toggleEternal(filteredGoals.value[selectedIndex.value])
    }
    return
  }

  // t - redigera beskrivning för det markerade målet
  if (event.key === 't') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      startEditDescription(filteredGoals.value[selectedIndex.value])
    }
    return
  }

  // s - toggle started
  if (event.key === 's') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      toggleStarted(filteredGoals.value[selectedIndex.value])
    }
    return
  }

  // c - sätt deadline (due-datum)
  if (event.key === 'c') {
    event.preventDefault()
    const goal = filteredGoals.value[selectedIndex.value]
    if (goal) {
      openDuePicker(goal)
    }
    return
  }

  // r - sätt upprepning för det markerade målet
  if (event.key === 'r') {
    event.preventDefault()
    const goal = filteredGoals.value[selectedIndex.value]
    if (goal) {
      openRecurrenceEditor(goal)
    }
    return
  }

  // x - delete
  if (event.key === 'x') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      goalToDelete.value = filteredGoals.value[selectedIndex.value]
      showDeleteConfirm.value = true
    }
    return
  }

  // j - nästa mål
  if (event.key === 'j') {
    event.preventDefault()
    const max = filteredGoals.value.length - 1
    selectedIndex.value = Math.min(selectedIndex.value + 1, max)
    return
  }

  // k - föregående mål
  if (event.key === 'k') {
    event.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
    return
  }

  // ArrowUp - flytta markerat grundmål upp
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    moveGoal('up')
    return
  }

  // ArrowDown - flytta markerat grundmål ner
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    moveGoal('down')
    return
  }

  // Enter - lägg till ett nytt grundmål
  if (event.key === 'Enter') {
    event.preventDefault()
    startNewGoal()
    return
  }

  // l - öppna markerat mål
  if (event.key === 'l') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      goToGoal(filteredGoals.value[selectedIndex.value].id)
    }
    return
  }

  // g - till första
  if (event.key === 'g' && !event.shiftKey) {
    event.preventDefault()
    selectedIndex.value = 0
    return
  }

  // G - till sista
  if (event.key === 'G') {
    event.preventDefault()
    selectedIndex.value = Math.max(0, filteredGoals.value.length - 1)
    return
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

// Återställ selectedIndex när mål ändras
watch(filteredGoals, () => {
  if (selectedIndex.value >= filteredGoals.value.length) {
    selectedIndex.value = Math.max(0, filteredGoals.value.length - 1)
  }
})

// Scrolla till markerat mål när selectedIndex ändras
watch(selectedIndex, async () => {
  await nextTick()
  const selectedElement = document.querySelector(`[data-goal-index="${selectedIndex.value}"]`) as HTMLElement
  if (selectedElement) {
    selectedElement.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest'
    })
  }
})

// Fokusera sök-input när sök öppnas
const searchInput = ref<HTMLInputElement | null>(null)
watch(showSearch, async (val) => {
  if (val) {
    await nextTick()
    searchInput.value?.focus()
  }
})

definePageMeta({
  title: 'Mina mål',
  ssr: false, // Disable SSR to ensure session is hydrated before API calls
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-6 py-8">
    <div class="flex flex-wrap items-start justify-between gap-4 mb-6">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-ink">Mina mål</h1>
        <p class="text-sm text-ink-dim mt-1">
          {{ filteredGoals.length }} av {{ displayedGoals.length }} grundmål
        </p>
      </div>
      <div class="flex items-center gap-2">
        <!-- Nytt grundmål -->
        <button @click="startNewGoal" class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink" title="Nytt grundmål (Enter)">
          <Icon name="lucide:plus" class="w-5 h-5" />
        </button>
        <!-- Visa endast påbörjade -->
        <button @click="showStarted = !showStarted"
          class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
          :class="showStarted ? 'text-yellow-400' : 'text-ink-dim hover:text-ink-muted'"
          :title="showStarted ? 'Visa alla mål' : 'Visa endast påbörjade mål'">
          <Icon name="lucide:circle-play" class="w-5 h-5"
            :style="{ opacity: showStarted ? 1 : 0.3 }" />
        </button>
        <span class="hidden md:inline-flex ml-2 px-3 py-1.5 rounded-full bg-surface-2 border border-line text-xs text-ink-dim">
          j/k navigera · ↑↓ sortera · l öppna · Enter nytt · d klar · e evigt · c deadline · t beskrivning · x ta bort
        </span>
      </div>
    </div>

    <div v-if="pending" class="text-ink-muted">
      Laddar...
    </div>

    <div v-else-if="authError" class="text-center py-12">
      <div class="mb-4">
        <Icon name="lucide:alert-circle" class="w-12 h-12 text-yellow-500 mx-auto mb-4" />
        <p class="text-yellow-400 text-lg mb-2">Sessionen har gått ut</p>
        <p class="text-ink-muted">Du omdirigeras till inloggningssidan...</p>
      </div>
      <NuxtLink to="/login">
        <UButton color="warning" icon="lucide:log-in">
          Logga in
        </UButton>
      </NuxtLink>
    </div>

    <div v-else-if="error" class="text-red-400">
      Fel vid laddning av mål: {{ error.message }}
    </div>

    <div v-else-if="filteredGoals.length === 0 && !showNewGoalInput"
      class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-10 text-center text-ink-muted">
      Inga mål hittades.
    </div>

    <div v-else class="space-y-2">
      <!-- Nytt grundmål -->
      <div
        v-if="showNewGoalInput"
        class="flex items-center gap-3 p-4 rounded-2xl border border-accent bg-accent/10"
      >
        <Icon name="heroicons:star" class="w-6 h-6 flex-shrink-0 text-accent-light" />
        <input
          ref="newGoalInput"
          v-model="newGoalTitle"
          type="text"
          placeholder="Nytt grundmål..."
          class="flex-1 bg-transparent border-none outline-none placeholder-ink-dim text-ink"
          @keydown.enter.prevent="saveNewGoal"
          @keydown.esc.prevent="cancelNewGoal"
        />
      </div>

      <div
        v-for="(goal, index) in filteredGoals"
        :key="goal.id"
        class="space-y-2"
      >
        <div
          :data-goal-index="index"
          class="group flex items-center gap-3 px-4 py-3.5 rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 cursor-pointer"
          :class="[
            index === selectedIndex ? 'ring-1 ring-accent' : '',
            isWaitingForNextOccurrence(goal) ? 'opacity-50' : '',
          ]"
          @click="goToGoal(goal.id)"
        >
          <span
            class="w-1 h-8 rounded-full transition-colors"
            :class="index === selectedIndex ? 'bg-accent' : 'bg-transparent group-hover:bg-line-strong'"
          />
          <Icon :name="goal.icon || 'heroicons:star'" class="w-6 h-6 flex-shrink-0 text-ink-muted group-hover:text-ink transition-colors"
            :class="goal.eternal ? ETERNAL_TEXT : ''" />
          <Icon v-if="goal.places?.length" name="lucide:map-pin"
            class="w-4 h-4 flex-shrink-0 text-accent"
            :title="`Platser: ${goal.places.map((p) => p.name).join(', ')}`" />
          <div class="flex-1 min-w-0">
            <span class="block truncate text-ink" :class="goal.eternal ? ETERNAL_TEXT : ''">{{ goal.title }}</span>
            <p v-if="goal.description" class="text-xs text-ink-dim truncate mt-0.5">
              {{ descriptionPreview(goal.description) }}
            </p>
          </div>
          <!-- Started (klicka eller s) -->
          <button @click.stop="toggleStarted(goal)"
            class="flex-shrink-0 p-1 rounded transition-colors"
            :class="goal.started ? 'text-yellow-400' : 'text-ink-dim hover:text-ink-muted'"
            :title="goal.started ? 'Avbörja (s)' : 'Påbörja (s)'">
            <Icon name="lucide:circle-play" class="w-5 h-5"
              :style="{ opacity: goal.started ? 1 : 0.25 }" />
          </button>
          <!-- Evigt mål (e) -->
          <button @click.stop="toggleEternal(goal)"
            class="flex-shrink-0 p-1 rounded transition-colors"
            :class="goal.eternal ? ETERNAL_ACCENT + ' ' + ETERNAL_ACCENT_HOVER : 'text-ink-dim hover:text-ink-muted'"
            :title="goal.eternal ? 'Ta bort evigt-märket (e)' : 'Markera som evigt mål (e)'">
            <Icon name="lucide:infinity" class="w-5 h-5"
              :style="{ opacity: goal.eternal ? 1 : 0.25 }" />
          </button>
          <!-- Due-datum / deadline (c) -->
          <template v-if="dueGoalId === goal.id">
            <input ref="dueInputRef" v-model="dueDraft" type="date" @click.stop
              class="flex-shrink-0 rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
              @change="saveDue(goal, dueDraft)"
              @keydown.enter.prevent="saveDue(goal, dueDraft)"
              @keydown.esc.prevent="dueGoalId = null" />
            <button @click.stop="saveDue(goal, '')"
              class="flex-shrink-0 p-1 rounded text-ink-dim transition-colors hover:text-ink-muted"
              title="Rensa deadline">
              <Icon name="lucide:x" class="w-4 h-4" />
            </button>
          </template>
          <template v-else>
            <button @click.stop="openDuePicker(goal)"
              class="flex-shrink-0 p-1 rounded transition-colors"
              :class="goal.due ? dueToneClass(goal.due) : 'text-ink-dim hover:text-ink-muted'"
              :title="goal.due ? `Deadline: ${formatDueDate(goal.due)}` : 'Sätt deadline (c)'">
              <Icon name="lucide:calendar-clock" class="w-5 h-5"
                :style="{ opacity: goal.due ? 1 : 0.25 }" />
            </button>
            <span v-if="goal.due" class="text-sm tabular-nums flex-shrink-0" :class="dueToneClass(goal.due)">
              {{ formatDueDate(goal.due) }}
            </span>
          </template>
          <!-- Upprepningar: endast antalet slutförda förekomster -->
          <div class="relative flex-shrink-0 flex items-center gap-1">
            <button v-if="isRecurringGoal(goal)" @click.stop="openRecurrenceEditor(goal)"
              class="flex items-center gap-1 text-sm tabular-nums text-ink-dim transition-colors hover:text-ink-muted"
              :title="`${describeGoalRecurrence(goal.recurrence_rule)} - ${recurrenceCount(goal)} slutförda (r)`">
              <Icon name="lucide:repeat" class="w-4 h-4" />
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
          </div>
          <span v-if="goal.finished" class="text-green-400 text-sm flex-shrink-0">✓ Klar</span>
          <span v-else-if="goal.eternal" :class="ETERNAL_ACCENT + ' text-sm flex-shrink-0'">∞</span>
          <span v-else-if="!isRecurringGoal(goal)" class="text-ink-dim text-sm tabular-nums flex-shrink-0">#{{ goal.id }}</span>
        </div>

        <!-- Redigera beskrivning (t) -->
        <div v-if="editingDescriptionGoalId === goal.id" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-4">
          <textarea
            id="goal-description-input"
            v-model="descriptionDraft"
            rows="4"
            class="w-full bg-surface-2 border border-accent rounded p-3 text-ink focus:outline-none focus:ring-2 focus:ring-accent resize-y"
            placeholder="Beskrivning - vad vill du uppnå?"
            @keydown="handleDescriptionKeydown($event, goal)"
            @keydown.esc.prevent="cancelEditDescription"
          />
          <div class="flex items-center justify-between mt-2 gap-4">
            <span class="text-xs text-ink-dim">Ctrl+Enter spara · Esc avbryta</span>
            <div class="flex gap-2">
              <button @click="cancelEditDescription" class="rounded-lg border border-line bg-surface-2 px-4 py-2 text-ink-muted transition-colors hover:border-line-strong hover:text-ink">Avbryt</button>
              <button @click="saveDescription(goal)"
                class="px-3 py-1.5 bg-accent-strong hover:bg-accent rounded-lg text-white transition-colors">
                Spara
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sökfält -->
    <div v-if="showSearch" class="mt-4">
      <input
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        placeholder="Sök mål..."
        class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full"
        @keydown.escape="showSearch = false"
      />
    </div>

    <!-- Delete confirmation modal -->
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
            @click="handleDelete(goalToDelete!)"
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
  </div>
</template>

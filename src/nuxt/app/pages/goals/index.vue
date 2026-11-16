<script setup lang="ts">
import type { Goal } from '~/types/goal'
import { AuthenticationError } from '~/composables/useGoalApi'
import { canBeFinished } from '~/utils/goalRules'
import { isRootGoal } from '~/utils/goalFilters'
import { ETERNAL_ACCENT, ETERNAL_ACCENT_HOVER, ETERNAL_TEXT } from '~/utils/eternalStyle'

const router = useRouter()
const { toggleGoalFinished, toggleGoalStarted, setGoalEternal, deleteGoal, loadAllGoals, createGoal } = useGoalApi()

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

// Endast grundmål (utan förälder) visas under /goals - undermål nås
// via sitt föräldermål.
const displayedGoals = computed(() => localGoals.value.filter(isRootGoal))
const showStarted = ref(false)

const filteredGoals = computed(() => {
  let result = displayedGoals.value
  if (showStarted.value) {
    result = result.filter(g => g.started !== null)
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(g => g.title.toLowerCase().includes(query))
  }
  return result
})

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

  // s - toggle started
  if (event.key === 's') {
    event.preventDefault()
    if (filteredGoals.value.length > 0) {
      toggleStarted(filteredGoals.value[selectedIndex.value])
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
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-3xl font-bold">Mina mål</h1>
      <div class="flex items-center gap-3">
        <!-- Nytt grundmål -->
        <button @click="startNewGoal"
          class="p-1 rounded transition-colors text-gray-500 hover:text-gray-300"
          title="Nytt grundmål (Enter)">
          <Icon name="lucide:plus" class="w-5 h-5" />
        </button>
        <!-- Visa endast påbörjade -->
        <button @click="showStarted = !showStarted"
          class="p-1 rounded transition-colors"
          :class="showStarted ? 'text-yellow-400 hover:text-yellow-300' : 'text-gray-500 hover:text-gray-300'"
          :title="showStarted ? 'Visa alla mål' : 'Visa endast påbörjade mål'">
          <Icon name="lucide:circle-play" class="w-5 h-5"
            :style="{ opacity: showStarted ? 1 : 0.3 }" />
        </button>
        <span class="text-gray-500 text-sm">
          j/k navigera · l öppna · Enter nytt grundmål · d klar · e evigt mål · x ta bort
        </span>
      </div>
    </div>

    <div v-if="pending" class="text-gray-400">
      Laddar...
    </div>

    <div v-else-if="authError" class="text-center py-12">
      <div class="mb-4">
        <Icon name="lucide:alert-circle" class="w-12 h-12 text-yellow-500 mx-auto mb-4" />
        <p class="text-yellow-400 text-lg mb-2">Sessionen har gått ut</p>
        <p class="text-gray-400">Du omdirigeras till inloggningssidan...</p>
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

    <div v-else-if="displayedGoals.length === 0 && !showNewGoalInput" class="text-gray-400">
      Inga mål hittades.
    </div>

    <div v-else class="space-y-2">
      <!-- Nytt grundmål -->
      <div
        v-if="showNewGoalInput"
        class="flex items-center gap-3 p-4 rounded-lg bg-blue-600 text-white"
      >
        <Icon name="heroicons:star" class="w-6 h-6 flex-shrink-0" />
        <input
          ref="newGoalInput"
          v-model="newGoalTitle"
          type="text"
          placeholder="Nytt grundmål..."
          class="flex-1 bg-transparent border-none outline-none placeholder-blue-200 text-white"
          @keydown.enter.prevent="saveNewGoal"
          @keydown.esc.prevent="cancelNewGoal"
        />
      </div>

      <div
        v-for="(goal, index) in filteredGoals"
        :key="goal.id"
        class="flex items-center gap-3 p-4 rounded-lg transition-all cursor-pointer"
        :class="index === selectedIndex 
          ? 'bg-blue-600 text-white' 
          : 'bg-gray-800 text-gray-200 hover:bg-gray-700'"
        @click="goToGoal(goal.id)"
      >
        <Icon :name="goal.icon || 'heroicons:star'" class="w-6 h-6 flex-shrink-0"
          :class="goal.eternal ? ETERNAL_TEXT : ''" />
        <!-- Started (klicka eller s) -->
        <button @click.stop="toggleStarted(goal)"
          class="flex-shrink-0 p-1 rounded transition-colors"
          :class="goal.started ? 'text-yellow-400' : 'text-gray-600 hover:text-gray-400'"
          :title="goal.started ? 'Avbörja (s)' : 'Påbörja (s)'">
          <Icon name="lucide:circle-play" class="w-5 h-5"
            :style="{ opacity: goal.started ? 1 : 0.25 }" />
        </button>
        <!-- Evigt mål (e) -->
        <button @click.stop="toggleEternal(goal)"
          class="flex-shrink-0 p-1 rounded transition-colors"
          :class="goal.eternal ? ETERNAL_ACCENT + ' ' + ETERNAL_ACCENT_HOVER : 'text-gray-600 hover:text-gray-400'"
          :title="goal.eternal ? 'Ta bort evigt-märket (e)' : 'Markera som evigt mål (e)'">
          <Icon name="lucide:infinity" class="w-5 h-5"
            :style="{ opacity: goal.eternal ? 1 : 0.25 }" />
        </button>
        <span class="flex-1 truncate" :class="goal.eternal ? ETERNAL_TEXT : ''">{{ goal.title }}</span>
        <span v-if="goal.finished" class="text-green-400 text-sm">✓ Klar</span>
        <span v-else-if="goal.eternal" :class="ETERNAL_ACCENT + ' text-sm'">∞</span>
        <span v-else class="text-gray-500 text-sm">#{{ goal.id }}</span>
      </div>
    </div>

    <!-- Sökfält -->
    <div v-if="showSearch" class="mt-4">
      <input
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        placeholder="Sök mål..."
        class="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
        @keydown.escape="showSearch = false"
      />
    </div>

    <!-- Delete confirmation modal -->
    <div 
      v-if="showDeleteConfirm && goalToDelete" 
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70"
      @click.self="showDeleteConfirm = false"
    >
      <div class="bg-gray-800 border border-gray-700 rounded-lg p-6 max-w-md">
        <h3 class="text-lg font-bold mb-2">Ta bort mål?</h3>
        <p class="text-gray-400 mb-4">
          Är du säker på att du vill ta bort "{{ goalToDelete.title }}"?
          Detta kan inte ångras.
        </p>
        <div class="flex gap-3">
          <button 
            @click="showDeleteConfirm = false"
            class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded transition-colors"
          >
            Avbryt
          </button>
          <button 
            @click="handleDelete(goalToDelete!)"
            class="px-4 py-2 bg-red-600 hover:bg-red-500 rounded transition-colors"
          >
            Ta bort
          </button>
        </div>
        <p class="text-gray-500 text-sm mt-4">
          Tryck Escape för att avbryta
        </p>
      </div>
    </div>
  </div>
</template>

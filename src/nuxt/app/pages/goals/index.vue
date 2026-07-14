<script setup lang="ts">
import type { Goal } from '~/types/goal'

const router = useRouter()
const { toggleGoalFinished, toggleGoalStarted, deleteGoal } = useGoalApi()

const { data: goals, pending, error, refresh } = await useFetch<Goal[]>('/api/goals')

const selectedIndex = ref(0)
const mode = ref<'normal' | 'insert'>('normal')

// Lokal state för att uppdatera UI direkt
const localGoals = ref<Goal[]>([])
watch(goals, (newGoals) => {
  if (newGoals) localGoals.value = newGoals
}, { immediate: true })

const displayedGoals = computed(() => localGoals.value)
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

// Öppna sök
const showSearch = ref(false)
const searchQuery = ref('')

// Delete confirmation
const showDeleteConfirm = ref(false)
const goalToDelete = ref<Goal | null>(null)

// Toggle finished
async function toggleFinished(goal: Goal) {
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

  // Enter eller l - öppna markerat mål
  if (event.key === 'Enter' || event.key === 'l') {
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
})
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-3xl font-bold">Mina mål</h1>
      <div class="flex items-center gap-3">
        <!-- Visa endast påbörjade -->
        <button @click="showStarted = !showStarted"
          class="p-1 rounded transition-colors"
          :class="showStarted ? 'text-yellow-400 hover:text-yellow-300' : 'text-gray-500 hover:text-gray-300'"
          :title="showStarted ? 'Visa alla mål' : 'Visa endast påbörjade mål'">
          <Icon name="lucide:circle-play" class="w-5 h-5"
            :style="{ opacity: showStarted ? 1 : 0.3 }" />
        </button>
        <span class="text-gray-500 text-sm">
          j/k navigera · l/Enter öppna · d klar · x ta bort
        </span>
      </div>
    </div>

    <div v-if="pending" class="text-gray-400">
      Laddar...
    </div>

    <div v-else-if="error" class="text-red-400">
      Fel vid laddning av mål: {{ error.message }}
    </div>

    <div v-else-if="displayedGoals.length === 0" class="text-gray-400">
      Inga mål hittades.
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="(goal, index) in filteredGoals"
        :key="goal.id"
        class="flex items-center gap-3 p-4 rounded-lg transition-all cursor-pointer"
        :class="index === selectedIndex 
          ? 'bg-blue-600 text-white' 
          : 'bg-gray-800 text-gray-200 hover:bg-gray-700'"
        @click="goToGoal(goal.id)"
      >
        <Icon :name="goal.icon || 'heroicons:star'" class="w-6 h-6 flex-shrink-0" />
        <!-- Started (klicka eller s) -->
        <button @click.stop="toggleStarted(goal)"
          class="flex-shrink-0 p-1 rounded transition-colors"
          :class="goal.started ? 'text-yellow-400' : 'text-gray-600 hover:text-gray-400'"
          :title="goal.started ? 'Avmarkerad (s)' : 'Påbörja (s)'">
          <Icon name="lucide:circle-play" class="w-5 h-5"
            :style="{ opacity: goal.started ? 1 : 0.25 }" />
        </button>
        <span class="flex-1 truncate">{{ goal.title }}</span>
        <span v-if="goal.finished" class="text-green-400 text-sm">✓ Klar</span>
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

<script setup lang="ts">
import type { Goal } from '~/types/goal'

// Väljare som gör ett mål beroende av ett annat mål.
//
// Målet som ska få ett beroende (targetId) skickas in, och väljaren
// söker bland alla mål i grafen. Enter väljer det markerade målet.
const props = defineProps<{
  open: boolean
  targetId: number | null
  targetTitle: string
}>()

const emit = defineEmits<{
  add: [goalId: number]
  remove: [goalId: number]
  close: []
}>()

const goalsStore = useGoalsStore()

const query = ref('')
const selectedIndex = ref(0)
const searchInput = ref<HTMLInputElement | null>(null)

// Mål som inte går att välja:
//   - målet självt
//   - mål som redan är beroenden av målet
//   - mål som (transitivt) beror på målet - de skulle bilda en cykel
const blockedIds = computed(() => {
  const blocked = new Set<number>()
  const targetId = props.targetId
  if (targetId === null) return blocked

  blocked.add(targetId)

  for (const dep of goalsStore.dependencies) {
    if (dep.goal_id === targetId) {
      blocked.add(dep.depends_on_id)
    }
  }

  // Mål som beror på målet. Om målet skulle bero på ett av dem blir det
  // en cykel - servert svarar med ett fel, så vi filtrerar bort dem här.
  const seen = new Set<number>([targetId])
  const queue = [targetId]
  while (queue.length > 0) {
    const current = queue.shift()!
    for (const dep of goalsStore.dependencies) {
      if (dep.depends_on_id !== current) continue
      if (seen.has(dep.goal_id)) continue
      seen.add(dep.goal_id)
      blocked.add(dep.goal_id)
      queue.push(dep.goal_id)
    }
  }

  return blocked
})

// Beroenden som redan finns för målet. De visas alltid överst, även när
// sökfältet är tomt, så att de snabbt kan tas bort.
const existingDependencies = computed<Goal[]>(() => {
  const targetId = props.targetId
  if (targetId === null) return []

  const dependencyIds = new Set(
    goalsStore.dependencies
      .filter((dependency) => dependency.goal_id === targetId)
      .map((dependency) => dependency.depends_on_id),
  )

  return goalsStore.goals
    .filter((goal) => dependencyIds.has(goal.id))
    .sort((a, b) => a.title.localeCompare(b.title, 'sv'))
})

// Alla nya mål som matchar sökningen, sorterade på titel.
const results = computed<Goal[]>(() => {
  const needle = query.value.trim().toLowerCase()
  const matched = goalsStore.goals.filter((goal) => {
    if (blockedIds.value.has(goal.id)) return false
    if (!needle) return true
    return goal.title.toLowerCase().includes(needle)
  })

  return [...matched].sort((a, b) => a.title.localeCompare(b.title, 'sv'))
})

// Sökvägen till målet, så att lika titlar går att skilja åt.
function parentPath(goalId: number): string {
  const parents = goalsStore.getParents(goalId)
  if (parents.length === 0) return ''
  return parents.map((parent) => parent.title).join(' › ')
}

function close() {
  emit('close')
}

function add(goalId: number) {
  emit('add', goalId)
}

function remove(goalId: number) {
  emit('remove', goalId)
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (results.value.length > 0) {
      selectedIndex.value = Math.min(selectedIndex.value + 1, results.value.length - 1)
    }
    return
  }

  if (event.key === 'ArrowUp') {
    event.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
    return
  }

  if (event.key === 'Enter') {
    event.preventDefault()
    const goal = results.value[selectedIndex.value]
    if (goal) {
      add(goal.id)
    }
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    close()
  }
}

watch(
  () => props.open,
  async (open) => {
    if (!open) return
    query.value = ''
    selectedIndex.value = 0
    await nextTick()
    searchInput.value?.focus()
  },
)

// En ny sökning börjar alltid överst i listan.
watch(query, () => {
  selectedIndex.value = 0
})

// Håll markeringen inom listan när sökresultatet ändras.
watch(results, () => {
  if (selectedIndex.value >= results.value.length) {
    selectedIndex.value = Math.max(results.value.length - 1, 0)
  }
})
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 z-[70] flex items-start justify-center bg-black/50 pt-[12vh]"
    @click.self="close"
  >
    <div class="w-full max-w-xl mx-4 bg-surface rounded-2xl border border-line shadow-xl overflow-hidden">
      <div class="px-4 py-3 border-b border-line">
        <h2 class="text-sm font-medium text-ink">Välj beroende för "{{ targetTitle }}"</h2>
        <p class="text-xs text-ink-dim mt-0.5">
          Målet kommer att vänta på det du väljer. Enter väljer, Esc avbryter.
        </p>
      </div>

      <div v-if="existingDependencies.length > 0" class="px-3 pt-3">
        <h3 class="px-1 mb-1 text-xs font-medium uppercase tracking-wide text-ink-dim">
          Befintliga beroenden
        </h3>
        <ul class="space-y-1">
          <li
            v-for="dependency in existingDependencies"
            :key="dependency.id"
            class="flex items-center gap-2 rounded-lg border border-orange-500/30 bg-orange-500/10 px-3 py-2"
          >
            <Icon name="lucide:clock" class="w-4 h-4 text-orange-400 flex-shrink-0" />
            <span class="text-ink text-sm truncate flex-1">{{ dependency.title }}</span>
            <button
              type="button"
              class="rounded-md p-1 text-ink-dim transition-colors hover:bg-red-500/20 hover:text-red-400"
              title="Ta bort beroende"
              @click.stop="remove(dependency.id)"
            >
              <Icon name="lucide:x" class="w-4 h-4" />
            </button>
          </li>
        </ul>
      </div>

      <div class="p-3">
        <input
          ref="searchInput"
          v-model="query"
          type="text"
          placeholder="Sök bland alla mål..."
          class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full text-sm"
          @keydown="handleKeydown"
        />
      </div>

      <ul v-if="results.length > 0" class="max-h-[50vh] overflow-y-auto px-2 pb-2 space-y-1">
        <li v-for="(goal, index) in results" :key="goal.id">
          <button
            type="button"
            @click.prevent.stop="add(goal.id)"
            @mouseenter="selectedIndex = index"
            class="w-full text-left px-3 py-2 rounded-lg transition-colors flex items-center gap-2 border"
            :class="index === selectedIndex ? 'bg-elevated border-accent' : 'border-transparent hover:bg-elevated'"
          >
            <Icon :name="goal.icon || 'heroicons:star'" class="w-4 h-4 text-ink-muted flex-shrink-0" />
            <span class="text-ink text-sm truncate">{{ goal.title }}</span>
            <span v-if="parentPath(goal.id)" class="text-xs text-ink-dim truncate max-w-[40%]">
              {{ parentPath(goal.id) }}
            </span>
            <span v-if="goal.finished" class="text-xs text-green-400 ml-auto flex-shrink-0">Klar</span>
          </button>
        </li>
      </ul>

      <div v-else class="px-4 pb-4 text-sm text-ink-dim">
        Inga matchande mål
      </div>
    </div>
  </div>
</template>

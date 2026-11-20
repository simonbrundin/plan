<script setup lang="ts">
import type { GoalWithWeight } from '~/types/goal'
import { normalizeGoalWeights, redistributeGoalWeights, type GoalWeight } from '~/utils/goalWeights'

const props = defineProps<{
  children: GoalWithWeight[]
  selectedIndex: number
}>()

const emit = defineEmits<{
  select: [index: number]
  close: []
  update: [weights: GoalWeight[]]
  commit: [weights: GoalWeight[]]
}>()

const draft = ref<GoalWeight[]>([])
const basis = ref<GoalWeight[]>([])
const draggingBoundary = ref<number | null>(null)
const editor = ref<HTMLElement | null>(null)
const root = ref<HTMLElement | null>(null)

const segments = computed(() => props.children.map((child) => ({
  child,
  weight: draft.value.find((item) => item.childId === child.id)?.weight ?? 0,
})))

function reset() {
  draft.value = normalizeGoalWeights(props.children)
  basis.value = draft.value.map((item) => ({ ...item }))
}

// Vikterna uppdateras lokalt medan en tangentsekvens eller dragning pågår.
// Annars skulle varje avrundat steg bli nästa stegs nya proportioner.
watch(
  () => props.children.map((child) => child.id).join(','),
  reset,
  { immediate: true },
)

watch(
  () => props.selectedIndex,
  () => {
    basis.value = draft.value.map((item) => ({ ...item }))
  },
)

function boundaryPosition(index: number) {
  return segments.value
    .slice(0, index + 1)
    .reduce((sum, segment) => sum + segment.weight, 0)
}

function updateDraft(weights: GoalWeight[]) {
  draft.value = weights
  emit('update', weights)
}

function startBoundary(event: PointerEvent, index: number) {
  event.preventDefault()
  event.stopPropagation()
  draggingBoundary.value = index
  basis.value = draft.value.map((item) => ({ ...item }))
  ;(event.currentTarget as HTMLElement).setPointerCapture?.(event.pointerId)
}

function moveBoundary(event: PointerEvent) {
  if (draggingBoundary.value === null || !editor.value) return

  const bounds = editor.value.getBoundingClientRect()
  const position = Math.max(0, Math.min(100, ((event.clientY - bounds.top) / bounds.height) * 100))
  const child = props.children[draggingBoundary.value]
  if (!child) return

  const before = basis.value
    .filter((item) => props.children.findIndex((candidate) => candidate.id === item.childId) < draggingBoundary.value!)
    .reduce((sum, item) => sum + item.weight, 0)
  updateDraft(redistributeGoalWeights(basis.value, child.id, position - before))
}

function endBoundary() {
  if (draggingBoundary.value === null) return
  draggingBoundary.value = null
  basis.value = draft.value.map((item) => ({ ...item }))
  emit('commit', draft.value)
}

function handleKeydown(event: KeyboardEvent) {
  const key = event.key.toLowerCase()
  if (!['j', 'k', 'h', 'l', 'escape'].includes(key)) return

  event.preventDefault()
  event.stopPropagation()
  if (key === 'j') emit('select', Math.min(props.selectedIndex + 1, props.children.length - 1))
  if (key === 'k') emit('select', Math.max(props.selectedIndex - 1, 0))
  if (key === 'h' || key === 'l') {
    const selected = props.children[props.selectedIndex]
    const currentWeight = draft.value.find((item) => item.childId === selected?.id)?.weight ?? 0
    if (selected) {
      const next = redistributeGoalWeights(
        basis.value,
        selected.id,
        currentWeight + (key === 'h' ? -1 : 1),
      )
      updateDraft(next)
      emit('commit', next)
    }
  }
  if (key === 'escape') emit('close')
}

onMounted(() => {
  root.value?.focus()

  window.addEventListener('pointermove', moveBoundary)
  window.addEventListener('pointerup', endBoundary)
})

onUnmounted(() => {
  window.removeEventListener('pointermove', moveBoundary)
  window.removeEventListener('pointerup', endBoundary)
})
</script>

<template>
  <div
    ref="root"
    tabindex="0"
    class="relative flex h-full w-full items-center justify-center outline-none"
    @keydown.capture="handleKeydown"
  >
    <div
      ref="editor"
      class="relative flex h-full w-full max-w-sm flex-col overflow-hidden rounded-xl border border-line bg-surface-2"
      @pointermove="moveBoundary"
    >
      <button
        v-for="(segment, index) in segments"
        :key="segment.child.id"
        class="relative flex min-h-0 w-full shrink-0 items-center justify-center overflow-hidden px-4 text-center transition-colors"
        :class="selectedIndex === index ? 'bg-accent/20 text-ink' : 'text-ink-muted hover:bg-elevated'"
        :style="{ height: `${segment.weight}%` }"
        :title="segment.child.title"
        @click="emit('select', index)"
      >
        <span class="flex min-w-0 items-center gap-2">
          <Icon :name="segment.child.icon || 'heroicons:star'" class="h-5 w-5 shrink-0" />
          <span class="truncate text-sm font-medium">{{ segment.child.title }}</span>
        </span>
        <span class="absolute right-3 text-sm font-mono tabular-nums text-ink-dim">
          {{ segment.weight }} %
        </span>
      </button>

      <button
        v-for="(_, index) in segments.slice(0, -1)"
        :key="`boundary-${index}`"
        class="group absolute inset-x-0 z-20 h-4 -translate-y-1/2 cursor-row-resize"
        :style="{ top: `${boundaryPosition(index)}%` }"
        :aria-label="`Ändra vikt mellan ${segments[index]?.child.title} och ${segments[index + 1]?.child.title}`"
        @pointerdown="startBoundary($event, index)"
      >
        <span class="absolute inset-x-0 top-1/2 h-px -translate-y-1/2 bg-line-strong transition-colors group-hover:bg-accent" />
      </button>
    </div>

    <span class="absolute right-4 top-1/2 -translate-y-1/2 text-lg font-mono tabular-nums text-ink-muted">
      100 %
    </span>
  </div>
</template>

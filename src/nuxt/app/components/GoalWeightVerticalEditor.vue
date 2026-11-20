<script setup lang="ts">
import type { GoalWithWeight } from '~/types/goal'
import { normalizeGoalWeights, redistributeGoalWeights, type GoalWeight } from '~/utils/goalWeights'

const props = defineProps<{
  children: GoalWithWeight[]
  selectedIndex: number
}>()

const emit = defineEmits<{
  select: [index: number]
  update: [weights: GoalWeight[]]
  commit: [weights: GoalWeight[]]
}>()

const draft = ref<GoalWeight[]>([])
const draggingBoundary = ref<number | null>(null)
const editor = ref<HTMLElement | null>(null)

const segments = computed(() => props.children.map((child) => ({
  child,
  weight: draft.value.find((item) => item.childId === child.id)?.weight ?? 0,
})))

function reset() {
  draft.value = normalizeGoalWeights(props.children)
}

watch(
  () => props.children.map((child) => `${child.id}:${child.weight}`).join(','),
  reset,
  { immediate: true },
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
  ;(event.currentTarget as HTMLElement).setPointerCapture?.(event.pointerId)
}

function moveBoundary(event: PointerEvent) {
  if (draggingBoundary.value === null || !editor.value) return

  const bounds = editor.value.getBoundingClientRect()
  const position = Math.max(0, Math.min(100, ((event.clientY - bounds.top) / bounds.height) * 100))
  const child = props.children[draggingBoundary.value]
  if (!child) return

  const before = segments.value
    .slice(0, draggingBoundary.value)
    .reduce((sum, segment) => sum + segment.weight, 0)
  updateDraft(redistributeGoalWeights(draft.value, child.id, position - before))
}

function endBoundary() {
  if (draggingBoundary.value === null) return
  draggingBoundary.value = null
  emit('commit', draft.value)
}

onMounted(() => {
  window.addEventListener('pointermove', moveBoundary)
  window.addEventListener('pointerup', endBoundary)
})

onUnmounted(() => {
  window.removeEventListener('pointermove', moveBoundary)
  window.removeEventListener('pointerup', endBoundary)
})
</script>

<template>
  <div class="flex h-full w-full items-center justify-center">
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
      </button>

      <button
        v-for="(_, index) in segments.slice(0, -1)"
        :key="`boundary-${index}`"
        class="absolute inset-x-0 z-20 h-4 -translate-y-1/2 cursor-row-resize border-y border-accent/60 bg-accent/10 opacity-0 transition-opacity hover:opacity-100 focus:opacity-100"
        :style="{ top: `${boundaryPosition(index)}%` }"
        :aria-label="`Ändra vikt mellan ${segments[index]?.child.title} och ${segments[index + 1]?.child.title}`"
        @pointerdown="startBoundary($event, index)"
      />
    </div>
  </div>
</template>

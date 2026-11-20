<script setup lang="ts">
import type { GoalWithWeight } from '~/types/goal'
import { normalizeGoalWeights, redistributeGoalWeights, type GoalWeight } from '~/utils/goalWeights'

const props = defineProps<{
  children: GoalWithWeight[]
  childId: number
}>()

const emit = defineEmits<{
  update: [weights: GoalWeight[]]
  commit: [weights: GoalWeight[]]
}>()

const draft = ref<GoalWeight[]>([])

const average = computed(() => (props.children.length ? 100 / props.children.length : 0))
const maxDeviation = computed(() => Math.max(average.value, 100 - average.value))
const childWeight = computed(() => draft.value.find((item) => item.childId === props.childId)?.weight ?? 0)
const position = computed(() => 50 + ((childWeight.value - average.value) / maxDeviation.value) * 44)

function reset() {
  draft.value = normalizeGoalWeights(props.children)
}

watch(
  () => props.children.map((child) => `${child.id}:${child.weight}`).join(','),
  reset,
  { immediate: true },
)

function update(event: Event) {
  const delta = Number((event.target as HTMLInputElement).value)
  draft.value = redistributeGoalWeights(draft.value, props.childId, average.value + delta)
  emit('update', draft.value)
}

function commit() {
  emit('commit', draft.value)
}
</script>

<template>
  <div
    v-if="children.length > 1"
    class="relative h-6 w-20 shrink-0"
    :title="`Vikt jämfört med de andra målen`"
  >
    <div class="absolute inset-x-0 top-1/2 h-px -translate-y-1/2 bg-line" />
    <div class="absolute left-1/2 top-1/2 h-4 w-px -translate-y-1/2 bg-ink-dim/70" />
    <div
      class="absolute top-1/2 h-1 -translate-y-1/2 rounded-full bg-accent/70"
      :style="{
        left: `${Math.min(50, position)}%`,
        width: `${Math.abs(position - 50)}%`,
      }"
    />
    <span
      class="absolute top-1/2 z-10 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-accent bg-canvas"
      :style="{ left: `${position}%` }"
    />
    <input
      type="range"
      :min="-maxDeviation"
      :max="maxDeviation"
      step="1"
      :value="childWeight - average"
      class="absolute inset-0 z-20 h-full w-full cursor-ew-resize opacity-0"
      :aria-label="`Ändra vikt för målet`"
      :aria-valuetext="`${childWeight} procent vikt`"
      @input="update"
      @change="commit"
    />
  </div>
</template>

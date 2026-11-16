<script setup lang="ts">
import type { GoalWithWeight } from '~/types/goal'

interface SwipeState {
  startX: number
  currentX: number
  isSwiping: boolean
  childId: number | null
}

const props = defineProps<{
  filteredChildren: GoalWithWeight[]
  showCompleted: boolean
  showStarted: boolean
  dragOverChildId: number | null
  draggedChildId: number | null
  selectedChildIndex: number
  swipeState: SwipeState
  mode: 'normal' | 'insert'
  editingGoalId: number | null
  editTitle: string
  weightEditingChildId: number | null
  tempWeight: number
  childDependencies: Record<number, { dependsOn: any[]; blocking: any[] }>
  depDragChildId: number | null
  depDragOverChildId: number | null
  isDepDragging: boolean
}>()

const emit = defineEmits<{
  'toggle-completed': []
  'toggle-started': []
  'open-child-search': []
  'dragstart': [event: DragEvent, childId: number]
  'dragover': [event: DragEvent, childId: number]
  'dragleave': []
  'drop': [event: DragEvent, childId: number]
  'touchstart': [event: TouchEvent, childId: number]
  'touchmove': [event: TouchEvent]
  'touchend': [child: GoalWithWeight]
  'open-icon-picker': [childId: number]
  'update:editTitle': [title: string]
  'save-edit': []
  'cancel-edit': []
  'start-weight-edit': [child: GoalWithWeight]
  'update:tempWeight': [value: number]
  'toggle-finished': [child: GoalWithWeight]
  'toggle-started': [child: GoalWithWeight]
  'dep-mousedown': [event: MouseEvent, childId: number]
  'select-child': [index: number]
  'delete-child': [child: GoalWithWeight]
}>()

// Hjälpfunktioner för dependencies
function isChildBlocked(childId: number): boolean {
  const info = props.childDependencies[childId]
  if (!info || info.dependsOn.length === 0) return false
  return info.dependsOn.some((d: any) => !d.finished)
}

function getChildWaitingForTitle(childId: number): string | null {
  const info = props.childDependencies[childId]
  if (!info || info.dependsOn.length === 0) return null
  const waitingFor = info.dependsOn.find((d: any) => !d.finished)
  return waitingFor?.title || null
}

function getSwipeOffset(childId: number): number {
  if (props.swipeState.isSwiping && props.swipeState.childId === childId) {
    const delta = props.swipeState.currentX - props.swipeState.startX
    return Math.max(0, Math.min(delta, 100))
  }
  return 0
}

function getWeightStyle(weight: number): { color: string; opacity: number; fontWeight?: string } {
  return { color: '#FFFFFF', opacity: 1 }
}

function getIconOpacity(child: GoalWithWeight): number {
  if (isChildBlocked(child.id) || child.finished) return 0.6
  return 1
}
</script>

<template>
  <div class="flex-1 overflow-y-auto min-h-0">
    <!-- Header with toggle buttons -->
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-semibold text-ink-muted"></h2>
      <div class="flex items-center gap-2">
        <!-- Visa avklarade -->
        <button
          @click="$emit('toggle-completed')"
          class="text-ink-dim hover:text-ink-muted transition-colors p-1 rounded hover:bg-surface-2"
          :title="showCompleted ? 'Dölj avklarade mål' : 'Visa avklarade mål'"
        >
          <Icon :name="showCompleted ? 'lucide:eye' : 'lucide:eye-off'" class="h-5 w-5" />
        </button>
        <!-- Visa endast påbörjade -->
        <button
          @click="$emit('toggle-started')"
          class="text-ink-dim hover:text-ink-muted transition-colors p-1 rounded hover:bg-surface-2"
          :class="showStarted ? 'text-yellow-400' : ''"
          :title="showStarted ? 'Visa alla mål' : 'Visa endast påbörjade mål'"
        >
          <Icon name="lucide:circle-play" class="h-5 w-5" :style="{ opacity: showStarted ? 1 : 0.3 }" />
        </button>
        <!-- Lägg till undermål -->
        <button
          @click="$emit('open-child-search')"
          class="text-ink-dim hover:text-ink-muted transition-colors p-1 rounded hover:bg-surface-2"
          title="Lägg till undermål"
        >
          <Icon name="lucide:plus" class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- Children list -->
    <ul v-if="filteredChildren.length > 0" class="space-y-3" :class="{ 'select-none': isDepDragging }">
      <li
        v-for="(child, index) in filteredChildren"
        :key="child.id"
        :data-child-index="index"
        draggable="true"
        class="relative overflow-hidden rounded-lg transition-opacity"
        :class="[
          isChildBlocked(child.id) ? 'ml-4 pl-3' : '',
          dragOverChildId === child.id ? 'ring-2 ring-fuchsia-500 ring-offset-2 ring-offset-canvas' : '',
          draggedChildId === child.id ? 'opacity-30' : '',
          depDragOverChildId === child.id && depDragChildId !== child.id ? 'ring-2 ring-orange-500 ring-offset-2 ring-offset-canvas' : ''
        ]"
        @dragstart="$emit('dragstart', $event, child.id)"
        @dragover="$emit('dragover', $event, child.id)"
        @dragleave="$emit('dragleave')"
        @drop="$emit('drop', $event, child.id)"
      >
        <!-- Reparent indicator -->
        <div
          v-if="dragOverChildId === child.id && draggedChildId !== child.id"
          class="absolute inset-0 bg-fuchsia-500/20 flex items-center justify-center z-10 rounded-lg"
        >
          <div class="bg-fuchsia-600 text-white px-3 py-1 rounded-full text-sm font-medium">
            Flytta in i detta mål
          </div>
        </div>

        <!-- Dependency indicator -->
        <div
          v-if="depDragOverChildId === child.id && depDragChildId !== child.id"
          class="absolute inset-0 bg-orange-500/20 flex items-center justify-center z-10 rounded-lg"
        >
          <div class="bg-orange-600 text-white px-3 py-1 rounded-full text-sm font-medium">
            🔗 Skapa beroende
          </div>
        </div>

        <!-- Swipe bakgrund -->
        <div
          class="absolute inset-0 flex items-center justify-start px-6"
          :class="child.finished ? 'bg-red-900/50' : 'bg-transparent'"
        >
          <span class="text-2xl">{{ child.finished ? "↩️" : "" }}</span>
        </div>

        <!-- Huvudinnehåll -->
        <div
          class="relative rounded-lg transition-all bg-surface"
          :class="{
            'border border-accent': selectedChildIndex === index
          }"
          :style="{
            transform: `translateX(${getSwipeOffset(child.id)}px)`,
            transition: swipeState.isSwiping ? 'none' : 'transform 0.3s ease'
          }"
          draggable="true"
          @dragstart.stop="$emit('dragstart', $event, child.id)"
          @dragover.stop="$emit('dragover', $event, child.id)"
          @dragleave.stop="$emit('dragleave')"
          @drop.stop.prevent="$emit('drop', $event, child.id)"
          @touchstart="$emit('touchstart', $event, child.id)"
          @touchmove="$emit('touchmove', $event)"
          @touchend="$emit('touchend', child)"
        >
          <!-- Insert mode - visa input -->
          <div v-if="mode === 'insert' && editingGoalId === child.id" class="p-4">
            <input
              :value="editTitle"
              @input="$emit('update:editTitle', ($event.target as HTMLInputElement).value)"
              type="text"
              class="w-full px-3 py-2 bg-surface-2 border border-accent rounded text-ink focus:outline-none focus:ring-2 focus:ring-accent"
              @keydown.enter.prevent="$emit('save-edit')"
              @keydown.esc.prevent="$emit('cancel-edit')"
            />
            <div class="text-xs text-ink-dim mt-2">
              Enter för att spara, Escape för att avbryta
            </div>
          </div>

          <!-- Normal mode - visa länk -->
          <div v-else class="flex items-center gap-2">
            <!-- Icon button -->
            <button
              @click.stop="$emit('open-icon-picker', child.id)"
              :class="[
                'flex-shrink-0 transition-colors rounded p-1',
                isChildBlocked(child.id) || child.finished ? 'text-ink-dim' : 'text-white hover:text-white hover:bg-elevated'
              ]"
              title="Ändra ikon"
            >
              <Icon
                :name="child.icon || 'heroicons:star'"
                class="w-5 h-5"
                :style="{ color: isChildBlocked(child.id) || child.finished ? '#4B5563' : '#FFFFFF', opacity: getIconOpacity(child) }"
              />
            </button>

            <!-- Started-toggle (s) -->
            <button
              @click.stop="$emit('toggle-started', child)"
              :class="[
                'flex-shrink-0 p-1 rounded transition-colors',
                isChildBlocked(child.id) || child.finished ? 'text-ink-dim' : child.started ? 'text-yellow-400 hover:text-yellow-300' : 'text-white hover:text-ink-muted'
              ]"
              :title="child.started ? 'Avbörja (s)' : 'Påbörja (s)'"
            >
              <Icon
                name="lucide:circle-play"
                class="w-5 h-5"
                :style="{ opacity: isChildBlocked(child.id) || child.finished ? 0.6 : child.started ? 1 : 0.25 }"
              />
            </button>

            <!-- Dependency-drag handle -->
            <button
              @mousedown.stop="$emit('dep-mousedown', $event, child.id)"
              :class="[
                'flex-shrink-0 p-1 rounded transition-colors hover:bg-elevated',
                isDepDragging && depDragChildId === child.id ? 'cursor-grabbing' : 'cursor-grab',
                isChildBlocked(child.id) ? 'text-ink-dim' : 'text-white hover:text-orange-400'
              ]"
              title="Dra för att skapa beroende"
            >
              <Icon name="lucide:link" class="w-5 h-5" />
            </button>

            <!-- Title link -->
            <NuxtLink
              :to="`/goal/${child.id}`"
              class="flex-1 p-4 block cursor-pointer"
              @click="$emit('select-child', index)"
            >
              <h3
                class="text-lg font-normal select-none"
                :class="child.finished || isChildBlocked(child.id) ? 'text-ink-dim' : 'text-white'"
                :style="child.finished || isChildBlocked(child.id) ? {} : getWeightStyle(child.weight)"
              >
                {{ child.title }}
              </h3>
              <p v-if="getChildWaitingForTitle(child.id)" class="text-xs text-ink-dim mt-1">
                🔗 Väntar på: {{ getChildWaitingForTitle(child.id) }}
              </p>
            </NuxtLink>

            <!-- Weight button -->
            <button
              v-if="weightEditingChildId !== child.id"
              @click.stop="$emit('start-weight-edit', child)"
              :class="[
                'flex-shrink-0 transition-colors p-2 rounded',
                isChildBlocked(child.id) ? 'text-ink-dim' : 'text-white hover:text-ink-muted hover:bg-surface-2'
              ]"
              title="Ändra vikt"
            >
              <span
                :class="[
                  'text-xs font-mono px-2 py-1 rounded bg-surface-2',
                  isChildBlocked(child.id) ? 'text-ink-dim' : 'text-white'
                ]"
              >
                {{ child.weight }}
              </span>
            </button>

            <!-- Weight editing -->
            <div v-if="weightEditingChildId === child.id" class="px-4 pb-4">
              <div class="flex items-center gap-2">
                <input
                  :value="tempWeight"
                  @input="$emit('update:tempWeight', ($event.target as HTMLInputElement).valueAsNumber)"
                  type="range"
                  min="1"
                  max="200"
                  step="1"
                  class="flex-1"
                />
                <span class="text-sm text-ink-muted w-8">{{ tempWeight }}</span>
                <button
                  @click="$emit('save-weight')"
                  class="text-green-400 hover:text-green-300"
                >
                  ✓
                </button>
                <button
                  @click="$emit('cancel-weight-edit')"
                  class="text-red-400 hover:text-red-300"
                >
                  ✗
                </button>
              </div>
            </div>
          </div>
        </div>
      </li>
    </ul>

    <!-- Empty state -->
    <div
      v-else
      class="text-ink-dim p-6 border border-line rounded-lg text-center"
    >
      Inga undermål ännu. Skapa ett för att dela upp detta mål i mindre delar.
    </div>
  </div>
</template>

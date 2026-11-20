<script setup lang="ts">
import type { LeaderCommand } from "~/composables/useLeaderKey";

interface Props {
  open: boolean
  commands: LeaderCommand[]
  pending?: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'execute-command': [key: string]
}>()

function handleCommand(cmd: string) {
  emit('execute-command', cmd)
}
</script>

<template>
  <Teleport to="body">
    <div v-if="props.open" class="fixed inset-0 z-[100] flex items-center justify-center" style="background: rgba(0,0,0,0.7)">
      <div class="bg-surface border border-line rounded-lg shadow-lg p-6 w-96 max-h-[80vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold text-white">Leader Commands</h2>
          <span
            v-if="props.pending"
            class="font-mono font-bold text-accent bg-elevated border border-line-strong rounded px-2 py-0.5 text-sm"
          >{{ props.pending }}</span>
        </div>

        <div class="space-y-3">
          <button
            v-for="command in props.commands"
            :key="command.keys"
            @click="handleCommand(command.keys)"
            class="w-full text-left px-4 py-3 bg-elevated hover:bg-elevated rounded border border-line-strong text-ink"
          >
            <span class="font-mono font-bold text-accent">{{ command.label }}</span>
            <span class="ml-3">{{ command.title }}</span>
          </button>
        </div>

        <div class="text-center mt-4 pt-4 border-t border-line">
          <span class="text-ink-dim text-sm">Press ESC or click outside to close</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

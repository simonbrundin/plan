<script setup lang="ts">
interface Props {
  open: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'execute-command': [key: string]
}>()

const router = useRouter()

function handleCommand(cmd: string) {
  emit('execute-command', cmd)
}
</script>

<template>
  <Teleport to="body">
    <div v-if="props.open" class="fixed inset-0 z-[100] flex items-center justify-center" style="background: rgba(0,0,0,0.7)">
      <div class="bg-surface border border-line rounded-lg shadow-lg p-6 w-96">
        <h2 class="text-lg font-bold text-white mb-4">Leader Commands</h2>
        
        <div class="space-y-3">
          <button 
            @click="handleCommand('fd')"
            class="w-full text-left px-4 py-3 bg-elevated hover:bg-elevated rounded border border-line-strong text-ink">
            <span class="font-mono font-bold text-accent">F+D</span>
            <span class="ml-3">Toggle completed goals</span>
          </button>
          
          <button 
            @click="handleCommand('i')"
            class="w-full text-left px-4 py-3 bg-elevated hover:bg-elevated rounded border border-line-strong text-ink">
            <span class="font-mono font-bold text-accent">I</span>
            <span class="ml-3">Change icon</span>
          </button>
          
          <button 
            @click="handleCommand('gi')"
            class="w-full text-left px-4 py-3 bg-elevated hover:bg-elevated rounded border border-line-strong text-ink">
            <span class="font-mono font-bold text-accent">G+I</span>
            <span class="ml-3">Go to inbox</span>
          </button>
          
          <button 
            @click="handleCommand('gg')"
            class="w-full text-left px-4 py-3 bg-elevated hover:bg-elevated rounded border border-line-strong text-ink">
            <span class="font-mono font-bold text-accent">G+G</span>
            <span class="ml-3">Go to goals list</span>
          </button>
        </div>
        
        <div class="text-center mt-4 pt-4 border-t border-line">
          <span class="text-ink-dim text-sm">Press ESC or click outside to close</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>
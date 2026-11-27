<script setup lang="ts">
const props = defineProps<{
  selectedCount: number;
  allSelected: boolean;
  busy?: boolean;
}>();

const emit = defineEmits<{
  (event: "toggle-all"): void;
  (
    event: "action",
    action: "now" | "started" | "agent" | "eternal" | "finished",
  ): void;
  (event: "clear"): void;
}>();
</script>

<template>
  <div
    v-if="props.selectedCount > 0"
    class="flex flex-wrap items-center gap-2 rounded-xl border border-accent/30 bg-accent/5 px-3 py-2"
  >
    <button
      type="button"
      class="rounded-lg border border-line bg-surface-2 px-2.5 py-1.5 text-sm text-ink-muted transition-colors hover:border-line-strong hover:text-ink"
      @click="emit('toggle-all')"
    >
      <Icon :name="props.allSelected ? 'lucide:square-minus' : 'lucide:square-check-big'" class="mr-1 inline h-4 w-4" />
      {{ props.allSelected ? 'Avmarkera alla' : 'Markera alla' }}
    </button>
    <span class="text-sm text-ink-muted">{{ props.selectedCount }} markerade:</span>
    <button type="button" :disabled="props.busy" class="rounded-lg border border-accent/40 bg-surface-2 px-2.5 py-1.5 text-sm text-accent transition-colors hover:border-accent hover:bg-accent/10 disabled:cursor-wait disabled:opacity-50" @click="emit('action', 'finished')">Klar</button>
    <button type="button" :disabled="props.busy" class="rounded-lg border border-accent/40 bg-surface-2 px-2.5 py-1.5 text-sm text-accent transition-colors hover:border-accent hover:bg-accent/10 disabled:cursor-wait disabled:opacity-50" @click="emit('action', 'now')">Nu</button>
    <button type="button" :disabled="props.busy" class="rounded-lg border border-accent/40 bg-surface-2 px-2.5 py-1.5 text-sm text-accent transition-colors hover:border-accent hover:bg-accent/10 disabled:cursor-wait disabled:opacity-50" @click="emit('action', 'started')">Påbörjat</button>
    <button type="button" :disabled="props.busy" class="rounded-lg border border-accent/40 bg-surface-2 px-2.5 py-1.5 text-sm text-accent transition-colors hover:border-accent hover:bg-accent/10 disabled:cursor-wait disabled:opacity-50" @click="emit('action', 'agent')">Agent</button>
    <button type="button" :disabled="props.busy" class="rounded-lg border border-accent/40 bg-surface-2 px-2.5 py-1.5 text-sm text-accent transition-colors hover:border-accent hover:bg-accent/10 disabled:cursor-wait disabled:opacity-50" @click="emit('action', 'eternal')">Evigt</button>
    <button type="button" class="ml-auto px-2 py-1 text-sm text-ink-dim hover:text-ink" @click="emit('clear')">Rensa</button>
    <span v-if="props.busy" class="text-xs text-ink-dim">Sparar…</span>
  </div>
</template>


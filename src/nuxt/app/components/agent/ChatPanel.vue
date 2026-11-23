<script setup lang="ts">
import type { AgentRun } from '~/types/agent';

// Chattpanelen - Plan är agentens gränssnitt. Slack används inte.
const props = defineProps<{
  goalId: number;
  goalTitle: string;
  run?: AgentRun | null;
}>();

const emit = defineEmits<{
  (event: 'conversation-started', conversationId: string): void;
}>();

const {
  messages,
  conversationId,
  approval,
  streaming,
  error,
  available,
  load,
  loadApproval,
  startConversation,
  send,
  decide,
} = useAgentChat(toRef(props, 'goalId'));

const draft = ref('');
const listRef = ref<HTMLElement | null>(null);
const feedback = ref('');

// Scrollar till slutet när det kommer nya meddelanden.
watch(
  () => messages.value.length,
  async () => {
    await nextTick();
    listRef.value?.scrollTo({ top: listRef.value.scrollHeight, behavior: 'smooth' });
  },
);

onMounted(async () => {
  await loadApproval();
  if (props.run?.conversation_id) {
    conversationId.value = props.run.conversation_id;
    await load(props.run.conversation_id);
  }
});

// Startar en konversation om det inte finns en.
const ensureConversation = async () => {
  if (conversationId.value) return conversationId.value;
  const id = await startConversation(props.goalTitle);
  if (id) {
    emit('conversation-started', id);
  }
  return id;
};

const submit = async () => {
  const text = draft.value;
  if (!text.trim() || streaming.value) return;

  draft.value = '';
  await ensureConversation();
  await send(text);
};

const approve = async () => {
  await decide('approve');
};

const reject = async () => {
  await decide('reject', feedback.value);
  feedback.value = '';
};
</script>

<template>
  <section class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong">
    <header class="flex items-center justify-between px-4 py-3 border-b border-line">
      <div class="flex items-center gap-2">
        <Icon name="lucide:sparkles" class="w-4 h-4 text-accent" />
        <h3 class="text-sm font-medium text-ink">Agent</h3>
        <span
          v-if="run"
          class="text-xs px-2 py-0.5 rounded-full border"
          :class="run.status === 'blocked' ? 'text-amber-300 border-amber-500/40 bg-amber-500/10' : 'text-ink-dim border-line bg-surface-2'"
        >
          {{ run.status === 'blocked' ? 'Väntar på dig' : run.phase }}
        </span>
      </div>
      <button
        v-if="!conversationId"
        @click="ensureConversation"
        class="text-xs text-accent hover:text-accent-light transition-colors"
      >
        Starta chatt
      </button>
    </header>

    <!-- Förslag som väntar på beslut -->
    <div v-if="approval" class="px-4 py-3 border-b border-line bg-amber-500/5">
      <p class="text-xs text-amber-300 mb-2">Agenten föreslår en definition:</p>
      <p class="text-sm text-ink font-medium">{{ approval.title }}</p>
      <p class="text-sm text-ink-muted whitespace-pre-wrap mt-1">{{ approval.description }}</p>
      <div class="flex items-center gap-2 mt-3">
        <button
          @click="approve"
          class="px-3 py-1.5 bg-accent-strong hover:bg-accent rounded-lg text-white text-sm transition-colors"
        >
          Godkänn
        </button>
        <input
          v-model="feedback"
          type="text"
          placeholder="Vad ska ändras?"
          class="flex-1 rounded-lg border border-line bg-surface-2 px-3 py-1.5 text-sm text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none"
        />
        <button
          @click="reject"
          class="px-3 py-1.5 border border-line bg-surface-2 rounded-lg text-ink-muted text-sm hover:border-line-strong hover:text-ink transition-colors"
        >
          Avvisa
        </button>
      </div>
    </div>

    <!-- Meddelanden -->
    <div ref="listRef" class="max-h-96 overflow-y-auto px-4 py-3 space-y-3">
      <p v-if="!messages.length" class="text-sm text-ink-dim">
        Inga meddelanden än. Skriv vad du vill ha hjälp med.
      </p>
      <div
        v-for="message in messages"
        :key="message.id"
        class="flex"
        :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
      >
        <div
          class="max-w-[85%] rounded-2xl px-3.5 py-2 text-sm whitespace-pre-wrap"
          :class="message.role === 'user'
            ? 'bg-accent-strong text-white'
            : message.role === 'tool'
              ? 'bg-surface-2 text-ink-dim font-mono text-xs'
              : 'bg-surface-2 text-ink'"
        >
          <span v-if="message.tool_name" class="block text-[10px] uppercase tracking-wide opacity-70 mb-1">
            {{ message.tool_name }}
          </span>
          {{ message.content }}
        </div>
      </div>
      <p v-if="streaming" class="text-xs text-ink-dim">Agenten skriver…</p>
    </div>

    <p v-if="error" class="px-4 pb-2 text-xs text-red-300">{{ error }}</p>
    <p v-else-if="!available" class="px-4 pb-2 text-xs text-ink-dim">
      Agent-API:t är inte nåbar just nu.
    </p>

    <!-- Skrivfält -->
    <form class="flex items-center gap-2 px-4 py-3 border-t border-line" @submit.prevent="submit">
      <input
        v-model="draft"
        type="text"
        placeholder="Skriv till agenten…"
        class="flex-1 rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-sm text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
      />
      <button
        type="submit"
        :disabled="streaming || !draft.trim()"
        class="px-3.5 py-2 bg-accent-strong hover:bg-accent disabled:opacity-40 rounded-lg text-white text-sm transition-colors"
      >
        Skicka
      </button>
    </form>
  </section>
</template>

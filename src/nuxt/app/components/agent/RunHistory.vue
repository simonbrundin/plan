<script setup lang="ts">
import type { AgentApproval, AgentRun, AgentRunEvent } from '~/types/agent';

const props = defineProps<{
  events: AgentRunEvent[];
  run?: AgentRun | null;
  approvals?: AgentApproval[];
}>();

const emit = defineEmits<{
  decided: [approvalId: number];
}>();

const feedback = ref('');
const busy = ref(false);
const errorMessage = ref<string | null>(null);

// Det godkännande som väntar på beslut - det senaste för körningen.
const pendingApproval = computed<AgentApproval | null>(() => {
  const list = props.approvals ?? [];
  const pending = list.filter((item) => item.status === 'pending');
  return pending.length ? pending[0] : null;
});

// Föreslagna ändringar - exakt det som kommer att skrivas.
const proposedChanges = computed(() => pendingApproval.value?.payload?.changes ?? []);

// Testet som föreslogs, och hur det gick när agenten testade själv.
const testCommand = computed(() => pendingApproval.value?.payload?.test_command ?? '');
const preVerification = computed(() => pendingApproval.value?.payload?.pre_verification);

function formatDate(value: string): string {
  return new Intl.DateTimeFormat('sv-SE', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(new Date(value));
}

function payloadSummary(event: AgentRunEvent): string {
  if (!event.payload || Object.keys(event.payload).length === 0) return '';
  return JSON.stringify(event.payload);
}

// Beslutet fattas av en människa - därför går det via Plan, där agentens
// egen nyckel inte kan fatta beslut.
async function decide(decision: 'approve' | 'reject') {
  const approval = pendingApproval.value;
  if (!approval) return;

  busy.value = true;
  errorMessage.value = null;
  try {
    await $fetch(`/api/agent/approvals/${approval.id}/decision`, {
      method: 'POST',
      body: { decision, feedback: feedback.value },
    });
    feedback.value = '';
    emit('decided', approval.id);
  } catch {
    errorMessage.value = 'Kunde inte spara beslutet.';
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <section
    v-if="pendingApproval"
    class="rounded-2xl border border-warning/40 bg-warning/5 p-4"
  >
    <header>
      <h3 class="text-sm font-medium text-ink">Förslag väntar på ditt beslut</h3>
      <p class="mt-1 text-xs text-ink-dim">
        {{ pendingApproval.title }} · {{ formatDate(pendingApproval.created_at) }}
      </p>
    </header>

    <p v-if="pendingApproval.description" class="mt-2 text-xs text-ink-muted">
      {{ pendingApproval.description }}
    </p>

    <div v-if="preVerification" class="mt-3 rounded-lg border border-line bg-surface p-3">
      <p class="text-xs font-medium text-ink">
        Agenten testade förslaget själv:
        <span :class="preVerification.passed ? 'text-success' : 'text-error'">
          {{ preVerification.passed ? 'gick igenom' : 'gick inte igenom' }}
        </span>
      </p>
      <pre
        v-if="preVerification.output"
        class="mt-2 max-h-40 overflow-auto rounded bg-surface-2 p-2 text-[11px] text-ink-dim"
      >{{ preVerification.output }}</pre>
    </div>

    <div v-if="testCommand" class="mt-3">
      <p class="text-xs font-medium text-ink">Test som ska köras</p>
      <code class="mt-1 block rounded bg-surface-2 px-2 py-1 text-[11px] text-ink-muted">
        {{ testCommand }}
      </code>
    </div>

    <details v-if="proposedChanges.length" class="mt-3">
      <summary class="cursor-pointer text-xs font-medium text-ink">
        {{ proposedChanges.length }}
        {{ proposedChanges.length === 1 ? 'fil ändras' : 'filer ändras' }}
      </summary>
      <ul class="mt-2 space-y-2">
        <li v-for="change in proposedChanges" :key="change.path">
          <code class="text-[11px] text-ink-muted">{{ change.path }}</code>
          <pre
            class="mt-1 max-h-60 overflow-auto rounded bg-surface-2 p-2 text-[11px] text-ink-dim"
          >{{ change.content }}</pre>
        </li>
      </ul>
    </details>

    <textarea
      v-model="feedback"
      rows="2"
      placeholder="Kommentar till beslutet (valfritt)"
      class="mt-3 w-full rounded-lg border border-line bg-surface p-2 text-xs text-ink placeholder:text-ink-dim"
    />

    <p v-if="errorMessage" class="mt-2 text-xs text-error">{{ errorMessage }}</p>

    <div class="mt-3 flex items-center gap-2">
      <button
        type="button"
        class="rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-content disabled:opacity-50"
        :disabled="busy"
        @click="decide('approve')"
      >
        Godkänn
      </button>
      <button
        type="button"
        class="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-ink disabled:opacity-50"
        :disabled="busy"
        @click="decide('reject')"
      >
        Avvisa
      </button>
    </div>
  </section>

  <section v-if="props.events.length" class="rounded-2xl border border-line bg-surface">
    <header class="border-b border-line px-4 py-3">
      <h3 class="text-sm font-medium text-ink">Körningshistorik</h3>
      <p class="mt-1 text-xs text-ink-dim">{{ props.events.length }} händelser från AgentRun</p>
      <dl
        v-if="props.run"
        class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-ink-muted"
      >
        <div class="flex items-center gap-1">
          <dt>Start:</dt>
          <dd class="text-ink-dim">{{ props.run.trigger }}</dd>
        </div>
        <div v-if="props.run.initiator" class="flex items-center gap-1">
          <dt>Initiativtagare:</dt>
          <dd class="text-ink-dim">{{ props.run.initiator }}</dd>
        </div>
        <div class="flex items-center gap-1">
          <dt>Profil:</dt>
          <dd class="text-ink-dim">{{ props.run.agent_profile }}</dd>
        </div>
        <div v-if="props.run.finished_at" class="flex items-center gap-1">
          <dt>Slut:</dt>
          <dd class="text-ink-dim">{{ formatDate(props.run.finished_at) }}</dd>
        </div>
      </dl>
      <p v-if="props.run?.result" class="mt-2 text-xs text-ink-muted">
        <span class="text-ink-dim">Resultat:</span> {{ props.run.result }}
      </p>
    </header>
    <ol class="divide-y divide-line">
      <li v-for="event in props.events" :key="event.id" class="px-4 py-3">
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm font-medium text-ink">{{ event.type }}</span>
          <time class="shrink-0 text-xs text-ink-dim" :datetime="event.created_at">
            {{ formatDate(event.created_at) }}
          </time>
        </div>
        <div class="mt-1 flex items-center gap-2 text-xs text-ink-muted">
          <span v-if="event.phase" class="rounded-full border border-line bg-surface-2 px-2 py-0.5">
            {{ event.phase }}
          </span>
          <code v-if="payloadSummary(event)" class="truncate text-ink-dim">{{ payloadSummary(event) }}</code>
        </div>
      </li>
    </ol>
  </section>
</template>

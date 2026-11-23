<script setup lang="ts">
import type { AgentRun, AgentRunPhase, AgentRunStatus } from "~/types/agent";
import type { Goal } from "~/types/goal";
import { useAgentApi } from "~/composables/useAgentApi";

// Sidan hämtar användarens mål efter att sessionen etablerats - därför
// ingen SSR.
definePageMeta({
  ssr: false,
});

const { fetchAgentGoals, fetchAgentRuns } = useAgentApi();

const goals = ref<Goal[]>([]);
const runs = ref<AgentRun[]>([]);
const pending = ref(true);
const errorMessage = ref("");

// Fasen visas som ord - den är agentens livscykel, inte målets status.
const phaseLabels: Record<AgentRunPhase, string> = {
  define: "Specificerar",
  plan: "Planerar",
  execute: "Utför",
  verify: "Verifierar",
  completed: "Klar",
};

const statusLabels: Record<AgentRunStatus, string> = {
  queued: "Väntar",
  running: "Pågår",
  blocked: "Väntar på dig",
  done: "Klar",
  failed: "Misslyckades",
};

function statusClasses(status: AgentRunStatus): string {
  switch (status) {
    case "running":
      return "text-sky-300 border-sky-500/40 bg-sky-500/10";
    case "blocked":
      return "text-amber-300 border-amber-500/40 bg-amber-500/10";
    case "done":
      return "text-emerald-300 border-emerald-500/40 bg-emerald-500/10";
    case "failed":
      return "text-red-300 border-red-500/40 bg-red-500/10";
    default:
      return "text-ink-muted border-line bg-surface-2";
  }
}

// Den senaste körningen för ett mål - den som visas på raden.
const latestRun = computed(() => {
  const byGoal = new Map<number, AgentRun>();
  for (const run of runs.value) {
    const existing = byGoal.get(run.goal_id);
    if (!existing || new Date(run.updated_at) > new Date(existing.updated_at)) {
      byGoal.set(run.goal_id, run);
    }
  }
  return byGoal;
});

function formatUpdated(value: string): string {
  return new Intl.DateTimeFormat("sv-SE", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(value));
}

onMounted(async () => {
  try {
    const [agentGoals, agentRuns] = await Promise.all([
      fetchAgentGoals(),
      fetchAgentRuns(),
    ]);
    goals.value = agentGoals;
    runs.value = agentRuns;
  } catch (error) {
    errorMessage.value = "Kunde inte ladda agentmålen.";
    console.error("Failed to load agent goals:", error);
  } finally {
    pending.value = false;
  }
});
</script>

<template>
  <main class="mx-auto w-full max-w-4xl px-6 py-8">
    <div class="mb-8">
      <p class="mb-2 text-sm font-medium uppercase tracking-wider text-accent">Agenter</p>
      <h1 class="text-3xl font-bold tracking-tight text-ink">Agentmål</h1>
      <p class="mt-2 text-ink-muted">
        Mål som är markerade för en agent. Agenten arbetar vidare och du kan godkänna
        det som behöver godkännande.
      </p>
    </div>

    <div v-if="pending" class="rounded-2xl border border-line bg-surface p-10 text-center text-ink-muted">
      Laddar agentmål...
    </div>

    <div v-else-if="errorMessage" class="rounded-2xl border border-red-500/30 bg-red-500/10 p-6 text-red-300">
      {{ errorMessage }}
    </div>

    <div v-else-if="goals.length === 0" class="rounded-2xl border border-line bg-surface p-10 text-center">
      <Icon name="i-lucide-bot" class="mx-auto mb-4 h-10 w-10 text-ink-dim" />
      <p class="text-lg font-medium text-ink">Inga agentmål.</p>
      <p class="mt-2 text-ink-muted">
        Markera ett mål med agentmärket så att en agent kan ta tag i det.
      </p>
    </div>

    <div v-else class="overflow-hidden rounded-2xl border border-line bg-surface">
      <div class="border-b border-line px-5 py-3 text-sm text-ink-muted">
        {{ goals.length }} agentmål
      </div>
      <ul class="divide-y divide-line">
        <li v-for="goal in goals" :key="goal.id">
          <NuxtLink
            :to="`/goal/${goal.id}`"
            class="flex items-start gap-4 px-5 py-4 transition-colors hover:bg-surface-2"
          >
            <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface-2 text-ink-muted">
              <Icon :name="goal.icon || 'heroicons:star'" class="h-5 w-5" />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block truncate font-medium text-ink">{{ goal.title }}</span>
              <template v-if="latestRun.get(goal.id)">
                <span class="mt-2 flex flex-wrap items-center gap-2">
                  <span class="rounded-full border border-line bg-surface-2 px-2 py-0.5 text-xs text-ink-muted">
                    {{ phaseLabels[latestRun.get(goal.id)!.phase] }}
                  </span>
                  <span
                    class="rounded-full border px-2 py-0.5 text-xs"
                    :class="statusClasses(latestRun.get(goal.id)!.status)"
                  >
                    {{ statusLabels[latestRun.get(goal.id)!.status] }}
                  </span>
                  <span class="text-xs text-ink-dim">
                    {{ formatUpdated(latestRun.get(goal.id)!.updated_at) }}
                  </span>
                </span>
                <span v-if="latestRun.get(goal.id)!.summary" class="mt-2 block text-sm text-ink-muted">
                  {{ latestRun.get(goal.id)!.summary }}
                </span>
              </template>
              <span v-else class="mt-1 block text-sm text-ink-dim">Väntar på att agenten ska ta tag i det.</span>
            </span>
            <Icon name="i-lucide-chevron-right" class="mt-1 h-5 w-5 shrink-0 text-ink-dim" />
          </NuxtLink>
        </li>
      </ul>
    </div>
  </main>
</template>

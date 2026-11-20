<script setup lang="ts">
import type { Goal } from "~/types/goal";
import { AuthenticationError, useGoalApi } from "~/composables/useGoalApi";

// Avslutade mål behöver inte SSR eftersom sidan hämtar användarens mål efter
// att sessionen etablerats i webbläsaren.
definePageMeta({
  ssr: false,
});

type Interval = "today" | "week" | "month" | "all";

const router = useRouter();
const { loadGoalGraph } = useGoalApi();
const goals = ref<Goal[]>([]);
const parentTitles = ref(new Map<number, string[]>());
const interval = ref<Interval>("today");
const pending = ref(true);
const errorMessage = ref("");

const intervals: Array<{ value: Interval; label: string }> = [
  { value: "today", label: "Idag" },
  { value: "week", label: "Senaste veckan" },
  { value: "month", label: "Senaste 30 dagarna" },
  { value: "all", label: "Alla avslutade" },
];

function startOfToday(): Date {
  const today = new Date();
  return new Date(today.getFullYear(), today.getMonth(), today.getDate());
}

function intervalStart(value: Interval): Date | null {
  const today = startOfToday();
  if (value === "today") return today;
  if (value === "week") return new Date(today.getTime() - 6 * 24 * 60 * 60 * 1000);
  if (value === "month") return new Date(today.getTime() - 29 * 24 * 60 * 60 * 1000);
  return null;
}

const filteredGoals = computed(() => {
  const start = intervalStart(interval.value);
  return goals.value
    .filter((goal) => goal.finished)
    .filter((goal) => !start || new Date(goal.finished!).getTime() >= start.getTime())
    .sort((a, b) => new Date(b.finished!).getTime() - new Date(a.finished!).getTime());
});

function formatFinished(value: string): string {
  return new Intl.DateTimeFormat("sv-SE", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

onMounted(async () => {
  try {
    const graph = await loadGoalGraph();
    goals.value = graph.goals;

    const titlesById = new Map(graph.goals.map((goal) => [goal.id, goal.title]));
    const titlesByChild = new Map<number, string[]>();
    for (const relation of graph.relations) {
      const title = titlesById.get(relation.parent_id);
      if (!title) continue;
      const titles = titlesByChild.get(relation.child_id) ?? [];
      titles.push(title);
      titlesByChild.set(relation.child_id, titles);
    }
    parentTitles.value = titlesByChild;
  } catch (error) {
    if (error instanceof AuthenticationError) {
      await router.push("/login");
      return;
    }
    errorMessage.value = "Kunde inte ladda avslutade mål.";
    console.error("Failed to load completed goals:", error);
  } finally {
    pending.value = false;
  }
});
</script>

<template>
  <main class="mx-auto w-full max-w-4xl px-6 py-8">
    <div class="mb-8 flex flex-wrap items-end justify-between gap-4">
      <div>
        <p class="mb-2 text-sm font-medium uppercase tracking-wider text-accent">Historik</p>
        <h1 class="text-3xl font-bold tracking-tight text-ink">Avslutade mål</h1>
        <p class="mt-2 text-ink-muted">Se vad du har fått gjort den senaste tiden.</p>
      </div>

      <label class="flex flex-col gap-1.5 text-sm text-ink-muted">
        <span>Visa</span>
        <select
          v-model="interval"
          class="min-w-48 rounded-lg border border-line bg-surface-2 px-3 py-2 text-ink outline-none transition-colors focus:border-accent"
        >
          <option v-for="option in intervals" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </label>
    </div>

    <div v-if="pending" class="rounded-2xl border border-line bg-surface p-10 text-center text-ink-muted">
      Laddar avslutade mål...
    </div>

    <div v-else-if="errorMessage" class="rounded-2xl border border-red-500/30 bg-red-500/10 p-6 text-red-300">
      {{ errorMessage }}
    </div>

    <div v-else-if="filteredGoals.length === 0" class="rounded-2xl border border-line bg-surface p-10 text-center">
      <Icon name="i-lucide-check-circle-2" class="mx-auto mb-4 h-10 w-10 text-ink-dim" />
      <p class="text-lg font-medium text-ink">Inga avslutade mål i det här intervallet.</p>
      <p class="mt-2 text-ink-muted">Välj ett längre intervall för att se äldre mål.</p>
    </div>

    <div v-else class="overflow-hidden rounded-2xl border border-line bg-surface">
      <div class="border-b border-line px-5 py-3 text-sm text-ink-muted">
        {{ filteredGoals.length }} avslutade mål
      </div>
      <ul class="divide-y divide-line">
        <li v-for="goal in filteredGoals" :key="goal.id">
          <NuxtLink
            :to="`/goal/${goal.id}`"
            class="flex items-center gap-4 px-5 py-4 transition-colors hover:bg-surface-2"
          >
            <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface-2 text-ink-muted">
              <Icon :name="goal.icon || 'heroicons:star'" class="h-5 w-5" />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block truncate font-medium text-ink">{{ goal.title }}</span>
              <span class="mt-1 block text-sm text-ink-muted">{{ formatFinished(goal.finished!) }}</span>
              <span class="mt-1 block truncate text-sm text-ink-dim">
                {{ parentTitles.get(goal.id)?.join(", ") || "Ingen" }}
              </span>
            </span>
            <Icon name="i-lucide-chevron-right" class="h-5 w-5 shrink-0 text-ink-dim" />
          </NuxtLink>
        </li>
      </ul>
    </div>
  </main>
</template>

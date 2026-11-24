<script setup lang="ts">
import type {
  AgentRunUsage,
  AgentUsageSummary,
  AgentBudget,
  BudgetStatus,
} from "~/types/agent";

// Sidan visar vad agentkörningarna kostar och vilka budgetgränser som
// gäller. Plan äger siffrorna - den här sidan läser bara.
definePageMeta({
  ssr: false,
});

interface UsageResponse {
  summary: AgentUsageSummary;
  rows: AgentRunUsage[];
}

const usage = ref<UsageResponse | null>(null);
const budgets = ref<AgentBudget[]>([]);
const budgetStatuses = ref<Record<number, BudgetStatus>>({});
const pending = ref(true);
const errorMessage = ref("");

// Den nya budgeten.
const newBudget = reactive({
  name: "",
  period: "month",
  hard_limit_usd: 10,
  warning_limit_usd: undefined as number | undefined,
  saving: false,
});

async function load() {
  pending.value = true;
  errorMessage.value = "";
  try {
    const [usageResponse, budgetResponse] = await Promise.all([
      $fetch<UsageResponse>("/api/agent/usage"),
      $fetch<AgentBudget[]>("/api/agent/budgets"),
    ]);
    usage.value = usageResponse;
    budgets.value = budgetResponse;
    await loadBudgetStatuses();
  } catch {
    errorMessage.value = "Kunde inte hämta kostnaderna.";
  } finally {
    pending.value = false;
  }
}

// Budgetarnas läge hämtas var för sig - varje svar kan stå på sin egen.
async function loadBudgetStatuses() {
  const entries = await Promise.all(
    budgets.value.map(async (budget) => {
      try {
        const status = await $fetch<BudgetStatus>(`/api/agent/budgets/${budget.id}/status`);
        return [budget.id, status] as const;
      } catch {
        return null;
      }
    }),
  );
  budgetStatuses.value = Object.fromEntries(
    entries.filter((entry): entry is readonly [number, BudgetStatus] => entry !== null),
  );
}

// En budget pausas när gränsen är nådd - annars kan agenten inte köra
// igen förrän perioden går ut.
async function toggleBudget(budget: AgentBudget) {
  try {
    await $fetch(`/api/agent/budgets/${budget.id}`, {
      method: "PATCH",
      body: { active: !budget.active },
    });
    await load();
  } catch {
    errorMessage.value = "Kunde inte ändra budgeten.";
  }
}

async function removeBudget(budget: AgentBudget) {
  try {
    await $fetch(`/api/agent/budgets/${budget.id}`, { method: "DELETE" });
    await load();
  } catch {
    errorMessage.value = "Kunde inte ta bort budgeten.";
  }
}

async function createBudget() {
  newBudget.saving = true;
  errorMessage.value = "";
  try {
    await $fetch("/api/agent/budgets", {
      method: "POST",
      body: {
        name: newBudget.name || "Budget",
        period: newBudget.period,
        hard_limit_usd: newBudget.hard_limit_usd,
        warning_limit_usd: newBudget.warning_limit_usd,
      },
    });
    newBudget.name = "";
    await load();
  } catch {
    errorMessage.value = "Kunde inte skapa budgeten.";
  } finally {
    newBudget.saving = false;
  }
}

// Kostnaden visas i dollar med fyra decimaler - agentkörningar är
// billiga, så två decimaler skulle visa noll.
function formatCost(value: number | null | undefined): string {
  if (value === null || value === undefined) return "-";
  return `$${value.toFixed(4)}`;
}

function formatDate(value: string | null | undefined): string {
  if (!value) return "-";
  return new Date(value).toLocaleString("sv-SE");
}

const periodLabels: Record<string, string> = {
  day: "Dag",
  week: "Vecka",
  month: "Månad",
};

onMounted(load);
</script>

<template>
  <section class="space-y-4">
    <header>
      <h1 class="text-lg font-medium text-ink">Agentkostnad</h1>
      <p class="mt-1 text-sm text-ink-dim">
        Kostnaden räknas fram i Plan ur en versionsstyrd prislista, så att
        samma priser gäller för alla körningar.
      </p>
    </header>

    <p v-if="pending" class="text-sm text-ink-dim">Hämtar kostnader…</p>
    <p v-else-if="errorMessage" class="text-sm text-error">{{ errorMessage }}</p>

    <template v-else-if="usage">
      <div class="grid gap-3 sm:grid-cols-4">
        <div class="rounded-2xl border border-line bg-surface p-4">
          <p class="text-xs text-ink-dim">Körningar</p>
          <p class="mt-1 text-xl font-medium text-ink">{{ usage.summary.runs }}</p>
        </div>
        <div class="rounded-2xl border border-line bg-surface p-4">
          <p class="text-xs text-ink-dim">Total kostnad</p>
          <p class="mt-1 text-xl font-medium text-ink">
            {{ formatCost(usage.summary.total_cost_usd) }}
          </p>
        </div>
        <div class="rounded-2xl border border-line bg-surface p-4">
          <p class="text-xs text-ink-dim">Indata-tokens</p>
          <p class="mt-1 text-xl font-medium text-ink">
            {{ usage.summary.input_tokens }}
          </p>
        </div>
        <div class="rounded-2xl border border-line bg-surface p-4">
          <p class="text-xs text-ink-dim">Utdata-tokens</p>
          <p class="mt-1 text-xl font-medium text-ink">
            {{ usage.summary.output_tokens }}
          </p>
        </div>
      </div>

      <div v-if="usage.summary.by_model?.length" class="rounded-2xl border border-line bg-surface">
        <header class="border-b border-line px-4 py-3">
          <h2 class="text-sm font-medium text-ink">Per modell</h2>
        </header>
        <ul class="divide-y divide-line">
          <li
            v-for="entry in usage.summary.by_model"
            :key="`${entry.provider}/${entry.model}`"
            class="flex items-center justify-between px-4 py-2 text-sm"
          >
            <span class="text-ink">
              {{ entry.model || "okänd modell" }}
              <span class="text-ink-dim">· {{ entry.runs }} körningar</span>
            </span>
            <span class="text-ink-dim">{{ formatCost(entry.cost_usd) }}</span>
          </li>
        </ul>
      </div>

      <div class="rounded-2xl border border-line bg-surface">
        <header class="border-b border-line px-4 py-3">
          <h2 class="text-sm font-medium text-ink">Körningar</h2>
          <p class="mt-1 text-xs text-ink-dim">
            {{ usage.rows.length }} rader. En ofullständig rad betyder att
            agenten inte hann rapportera användningen.
          </p>
        </header>
        <ul class="divide-y divide-line">
          <li v-for="row in usage.rows" :key="row.id" class="px-4 py-3 text-sm">
            <div class="flex items-center justify-between gap-3">
              <span class="text-ink">
                Körning #{{ row.agent_run_id }}
                <span class="text-ink-dim">· mål #{{ row.goal_id }}</span>
              </span>
              <span class="text-ink-dim">{{ formatCost(row.cost_usd) }}</span>
            </div>
            <p class="mt-1 text-xs text-ink-muted">
              {{ row.model || "okänd modell" }} · {{ row.input_tokens }} in ·
              {{ row.output_tokens }} ut · {{ row.tool_calls }} verktygsanrop ·
              prislistan {{ row.price_list_version }}
              <span v-if="row.price_source === 'fallback'" class="text-warning">
                (schablonpris)
              </span>
              <span v-if="row.status === 'incomplete'" class="text-warning">
                (ofullständig)
              </span>
            </p>
            <p class="mt-1 text-xs text-ink-dim">{{ formatDate(row.created_at) }}</p>
          </li>
        </ul>
      </div>
    </template>

    <div class="rounded-2xl border border-line bg-surface p-4">
      <h2 class="text-sm font-medium text-ink">Budgetgränser</h2>
      <p class="mt-1 text-xs text-ink-dim">
        En hård gräns stoppar nya körningar när den nås. En varningsgräns
        varnar bara.
      </p>

      <ul v-if="budgets.length" class="mt-3 space-y-2">
        <li
          v-for="budget in budgets"
          :key="budget.id"
          class="rounded-lg border border-line bg-surface-2 px-3 py-2 text-sm"
        >
          <div class="flex items-center justify-between gap-3">
            <span class="text-ink">
              {{ budget.name }}
              <span class="text-ink-dim">
                · {{ periodLabels[budget.period] ?? budget.period }} ·
                gräns {{ formatCost(budget.hard_limit_usd) }}
              </span>
              <span v-if="!budget.active" class="text-warning">(pausad)</span>
            </span>
            <span class="flex items-center gap-2">
              <button
                type="button"
                class="text-xs text-ink-dim underline"
                @click="toggleBudget(budget)"
              >
                {{ budget.active ? "Pausa" : "Aktivera" }}
              </button>
              <button
                type="button"
                class="text-xs text-ink-dim underline"
                @click="removeBudget(budget)"
              >
                Ta bort
              </button>
            </span>
          </div>
          <p v-if="budgetStatuses[budget.id]" class="mt-1 text-xs text-ink-dim">
            Använd {{ formatCost(budgetStatuses[budget.id].spent_usd) }} ·
            reserverat {{ formatCost(budgetStatuses[budget.id].reserved_usd) }}
            <span v-if="budgetStatuses[budget.id].exceeded" class="text-warning">
              · gränsen är nådd
            </span>
            <span v-else-if="budgetStatuses[budget.id].warning" class="text-warning">
              · varning
            </span>
          </p>
        </li>
      </ul>
      <p v-else class="mt-3 text-xs text-ink-dim">Inga budgetar ännu.</p>

      <form class="mt-4 grid gap-3 sm:grid-cols-4" @submit.prevent="createBudget">
        <input
          v-model="newBudget.name"
          placeholder="Namn"
          class="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-dim"
        />
        <select
          v-model="newBudget.period"
          class="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
        >
          <option value="day">Dag</option>
          <option value="week">Vecka</option>
          <option value="month">Månad</option>
        </select>
        <input
          v-model.number="newBudget.hard_limit_usd"
          type="number"
          step="0.01"
          min="0"
          placeholder="Hård gräns i USD"
          class="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-dim"
        />
        <button
          type="submit"
          class="rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-content disabled:opacity-50"
          :disabled="newBudget.saving"
        >
          Lägg till budget
        </button>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { WorkflowPlan } from '~/types/workflow'
import { checkpointStatusLabel, workflowStatusLabel } from '~/types/workflow'
import { useWorkflowApi } from '~/composables/useWorkflowApi'

const props = defineProps<{
  feature: string
  description: string
}>()

const { listPlans, createPlan, approvePlan, rejectPlan } = useWorkflowApi()
const plan = ref<WorkflowPlan | null>(null)
const loading = ref(false)
const deciding = ref(false)
const error = ref('')
const expanded = ref(false)

const progress = computed(() => {
  const checkpoints = plan.value?.checkpoints ?? []
  const passed = checkpoints.filter((checkpoint) =>
    checkpoint.status === 'passed' || checkpoint.status === 'skipped',
  ).length
  return { passed, total: checkpoints.length }
})

const planStatusClass = computed(() => {
  switch (plan.value?.status) {
    case 'approved': return 'text-green-400'
    case 'rejected': return 'text-red-400'
    case 'completed': return 'text-blue-400'
    default: return 'text-amber-300'
  }
})

async function loadPlan() {
  try {
    const plans = await listPlans()
    plan.value = plans
      .filter((candidate) => candidate.feature === props.feature)
      .sort((a, b) => b.createdAt.localeCompare(a.createdAt))[0] ?? null
  } catch {
    // Agenten kan vara avstängd i dev. Det ska inte göra mål-sidan obrukbar.
    plan.value = null
  }
}

async function startPlan() {
  loading.value = true
  error.value = ''
  try {
    plan.value = await createPlan(props.feature, props.description)
    expanded.value = true
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'Kunde inte skapa checkpoint-planen.'
  } finally {
    loading.value = false
  }
}

async function decide(decision: 'approve' | 'reject') {
  if (!plan.value) return
  deciding.value = true
  error.value = ''
  try {
    if (decision === 'approve') await approvePlan(plan.value.id)
    else await rejectPlan(plan.value.id)
    await loadPlan()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'Kunde inte spara beslutet.'
  } finally {
    deciding.value = false
  }
}

onMounted(loadPlan)
</script>

<template>
  <section class="min-h-0 flex-1 overflow-y-auto bg-surface p-4" aria-label="Checkpoint-workflow">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h3 class="text-sm font-medium text-ink flex items-center gap-2">
          <Icon name="lucide:route" class="w-4 h-4 text-accent" />
          Checkpoint-plan
        </h3>
        <p v-if="plan" class="text-xs text-ink-dim mt-1">
          {{ progress.passed }}/{{ progress.total }} checkpoints klara ·
          <span :class="planStatusClass">{{ workflowStatusLabel(plan.status) }}</span>
        </p>
        <p v-else class="text-xs text-ink-dim mt-1">Dela upp målet i verifierbara checkpoints.</p>
      </div>
      <button
        v-if="!plan"
        type="button"
        class="rounded-lg bg-accent-strong px-3 py-1.5 text-xs text-white hover:bg-accent disabled:opacity-50"
        :disabled="loading"
        @click="startPlan"
      >
        {{ loading ? 'Skapar…' : 'Skapa plan' }}
      </button>
      <button
        v-else
        type="button"
        class="rounded-lg px-2 py-1 text-xs text-ink-muted hover:bg-elevated hover:text-ink"
        :aria-expanded="expanded"
        @click="expanded = !expanded"
      >
        {{ expanded ? 'Dölj' : 'Visa' }}
      </button>
    </div>

    <p v-if="error" class="mt-3 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-300" role="alert">
      {{ error }}
    </p>

    <div v-if="plan && expanded" class="mt-4 space-y-3">
      <div class="flex items-center justify-between gap-3">
        <p class="text-xs text-ink-muted">Planstatus: <span :class="planStatusClass">{{ workflowStatusLabel(plan.status) }}</span></p>
        <div v-if="plan.status === 'pending'" class="flex gap-2">
          <button type="button" class="rounded-lg border border-red-500/30 px-2.5 py-1 text-xs text-red-300 hover:bg-red-500/10 disabled:opacity-50" :disabled="deciding" @click="decide('reject')">Avvisa</button>
          <button type="button" class="rounded-lg bg-green-600 px-2.5 py-1 text-xs text-white hover:bg-green-500 disabled:opacity-50" :disabled="deciding" @click="decide('approve')">Godkänn plan</button>
        </div>
      </div>

      <ol class="space-y-2">
        <li v-for="(checkpoint, index) in plan.checkpoints" :key="checkpoint.id" class="rounded-xl border border-line bg-surface-2 p-3">
          <div class="flex items-start gap-3">
            <span class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-elevated text-xs text-ink-muted">{{ index + 1 }}</span>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <strong class="text-sm text-ink">{{ checkpoint.name }}</strong>
                <span class="text-xs text-ink-dim">{{ checkpointStatusLabel(checkpoint.status) }}</span>
              </div>
              <p v-if="checkpoint.description" class="mt-1 text-xs text-ink-muted">{{ checkpoint.description }}</p>
              <div v-if="checkpoint.gates.length" class="mt-2 flex flex-wrap gap-1">
                <span v-for="gate in checkpoint.gates" :key="gate" class="rounded bg-elevated px-1.5 py-0.5 text-[11px] text-ink-dim">{{ gate }}</span>
              </div>
            </div>
          </div>
        </li>
      </ol>
    </div>
  </section>
</template>

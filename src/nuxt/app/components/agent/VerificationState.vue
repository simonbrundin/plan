<script setup lang="ts">
import type { GoalEvidence, VerificationPlan, VerificationResult } from '~/utils/goalVerification';
import {
	getVerificationOutcomePresentation,
	getVerificationStrengthPresentation,
	getVerificationTypePresentation,
	needsApproval,
} from '~/utils/goalVerification';
import { isApiFeatureMissing, useGoalVerification } from '~/composables/useGoalVerification';

/**
 * Agentens bild av verifieringen - plan, bevis och omdöme.
 *
 * Panelen läser in sin egen status för målet. Agenten får lämna bevis och
 * köra verifieringen, men omdömet blir först giltigt när en människa
 * godkänner det - därför är knappen alltid grannens.
 */
const props = defineProps<{
	goalId: number;
}>();

const emit = defineEmits<{
	changed: [];
}>();

const { loadVerification, runVerification, approveResult, rejectResult } =
	useGoalVerification();

const plan = ref<VerificationPlan | null>(null);
const evidence = ref<GoalEvidence[]>([]);
const result = ref<VerificationResult | null>(null);
const busy = ref(false);
const errorMessage = ref<string | null>(null);
// Sant om API:t som kör inte stödjer verifieringen ännu.
const unsupported = ref(false);

const typePresentation = computed(() => getVerificationTypePresentation(plan.value?.type));
const strengthPresentation = computed(() =>
	getVerificationStrengthPresentation(plan.value?.strength),
);
const outcomePresentation = computed(() =>
	getVerificationOutcomePresentation(result.value?.outcome),
);
const waitingForApproval = computed(() => needsApproval(result.value));

async function refresh() {
	try {
		const state = await loadVerification(props.goalId);
		unsupported.value = state.unsupported === true;
		plan.value = state.verification?.id ? state.verification : null;
		evidence.value = state.evidence;
		result.value = state.result ?? null;
	} catch (error) {
		console.error('Failed to load verification:', error);
	}
}

onMounted(() => {
	void refresh();
});

watch(
	() => props.goalId,
	() => {
		void refresh();
	},
);

async function run() {
	busy.value = true;
	errorMessage.value = null;
	try {
		await runVerification(props.goalId);
		await refresh();
		emit('changed');
	} catch (error) {
		if (isApiFeatureMissing(error)) {
			unsupported.value = true;
			return;
		}
		errorMessage.value = 'Kunde inte köra verifieringen.';
	} finally {
		busy.value = false;
	}
}

async function approve(resultId: number) {
	busy.value = true;
	errorMessage.value = null;
	try {
		await approveResult(resultId);
		await refresh();
		emit('changed');
	} catch (error) {
		if (isApiFeatureMissing(error)) {
			unsupported.value = true;
			return;
		}
		errorMessage.value = 'Kunde inte godkänna omdömet.';
	} finally {
		busy.value = false;
	}
}

async function reject(resultId: number) {
	busy.value = true;
	errorMessage.value = null;
	try {
		await rejectResult(resultId);
		await refresh();
		emit('changed');
	} catch (error) {
		if (isApiFeatureMissing(error)) {
			unsupported.value = true;
			return;
		}
		errorMessage.value = 'Kunde inte avslå omdömet.';
	} finally {
		busy.value = false;
	}
}
</script>

<template>
	<section class="flex-shrink-0 border-b border-line p-4" data-testid="agent-verification">
		<p
			v-if="unsupported"
			class="text-xs text-ink-dim"
			data-testid="agent-verification-unsupported"
		>
			Det här API:t stödjer inte verifiering ännu.
		</p>

		<template v-else>
		<header class="flex items-center justify-between gap-2">
			<h3 class="text-sm font-medium text-ink">Verifiering</h3>
			<button
				v-if="plan"
				type="button"
				class="text-xs text-ink-dim hover:text-ink disabled:opacity-50"
				:disabled="busy"
				data-testid="agent-verification-run"
				@click="run"
			>
				Kör verifieringen
			</button>
		</header>

		<p v-if="!plan" class="mt-1 text-xs text-ink-dim">
			Målet har ingen verifieringsplan. Planen är valfri - den sätts när definitionen
			godkänns.
		</p>

		<div v-else class="mt-2 space-y-1">
			<p class="flex items-center gap-2 text-xs">
				<span :class="typePresentation.colorClass">{{ typePresentation.label }}</span>
				<span class="text-ink-dim">·</span>
				<span :class="strengthPresentation.colorClass">{{ strengthPresentation.label }}</span>
			</p>
			<p class="text-xs text-ink-dim">{{ plan.criterion }}</p>
		</div>

		<div v-if="evidence.length" class="mt-2">
			<p class="text-xs font-medium text-ink">Bevis ({{ evidence.length }})</p>
			<ul class="mt-1 space-y-0.5">
				<li v-for="item in evidence" :key="item.id" class="text-xs text-ink-dim">
					{{ item.value }}
				</li>
			</ul>
		</div>

		<div v-if="result" class="mt-2">
			<p class="flex items-center gap-2 text-xs">
				<span class="font-medium text-ink">Omdöme:</span>
				<span :class="outcomePresentation.colorClass">{{ outcomePresentation.label }}</span>
			</p>
			<p v-if="result.note" class="mt-0.5 text-xs text-ink-dim">{{ result.note }}</p>

			<div v-if="waitingForApproval" class="mt-2 flex items-center gap-2">
				<button
					type="button"
					class="rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-content disabled:opacity-50"
					:disabled="busy"
					data-testid="agent-verification-approve"
					@click="approve(result.id)"
				>
					Godkänn
				</button>
				<button
					type="button"
					class="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-ink disabled:opacity-50"
					:disabled="busy"
					data-testid="agent-verification-reject"
					@click="reject(result.id)"
				>
					Avslå
				</button>
				<span class="text-xs text-ink-dim">Agenten får inte besluta om omdömet.</span>
			</div>
		</div>

		<p v-if="errorMessage" class="mt-2 text-xs text-error">{{ errorMessage }}</p>
		</template>
	</section>
</template>

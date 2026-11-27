<script setup lang="ts">
import {
	VERIFICATION_STRENGTHS,
	VERIFICATION_TYPES,
	describeCompletion,
	getVerificationOutcomePresentation,
	getVerificationStrengthPresentation,
	getVerificationTypePresentation,
	needsApproval,
	type VerificationPlan,
	type VerificationResult,
} from "~/utils/goalVerification";

// Verifiering och bevis för ett mål.
//
// Panelen visar tre saker: om målet har en verifieringsplan, vilka bevis
// som finns, och vad omdömet blev. Allt är valfritt - ett mål utan plan
// visar bara behovet av att lägga till en.
//
// Agenten får lämna bevis och köra verifieringen, men aldrig godkänna
// omdömet. Knappen "Godkänn" är därför alltid människans.
const props = defineProps<{
	/** Målets id. */
	goalId: number;
	/** Planen, eller null om målet saknar plan. */
	plan: VerificationPlan | null;
	/** Bevisen som samlats in. */
	evidence: { id: number; type: string; value: string; created_at: string }[];
	/** Det senaste omdömet, om det finns. */
	result: VerificationResult | null;
	/** Hur målet avslutades ("deklarerat" eller "verifierat"). */
	completionKind?: string;
	/** Målets titel - visas i bevislistan. */
	goalTitle?: string;
}>();

const emit = defineEmits<{
	/** Planen har ändrats. */
	save: [plan: { type: string; criterion: string; strength: string }];
	/** Planen ska tas bort. */
	remove: [];
	/** Ett bevis har lagts till. */
	evidence: [evidence: { type: string; value: string }];
	/** Verifieringen ska köras. */
	run: [];
	/** Omdömet ska godkännas. */
	approve: [resultId: number];
	/** Omdömet ska avslås. */
	reject: [resultId: number];
}>();

const editing = ref(false);
const type = ref<string>(VERIFICATION_TYPES[0]);
const criterion = ref("");
const strength = ref<string>(VERIFICATION_STRENGTHS[1]);
const newEvidence = ref("");
const busy = ref(false);

const typePresentation = computed(() =>
	getVerificationTypePresentation(props.plan?.type),
);
const strengthPresentation = computed(() =>
	getVerificationStrengthPresentation(props.plan?.strength),
);
const outcomePresentation = computed(() =>
	getVerificationOutcomePresentation(props.result?.outcome),
);
const completion = computed(() => describeCompletion(props.completionKind));

// Omdömet är giltigt först när en människa godkänt det.
const waitingForApproval = computed(() => needsApproval(props.result));

function startEditing() {
	type.value = props.plan?.type ?? VERIFICATION_TYPES[0];
	criterion.value = props.plan?.criterion ?? "";
	strength.value = props.plan?.strength ?? VERIFICATION_STRENGTHS[1];
	editing.value = true;
}

function cancelEditing() {
	editing.value = false;
}

function save() {
	if (!criterion.value.trim()) return;
	busy.value = true;
	emit("save", {
		type: type.value,
		criterion: criterion.value.trim(),
		strength: strength.value,
	});
	editing.value = false;
	busy.value = false;
}

function addEvidence() {
	if (!newEvidence.value.trim()) return;
	busy.value = true;
	emit("evidence", { type: "text", value: newEvidence.value.trim() });
	newEvidence.value = "";
	busy.value = false;
}
</script>

<template>
	<section class="space-y-3" data-testid="goal-verification">
		<!-- Avslutandet - hur målet blev klart. -->
		<p
			v-if="completion"
			class="flex items-center gap-2 text-xs text-ink-dim"
			data-testid="goal-completion-kind"
		>
			<span class="font-medium text-ink">{{ completion.label }}</span>
			<span>{{ completion.description }}</span>
		</p>

		<!-- Planen. -->
		<div class="rounded-lg border border-line p-3">
			<div class="flex items-center justify-between gap-2">
				<h3 class="text-sm font-medium text-ink">Verifiering</h3>
				<button
					v-if="plan && !editing"
					class="text-xs text-ink-dim hover:text-ink"
					type="button"
					data-testid="verification-edit"
					@click="startEditing"
				>
					Ändra
				</button>
			</div>

			<!-- Ingen plan - den är valfri. -->
			<p
				v-if="!plan && !editing"
				class="mt-1 text-xs text-ink-dim"
				data-testid="verification-empty"
			>
				Målet har ingen verifieringsplan. Planen är valfri - utan den
				räknas målet som klart när du säger det.
			</p>

			<div v-if="plan && !editing" class="mt-2 space-y-1">
				<p class="flex items-center gap-2 text-sm text-ink">
					<span :class="typePresentation.colorClass">
						{{ typePresentation.label }}
					</span>
					<span class="text-ink-dim">·</span>
					<span :class="strengthPresentation.colorClass">
						{{ strengthPresentation.label }}
					</span>
				</p>
				<p class="text-xs text-ink-dim">{{ plan.criterion }}</p>
				<button
					class="text-xs text-ink-dim hover:text-ink"
					type="button"
					data-testid="verification-remove"
					@click="emit('remove')"
				>
					Ta bort planen
				</button>
			</div>

			<!-- Redigering. -->
			<div v-if="editing" class="mt-2 space-y-2">
				<label class="block text-xs text-ink-dim">
					Typ
					<select
						v-model="type"
						class="mt-1 w-full rounded border border-line bg-transparent px-2 py-1 text-sm text-ink"
						data-testid="verification-type"
					>
						<option v-for="t in VERIFICATION_TYPES" :key="t" :value="t">
							{{ getVerificationTypePresentation(t).label }}
						</option>
					</select>
				</label>
				<label class="block text-xs text-ink-dim">
					Villkor
					<input
						v-model="criterion"
						class="mt-1 w-full rounded border border-line bg-transparent px-2 py-1 text-sm text-ink"
						data-testid="verification-criterion"
						placeholder="Vad ska vara uppfyllt?"
					/>
				</label>
				<label class="block text-xs text-ink-dim">
					Styrka
					<select
						v-model="strength"
						class="mt-1 w-full rounded border border-line bg-transparent px-2 py-1 text-sm text-ink"
						data-testid="verification-strength"
					>
						<option v-for="s in VERIFICATION_STRENGTHS" :key="s" :value="s">
							{{ getVerificationStrengthPresentation(s).label }}
						</option>
					</select>
				</label>
				<div class="flex gap-2">
					<button
						class="rounded bg-accent px-3 py-1 text-xs font-medium text-ink"
						type="button"
						:disabled="!criterion.trim() || busy"
						data-testid="verification-save"
						@click="save"
					>
						Spara
					</button>
					<button
						class="rounded border border-line px-3 py-1 text-xs text-ink-dim"
						type="button"
						@click="cancelEditing"
					>
						Avbryt
					</button>
				</div>
			</div>

			<button
				v-if="!plan && !editing"
				class="mt-2 rounded border border-line px-3 py-1 text-xs text-ink-dim hover:text-ink"
				type="button"
				data-testid="verification-add"
				@click="startEditing"
			>
				Lägg till verifieringsplan
			</button>
		</div>

		<!-- Bevisen. -->
		<div class="rounded-lg border border-line p-3">
			<h3 class="text-sm font-medium text-ink">Bevis</h3>
			<ul v-if="evidence.length" class="mt-2 space-y-1">
				<li
					v-for="item in evidence"
					:key="item.id"
					class="text-xs text-ink-dim"
					data-testid="evidence-item"
				>
					{{ item.value }}
				</li>
			</ul>
			<p v-else class="mt-1 text-xs text-ink-dim">Inga bevis ännu.</p>

			<div class="mt-2 flex gap-2">
				<input
					v-model="newEvidence"
					class="flex-1 rounded border border-line bg-transparent px-2 py-1 text-sm text-ink"
					data-testid="evidence-input"
					placeholder="Skriv vad du har gjort"
					@keyup.enter="addEvidence"
				/>
				<button
					class="rounded border border-line px-3 py-1 text-xs text-ink-dim hover:text-ink"
					type="button"
					:disabled="!newEvidence.trim() || busy"
					data-testid="evidence-add"
					@click="addEvidence"
				>
					Lägg till
				</button>
			</div>
		</div>

		<!-- Omdömet. -->
		<div v-if="plan" class="rounded-lg border border-line p-3">
			<div class="flex items-center justify-between gap-2">
				<h3 class="text-sm font-medium text-ink">Omdöme</h3>
				<button
					class="text-xs text-ink-dim hover:text-ink"
					type="button"
					:disabled="busy"
					data-testid="verification-run"
					@click="emit('run')"
				>
					Kör verifieringen
				</button>
			</div>

			<p
				v-if="result"
				class="mt-2 flex items-center gap-2 text-sm"
				data-testid="verification-outcome"
			>
				<span :class="outcomePresentation.colorClass">
					{{ outcomePresentation.label }}
				</span>
				<span v-if="result.note" class="text-xs text-ink-dim">
					{{ result.note }}
				</span>
			</p>
			<p v-else class="mt-1 text-xs text-ink-dim">
				Verifieringen har inte körts ännu.
			</p>

			<div
				v-if="waitingForApproval && result"
				class="mt-2 flex items-center gap-2"
				data-testid="verification-approval"
			>
				<button
					class="rounded bg-accent px-3 py-1 text-xs font-medium text-ink"
					type="button"
					:disabled="busy"
					data-testid="verification-approve"
					@click="emit('approve', result.id)"
				>
					Godkänn
				</button>
				<button
					class="rounded border border-line px-3 py-1 text-xs text-ink-dim hover:text-ink"
					type="button"
					:disabled="busy"
					data-testid="verification-reject"
					@click="emit('reject', result.id)"
				>
					Avslå
				</button>
				<span class="text-xs text-ink-dim">
					Omdömet blir giltigt först när du godkänner det.
				</span>
			</div>
		</div>
	</section>
</template>

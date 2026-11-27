<script setup lang="ts">
import type { AgentApproval, AgentRun, AgentRunEvent } from '~/types/agent';
import { agentRunEventLabel, agentRunPhaseLabel } from '~/types/agent';
import { apiRequest } from '~/utils/api';

/**
 * Körningshistoriken - en kompakt tidslinje över händelserna i den
 * senaste körningen. Listan kan vara lång, så den visas bakad nyckelord
 * ("Visa fler") och varje händelse har sin payload ihopfällt.
 */
const props = defineProps<{
	events: AgentRunEvent[];
	run?: AgentRun | null;
	approvals?: AgentApproval[];
}>();

const emit = defineEmits<{
	decided: [approvalId: number];
	cancelled: [];
}>();

const feedback = ref('');
const busy = ref(false);
const errorMessage = ref<string | null>(null);

// Hur många händelser som syns utan att man ber om fler.
const VISIBLE_EVENTS = 12;
const showAllEvents = ref(false);

// Det godkännande som väntar på beslut - det senaste för körningen.
const pendingApproval = computed<AgentApproval | null>(() => {
	const list = props.approvals ?? [];
	const pending = list.filter((item) => item.status === 'pending');
	return pending[0] ?? null;
});

// Föreslagna ändringar - exakt det som kommer att skrivas.
const proposedChanges = computed(() => pendingApproval.value?.payload?.changes ?? []);

// Fält som agenten föreslår ändrade på målet - till exempel verifieringsplan
// och åtgärdsprofil. De visas innan beslutet, så att man vet vad man
// godkänner.
const proposedFields = computed(() => pendingApproval.value?.payload?.fields ?? []);

// Noteringar om värden agenten hoppade över.
const proposalNotes = computed(() => pendingApproval.value?.payload?.notes ?? []);

// Testet som föreslogs, och hur det gick när agenten testade själv.
const testCommand = computed(() => pendingApproval.value?.payload?.test_command ?? '');
const preVerification = computed(() => pendingApproval.value?.payload?.pre_verification);

// Händelserna bakad nyckelord - det senaste först.
const sortedEvents = computed(() =>
	[...props.events].sort((a, b) => b.created_at.localeCompare(a.created_at)),
);

const visibleEvents = computed(() =>
	showAllEvents.value ? sortedEvents.value : sortedEvents.value.slice(0, VISIBLE_EVENTS),
);

const hiddenCount = computed(() => Math.max(0, sortedEvents.value.length - VISIBLE_EVENTS));

// En händelses payload visas bara när man klickar på den.
const expandedEvent = ref<number | null>(null);

function formatDate(value: string): string {
	return new Intl.DateTimeFormat('sv-SE', {
		dateStyle: 'short',
		timeStyle: 'short',
	}).format(new Date(value));
}

function formatTime(value: string): string {
	return new Intl.DateTimeFormat('sv-SE', { timeStyle: 'short' }).format(
		new Date(value),
	);
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
		await apiRequest(`agent-approvals/${approval.id}/decision`, {
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

// Avbryter körningen. Det är ett sätt att stanna - jobbet och podden
// städas i klustret när jobbcontrollern ser att körningen är avbruten.
const canCancel = computed(
	() => props.run?.status === 'running' || props.run?.status === 'waiting_approval',
);

async function cancel() {
	if (!props.run) return;

	busy.value = true;
	errorMessage.value = null;
	try {
		await apiRequest(`agent-runs/${props.run.id}/cancel`, { method: 'POST' });
		emit('cancelled');
	} catch {
		errorMessage.value = 'Kunde inte avbryta körningen.';
	} finally {
		busy.value = false;
	}
}
</script>

<template>
	<div class="flex min-h-0 flex-1 flex-col bg-surface">
		<!-- Förslag som väntar på ditt beslut -->
		<section
			v-if="pendingApproval"
			class="flex-shrink-0 border-b border-warning/40 bg-warning/5 p-4"
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

			<!-- Fält agenten föreslår - t.ex. verifieringsplan och profil. -->
			<div v-if="proposedFields.length" class="mt-3" data-testid="proposal-fields">
				<p class="text-xs font-medium text-ink">Förslaget ändrar</p>
				<dl class="mt-1 space-y-1">
					<div v-for="field in proposedFields" :key="field.field" class="flex items-baseline gap-2 text-xs">
						<dt class="shrink-0 text-ink-dim">{{ field.label }}</dt>
						<dd class="text-ink">
							<span v-if="field.from" class="text-ink-dim line-through">{{ field.from }}</span>
							<span v-if="field.from" class="mx-1">→</span>
							<span>{{ field.to }}</span>
						</dd>
					</div>
				</dl>
			</div>

			<ul v-if="proposalNotes.length" class="mt-2 space-y-0.5">
				<li v-for="(note, index) in proposalNotes" :key="index" class="text-[11px] text-ink-dim">
					{{ note }}
				</li>
			</ul>

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

		<!-- Körningshistorik -->
		<section class="flex min-h-0 flex-1 flex-col">
			<header class="flex flex-shrink-0 items-start justify-between gap-3 border-b border-line px-4 py-3">
				<div>
					<h3 class="text-sm font-medium text-ink">Körningshistorik</h3>
					<p class="mt-1 text-xs text-ink-dim">
						{{ props.events.length }} händelser från AgentRun
					</p>
				</div>
				<button
					v-if="canCancel"
					type="button"
					class="flex-shrink-0 rounded-lg border border-error/50 px-3 py-1.5 text-xs font-medium text-error disabled:opacity-50"
					:disabled="busy"
					@click="cancel"
				>
					Avbryt körningen
				</button>
			</header>

			<!-- Metadata om körningen - utfällt som standard -->
			<details v-if="props.run" class="flex-shrink-0 border-b border-line px-4 py-2">
				<summary class="cursor-pointer text-xs text-ink-dim hover:text-ink-muted">
					Mer om körningen
				</summary>
				<dl
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
				<p v-if="props.run.result" class="mt-2 text-xs text-ink-muted">
					<span class="text-ink-dim">Resultat:</span> {{ props.run.result }}
				</p>
				<p v-if="errorMessage" class="mt-2 text-xs text-error">{{ errorMessage }}</p>
			</details>

			<ol v-if="props.events.length" class="min-h-0 flex-1 divide-y divide-line overflow-y-auto">
				<li v-for="event in visibleEvents" :key="event.id" class="px-4 py-2.5">
					<div class="flex items-center justify-between gap-3">
						<span class="text-sm font-medium text-ink">{{ agentRunEventLabel(event.type) }}</span>
						<time class="shrink-0 text-xs text-ink-dim" :datetime="event.created_at">
							{{ formatTime(event.created_at) }}
						</time>
					</div>
					<div class="mt-1 flex items-center gap-2 text-xs text-ink-muted">
						<span v-if="event.phase" class="rounded-full border border-line bg-surface-2 px-2 py-0.5">
							{{ agentRunPhaseLabel(event.phase) }}
						</span>
						<button
							v-if="payloadSummary(event)"
							type="button"
							class="truncate text-left text-ink-dim transition-colors hover:text-ink"
							@click="expandedEvent = expandedEvent === event.id ? null : event.id"
						>
							{{ expandedEvent === event.id ? 'Dölj detaljer' : 'Visa detaljer' }}
						</button>
					</div>
					<pre
						v-if="expandedEvent === event.id && payloadSummary(event)"
						class="mt-2 max-h-60 overflow-auto whitespace-pre-wrap break-words rounded bg-surface-2 p-2 text-[11px] text-ink-dim"
					>{{ JSON.stringify(event.payload, null, 2) }}</pre>
				</li>
			</ol>
			<p v-else class="px-4 py-6 text-center text-xs text-ink-dim">
				Ingen körning ännu för det här målet.
			</p>

			<button
				v-if="hiddenCount > 0 && !showAllEvents"
				type="button"
				class="flex-shrink-0 border-t border-line px-4 py-2 text-xs text-ink-dim transition-colors hover:bg-surface-2 hover:text-ink"
				@click="showAllEvents = true"
			>
				Visa {{ hiddenCount }} till
			</button>
		</section>
	</div>
</template>

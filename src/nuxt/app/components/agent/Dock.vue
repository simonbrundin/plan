<script setup lang="ts">
import type { AgentApproval, AgentRun, AgentRunEvent } from '~/types/agent';
import { agentRunPhaseLabel, agentRunStatusLabel } from '~/types/agent';

/**
 * Agent-docken - ett eget fönster bredvid målet.
 *
 * Tidigare låg chatt, körningshistorik och checkpoint-plan i samma
 * lodräta flöde som målets innehåll. Då trängdes undermålen nedanför
 * chatten och det krävdes mycket scrolling för att nå dem. Nu ligger
 * agentdelen i en dock med flikar, så att bara en del syns åt gången
 * och målets innehåll alltid är kvar på sin plats.
 */
const props = defineProps<{
	goalId: number;
	goalTitle: string;
	goalDescription?: string;
	run?: AgentRun | null;
	events: AgentRunEvent[];
	approvals: AgentApproval[];
}>();

const emit = defineEmits<{
	close: [];
	'conversation-started': [conversationId: string];
	decided: [];
	cancelled: [];
}>();

type Tab = 'chat' | 'run' | 'plan';

const activeTab = ref<Tab>('chat');

// Beslut som väntar på en människa - visas som badge på fliken Körning.
const pendingCount = computed(
	() => props.approvals.filter((item) => item.status === 'pending').length,
);

// En körning som är igång ska gå att se med en blick.
const statusTone = computed(() => {
	switch (props.run?.status) {
		case 'running':
			return 'text-blue-300 border-blue-500/40 bg-blue-500/10';
		case 'waiting_approval':
			return 'text-amber-300 border-amber-500/40 bg-amber-500/10';
		case 'failed':
			return 'text-red-300 border-red-500/40 bg-red-500/10';
		case 'succeeded':
			return 'text-green-300 border-green-500/40 bg-green-500/10';
		default:
			return 'text-ink-dim border-line bg-surface-2';
	}
});

const statusLabel = computed(() => {
	if (!props.run) return '';
	if (props.run.status === 'running') return agentRunPhaseLabel(props.run.phase);
	return agentRunStatusLabel(props.run.status);
});

const tabs = computed(() => [
	{ id: 'chat' as Tab, label: 'Chatt', icon: 'lucide:message-square' },
	{
		id: 'run' as Tab,
		label: 'Körning',
		icon: 'lucide:activity',
		badge: pendingCount.value,
	},
	{ id: 'plan' as Tab, label: 'Plan', icon: 'lucide:route' },
]);
</script>

<template>
	<div
		class="flex min-h-0 flex-col border-l border-line bg-surface shadow-2xl"
		:class="
			// Docken är en fast kolumn längst till höger. På breda skärmar
			// ligger den ovanför kalenderspalten, så att målets innehåll
			// behåller sin fulla bredd. På små skärmar lägger den sig
			// över sidan och skjuter undan innehållet.
			'fixed inset-x-0 bottom-12 top-0 z-40 w-full lg:bottom-0 lg:left-auto lg:right-0 lg:w-[24rem] xl:w-[26rem]'
		"
	>
		<header class="flex flex-shrink-0 items-center justify-between gap-2 border-b border-line px-4 py-2.5">
			<div class="flex min-w-0 items-center gap-2">
				<Icon name="lucide:sparkles" class="h-4 w-4 flex-shrink-0 text-accent" />
				<h2 class="truncate text-sm font-medium text-ink">Agent</h2>
				<span
					v-if="run"
					class="flex-shrink-0 rounded-full border px-2 py-0.5 text-[11px]"
					:class="statusTone"
				>
					{{ statusLabel }}
				</span>
			</div>
			<button
				type="button"
				class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink"
				title="Stäng agentpanelen"
				@click="emit('close')"
			>
				<Icon name="lucide:x" class="h-4 w-4" />
			</button>
		</header>

		<!-- Flikar - bara en del av agenten syns åt gången -->
		<nav class="flex flex-shrink-0 items-center gap-1 border-b border-line px-2 pt-1.5">
			<button
				v-for="tab in tabs"
				:key="tab.id"
				type="button"
				class="flex items-center gap-1.5 border-b-2 px-3 py-2 text-xs transition-colors"
				:class="
					activeTab === tab.id
						? 'border-accent text-ink'
						: 'border-transparent text-ink-dim hover:text-ink-muted'
				"
				:aria-current="activeTab === tab.id"
				@click="activeTab = tab.id"
			>
				<Icon :name="tab.icon" class="h-3.5 w-3.5" />
				{{ tab.label }}
				<span
					v-if="tab.badge"
					class="rounded-full bg-amber-500/20 px-1.5 text-[10px] text-amber-300"
				>
					{{ tab.badge }}
				</span>
			</button>
		</nav>

		<!-- Innehållet. Panelena behålls monterade men dolda, så att
		     chatten inte tappar sin historik när man byter flik. -->
		<div class="flex min-h-0 flex-1 flex-col">
			<AgentChatPanel
				v-show="activeTab === 'chat'"
				:goal-id="goalId"
				:goal-title="goalTitle"
				:goal-description="goalDescription"
				:run="run"
				@conversation-started="emit('conversation-started', $event)"
			/>
			<!-- Körningen: verifieringen ligger överst, sedan historiken. -->
			<div v-show="activeTab === 'run'" class="flex min-h-0 flex-1 flex-col">
				<AgentVerificationState :goal-id="goalId" @changed="emit('decided')" />
				<AgentRunHistory
					class="min-h-0 flex-1"
					:events="events"
					:run="run"
					:approvals="approvals"
					@decided="emit('decided')"
					@cancelled="emit('cancelled')"
				/>
			</div>
			<AgentCheckpointWorkflowPanel
				v-show="activeTab === 'plan'"
				:feature="goalTitle"
				:description="goalDescription || ''"
			/>
		</div>
	</div>
</template>

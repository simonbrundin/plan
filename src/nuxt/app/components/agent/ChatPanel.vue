<script setup lang="ts">
import type { AgentConversation, AgentMessage, AgentRun } from '~/types/agent';
import { renderMarkdown } from '~/utils/markdown';

// Chattpanelen - Plan är agentens gränssnitt. Slack används inte.
const props = defineProps<{
	goalId: number;
	goalTitle: string;
	goalDescription?: string;
	run?: AgentRun | null;
}>();

const emit = defineEmits<{
	(event: 'conversation-started', conversationId: string): void;
}>();

const {
	messages,
	conversationId,
	conversations,
	approval,
	streaming,
	error,
	available,
	load,
	loadApproval,
	loadLatestForGoal,
	openConversation,
	startConversation,
	send,
	decide,
} = useAgentChat(toRef(props, 'goalId'));

const draft = ref('');
const listRef = ref<HTMLElement | null>(null);
const inputRef = ref<HTMLTextAreaElement | null>(null);
const feedback = ref('');
let approvalTimer: ReturnType<typeof setInterval> | undefined;

// Ett förslag om målets utseende visar varje fält som föreslås ändrat.
// Ett kodförslag visar i stället diffen.
const isDefinition = computed(
	() => approval.value?.payload?.kind === 'goal_definition',
);

const approvalFields = computed(() => {
	if (!isDefinition.value) return [];
	return approval.value?.payload?.fields ?? [];
});

const approvalHeading = computed(() =>
	isDefinition.value
		? 'Agenten föreslår en definition:'
		: 'Agenten föreslår en ändring:',
);

// Sant när chatten är scrollad till slutet. Då ska nya meddelanden
// scrolla med - annars ska man kunna läsa i fred och röra sig själv.
const atBottom = ref(true);

// Verktygsanrop visas ihop - de är sällan det man vill läsa först.
const expandedTools = ref<Set<string>>(new Set());

const onScroll = () => {
	const el = listRef.value;
	if (!el) return;
	atBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
};

// Scrollar till slutet när det kommer nya meddelanden.
watch(
	() => messages.value.length,
	async () => {
		await nextTick();
		if (atBottom.value) scrollToBottom();
	},
);

const scrollToBottom = () => {
	listRef.value?.scrollTo({ top: listRef.value.scrollHeight, behavior: 'smooth' });
};

// Fältet växer med texten, men inte mer än några rader.
const resizeInput = () => {
	const el = inputRef.value;
	if (!el) return;
	el.style.height = 'auto';
	el.style.height = `${Math.min(el.scrollHeight, 140)}px`;
};

watch(draft, () => nextTick(resizeInput));

const toggleTool = (message: AgentMessage) => {
	const next = new Set(expandedTools.value);
	if (next.has(message.id)) next.delete(message.id);
	else next.add(message.id);
	expandedTools.value = next;
};

const formatTime = (value?: string) => {
	if (!value) return '';
	try {
		return new Intl.DateTimeFormat('sv-SE', { timeStyle: 'short' }).format(
			new Date(value),
		);
	} catch {
		return '';
	}
};

onMounted(async () => {
	await loadApproval();
	// Godkännandet hämtas om - agenten kan skapa det när som helst,
	// och en human gate som inte syns är ingen human gate.
	approvalTimer = setInterval(() => {
		if (document.visibilityState === 'visible') void loadApproval();
	}, 15000);
	// Samtalen för målet hämtas - de är den väg chatten hittas tillbaka
	// på. Finns inget samtal för målet används den körningens samtal.
	const senaste = await loadLatestForGoal(props.goalId);
	if (!senaste && props.run?.conversation_id) {
		conversationId.value = props.run.conversation_id;
		await load(props.run.conversation_id);
	}
});

onUnmounted(() => {
	if (approvalTimer) clearInterval(approvalTimer);
});

// Startar en konversation om det inte finns en.
const ensureConversation = async () => {
	if (conversationId.value) return conversationId.value;
	const id = await startConversation(props.goalTitle, props.goalId);
	if (id) {
		emit('conversation-started', id);
	}
	return id;
};

// Startar en ny chatt - den gamla behålls och kan öppnas igen.
const startNewConversation = async () => {
	const id = await startConversation(props.goalTitle, props.goalId);
	if (!id) return;
	messages.value = [];
	emit('conversation-started', id);
};

// Öppnar en tidigare chatt för målet.
const onSelectConversation = async (event: Event) => {
	const target = event.target as HTMLSelectElement;
	if (!target.value) return;
	await openConversation(target.value);
};

// Etikett för en chatt i listan - titel och tidpunkt.
const conversationLabel = (conversation: AgentConversation) => {
	const time = formatTime(conversation.updated_at ?? conversation.created_at);
	return time ? `${conversation.title} · ${time}` : conversation.title;
};

const submit = async () => {
	const text = draft.value;
	if (!text.trim() || streaming.value) return;

	draft.value = '';
	await nextTick();
	resizeInput();
	await ensureConversation();
	await send(text, {
		id: props.goalId,
		title: props.goalTitle,
		description: props.goalDescription,
	});
};

// Enter skickar, Shift+Enter ger en ny rad.
const onKeydown = (event: KeyboardEvent) => {
	if (event.key === 'Enter' && !event.shiftKey) {
		event.preventDefault();
		void submit();
	}
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
	<section class="flex min-h-0 flex-1 flex-col bg-surface">
		<!-- Rubriken med agentens status ligger i docken - här visas chattens
		     samtal och knappen för att starta en ny. -->
		<header
			class="flex flex-shrink-0 items-center justify-between gap-2 border-b border-line px-4 py-2"
		>
			<template v-if="conversationId">
				<select
					v-if="conversations.length > 1"
					:value="conversationId"
					class="min-w-0 flex-1 truncate rounded-lg border border-line bg-surface-2 px-2 py-1 text-xs text-ink"
					@change="onSelectConversation"
				>
					<option v-for="item in conversations" :key="item.id" :value="item.id">
						{{ conversationLabel(item) }}
					</option>
				</select>
				<p v-else class="text-xs text-ink-dim">Chatt med agenten</p>
				<button
					type="button"
					@click="startNewConversation"
					class="shrink-0 text-xs text-accent transition-colors hover:text-accent-light"
				>
					Ny chatt
				</button>
			</template>
			<template v-else>
				<p class="text-xs text-ink-dim">Ingen konversation ännu.</p>
				<button
					type="button"
					@click="ensureConversation"
					class="text-xs text-accent transition-colors hover:text-accent-light"
				>
					Starta chatt
				</button>
			</template>
		</header>

		<!-- Förslag som väntar på beslut -->
		<div v-if="approval" class="flex-shrink-0 border-b border-line bg-amber-500/5 px-4 py-3">
			<p class="mb-2 text-xs text-amber-300">{{ approvalHeading }}</p>
			<p class="text-sm font-medium text-ink">{{ approval.title }}</p>
			<p class="mt-1 whitespace-pre-wrap text-sm text-ink-muted">{{ approval.description }}</p>
			<ul v-if="approvalFields.length" class="mt-2 space-y-1 text-sm">
				<li v-for="field in approvalFields" :key="field.field">
					<span class="text-ink">{{ field.label }}</span>
					<span class="text-ink-dim">{{ field.from || 'ingen' }}</span>
					<span class="text-ink-dim">→</span>
					<span class="text-ink">{{ field.to }}</span>
				</li>
			</ul>
			<div class="mt-3 flex items-center gap-2">
				<button
					@click="approve"
					class="rounded-lg bg-accent-strong px-3 py-1.5 text-sm text-white transition-colors hover:bg-accent"
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
					class="rounded-lg border border-line bg-surface-2 px-3 py-1.5 text-sm text-ink-muted transition-colors hover:border-line-strong hover:text-ink"
				>
					Avvisa
				</button>
			</div>
		</div>

		<!-- Meddelanden -->
		<div class="relative min-h-0 flex-1">
			<div
				ref="listRef"
				class="h-full space-y-3 overflow-y-auto px-4 py-3"
				@scroll="onScroll"
			>
				<p v-if="!messages.length" class="text-sm text-ink-dim">
					Inga meddelanden än. Skriv vad du vill ha hjälp med - till exempel
					så här: "Definiera målet och dela upp det i checkpoints".
				</p>
				<div
					v-for="message in messages"
					:key="message.id"
					class="flex"
					:class="message.role === 'user' ? 'justify-end' : 'justify-start'"
				>
					<!-- Verktygsanrop: ihopfällt tills man vill se dem -->
					<div
						v-if="message.role === 'tool'"
						class="max-w-[90%] rounded-xl border border-line bg-surface-2 px-3 py-2 text-xs text-ink-dim"
					>
						<button
							type="button"
							class="flex w-full items-center gap-1.5 text-left font-mono"
							@click="toggleTool(message)"
						>
							<Icon
								:name="expandedTools.has(message.id) ? 'lucide:chevron-down' : 'lucide:chevron-right'"
								class="h-3 w-3 flex-shrink-0"
							/>
							<span class="truncate">{{ message.tool_name || 'verktyg' }}</span>
						</button>
						<pre
							v-if="expandedTools.has(message.id)"
							class="mt-2 max-h-48 overflow-auto whitespace-pre-wrap break-words text-[11px]"
						>{{ message.content }}</pre>
					</div>

					<!-- Vanliga meddelanden -->
					<div v-else class="max-w-[90%]">
						<div
							class="rounded-2xl px-3.5 py-2 text-sm break-words"
							:class="
								message.role === 'user'
									? 'bg-accent-strong text-white whitespace-pre-wrap'
									: message.role === 'system'
										? 'border border-line bg-surface-2 text-ink-muted whitespace-pre-wrap'
										: 'bg-surface-2 text-ink'
							"
						>
							<!-- Agentens svar är markdown och renderas som HTML -->
							<div
								v-if="message.role === 'assistant'"
								class="chat-markdown"
								v-html="renderMarkdown(message.content)"
							/>
							<template v-else>{{ message.content }}</template>
						</div>
						<p
							class="mt-0.5 px-1 text-[10px] text-ink-dim"
							:class="message.role === 'user' ? 'text-right' : 'text-left'"
						>
							{{ message.role === 'user' ? 'Du' : 'Agenten' }}
							<template v-if="formatTime(message.created_at)">
								· {{ formatTime(message.created_at) }}
							</template>
						</p>
					</div>
				</div>
				<p v-if="streaming" class="text-xs text-ink-dim">Agenten skriver…</p>
			</div>

			<!-- Knapp som tar ner till de senaste meddelandena -->
			<button
				v-if="!atBottom && messages.length"
				type="button"
				class="absolute bottom-3 left-1/2 flex -translate-x-1/2 items-center gap-1.5 rounded-full border border-line bg-elevated px-3 py-1.5 text-xs text-ink-muted shadow-lg transition-colors hover:text-ink"
				@click="scrollToBottom"
			>
				<Icon name="lucide:arrow-down" class="h-3.5 w-3.5" />
				Nya meddelanden
			</button>
		</div>

		<p v-if="error" class="flex-shrink-0 px-4 pb-2 text-xs text-red-300">{{ error }}</p>
		<p v-else-if="!available" class="flex-shrink-0 px-4 pb-2 text-xs text-ink-dim">
			Agent-API:t är inte nåbar just nu.
		</p>

		<!-- Skrivfält -->
		<form
			class="flex flex-shrink-0 items-end gap-2 border-t border-line px-4 py-3"
			@submit.prevent="submit"
		>
			<textarea
				ref="inputRef"
				v-model="draft"
				rows="1"
				placeholder="Skriv till agenten…"
				class="max-h-36 flex-1 resize-none rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-sm text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
				@keydown="onKeydown"
			/>
			<button
				type="submit"
				:disabled="streaming || !draft.trim()"
				class="rounded-lg bg-accent-strong px-3.5 py-2 text-sm text-white transition-colors hover:bg-accent disabled:opacity-40"
			>
				Skicka
			</button>
		</form>
		<p class="flex-shrink-0 px-4 pb-2 text-[10px] text-ink-dim">
			Enter skickar · Shift+Enter ger ny rad
		</p>
	</section>
</template>

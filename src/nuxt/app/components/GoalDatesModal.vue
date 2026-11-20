<script setup lang="ts">
import { closeGoalDatesModal } from "~/composables/useGoalDatesModal";

/**
 * Popup där både startdatum och due-datum sätts för ett mål.
 *
 * Öppnas med leader + d. Startdatumet anger från vilket datum målet får
 * dyka upp i prioriteringslistan, due-datumet när målet ska vara klart.
 *
 * Popupen skickar båda värdena, men sidan sparar bara de fält som
 * ändrats - ett tomt fält ska inte nollställa det andra.
 */
const props = defineProps<{
	/** Om popupen är synlig. */
	open: boolean;
	/** Titeln på målet datumen gäller. */
	goalTitle?: string;
	/** Målets nuvarande startdatum ("YYYY-MM-DD"). */
	startDate?: string | null;
	/** Målets nuvarande due-datum ("YYYY-MM-DD"). */
	dueDate?: string | null;
}>();

const emit = defineEmits<{
	/** Användaren sparade båda datumen. */
	save: [payload: { startDate: string | null; dueDate: string | null }];
	/** Användaren avbröt. */
	close: [];
}>();

const startDraft = ref(props.startDate ?? "");
const dueDraft = ref(props.dueDate ?? "");

// Varje gång popupen öppnas ska fälten visa målets aktuella datum.
watch(
	() => [props.open, props.startDate, props.dueDate] as const,
	([open]) => {
		if (!open) return;
		startDraft.value = props.startDate ?? "";
		dueDraft.value = props.dueDate ?? "";
	},
);

function save() {
	emit("save", {
		startDate: startDraft.value ? startDraft.value : null,
		dueDate: dueDraft.value ? dueDraft.value : null,
	});
}

function close() {
	// Stäng den delade popup-state:n direkt. Eventet skickas också så
	// att layouten kan reagera, men popupen ska inte vara beroende av
	// att parent-handlern lyckas köras.
	closeGoalDatesModal();
	emit("close");
}

// Escape stänger popupen - tangenten ska inte också nå sidans egna
// tangentbevakare (t.ex. för att stänga en dialogruta).
function handleKeydown(event: KeyboardEvent) {
	if (event.key === "Escape") {
		event.preventDefault();
		event.stopPropagation();
		close();
	}
}

onMounted(() => window.addEventListener("keydown", handleKeydown, true));
onUnmounted(() => window.removeEventListener("keydown", handleKeydown, true));
</script>

<template>
	<Teleport to="body">
		<div
			v-if="props.open"
			class="fixed inset-0 z-[100] flex items-center justify-center"
			style="background: rgba(0, 0, 0, 0.7)"
			@click.self="close"
		>
			<div class="bg-surface border border-line rounded-lg shadow-lg p-6 w-96">
				<h2 class="text-lg font-bold text-white">Datum</h2>
				<p v-if="props.goalTitle" class="text-sm text-ink-dim mt-1 mb-4">
					{{ props.goalTitle }}
				</p>

				<label class="block mb-4">
					<span class="text-sm text-ink">Startdatum</span>
					<span class="block text-xs text-ink-dim mt-0.5 mb-1">
						Målet får tidigast dyka upp i prioriteringslistan från detta
						datum. Tomt = målet syns direkt.
					</span>
					<input
						v-model="startDraft"
						type="date"
						class="w-full rounded border border-line bg-surface-2 px-3 py-1.5 text-sm text-ink focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
						@keydown.enter.prevent="save"
					/>
				</label>

				<label class="block mb-5">
					<span class="text-sm text-ink">Deadline</span>
					<span class="block text-xs text-ink-dim mt-0.5 mb-1">
						När målet ska vara klart. Tomt = ingen deadline.
					</span>
					<input
						v-model="dueDraft"
						type="date"
						class="w-full rounded border border-line bg-surface-2 px-3 py-1.5 text-sm text-ink focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
						@keydown.enter.prevent="save"
					/>
				</label>

				<div class="flex gap-2 justify-end">
					<button
						type="button"
						class="px-3 py-1 text-sm rounded bg-accent text-white hover:opacity-90"
						@click="save"
					>
						Spara
					</button>
					<button
						type="button"
						class="px-3 py-1 text-sm rounded text-ink-dim hover:text-ink"
						@click.stop="close"
					>
						Avbryt
					</button>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import type { Place } from "~/types/place";

// Väljare som sätter platser på ett mål.
//
// Öppnas med leader + p på alla sidor där ett mål är markerat. Varje
// ändring sparas direkt - listan skickas ut när en plats kryssas i eller
// ur.
const props = defineProps<{
	open: boolean;
	/** Titeln på målet platserna sätts på. */
	targetTitle: string;
	/** Platser som redan sitter på målet. */
	selected: Place[];
	/** Alla platser användaren har. */
	available: Place[];
}>();

const emit = defineEmits<{
	/** Användaren har ändrat platslistan. */
	"update:selected": [places: Place[]];
	/** Användaren vill stänga väljaren. */
	close: [];
}>();

const draft = ref<Place[]>([...props.selected]);

watch(
	() => props.selected,
	(places) => {
		draft.value = [...places];
	},
);

function isSelected(placeId: number): boolean {
	return draft.value.some((p) => p.id === placeId);
}

function toggle(place: Place) {
	if (isSelected(place.id)) {
		draft.value = draft.value.filter((p) => p.id !== place.id);
	} else {
		draft.value = [...draft.value, place];
	}
	emit("update:selected", draft.value);
}

function handleKeydown(event: KeyboardEvent) {
	if (!props.open) return;
	if (event.key === "Escape") {
		event.preventDefault();
		emit("close");
	}
}

// Escape måste fångas globalt - modalen har inte nödvändigtvis fokus
// när den öppnas.
onMounted(() => {
	window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
	window.removeEventListener("keydown", handleKeydown);
});
</script>

<template>
	<div
		v-if="open"
		class="fixed inset-0 z-[70] flex items-start justify-center bg-black/50 pt-[12vh]"
		@click.self="emit('close')"
	>
		<div
			class="w-full max-w-md mx-4 bg-surface rounded-2xl border border-line shadow-xl overflow-hidden"
			@click.stop
		>
			<div class="px-4 py-3 border-b border-line">
				<h2 class="text-sm font-medium text-ink">
					Platser för "{{ targetTitle }}"
				</h2>
				<p class="text-xs text-ink-dim mt-0.5">
					Var målet kan slutföras. Mål utan plats är görbara överallt.
					Esc stänger.
				</p>
			</div>

			<div v-if="available.length === 0" class="px-4 py-6 text-sm text-ink-dim">
				Inga platser ännu. Skapa en plats på sidan Platser.
			</div>

			<ul v-else class="max-h-[50vh] overflow-y-auto px-2 py-2 space-y-1">
				<li v-for="place in available" :key="place.id">
					<label
						class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer hover:bg-elevated transition-colors"
					>
						<input
							type="checkbox"
							:checked="isSelected(place.id)"
							@change="toggle(place)"
							class="accent-accent"
						/>
						<span class="text-sm text-ink">{{ place.name }}</span>
						<span class="text-xs text-ink-dim ml-auto flex-shrink-0">
							{{ place.radius_meters }} m
						</span>
					</label>
				</li>
			</ul>
		</div>
	</div>
</template>

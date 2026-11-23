<script setup lang="ts">
import type { Tag } from "~/types/tag";

// Väljare som sätter taggar på ett mål.
//
// Tangenten 'g' på målsidan öppnar väljaren. Man skriver ett taggnamn
// och trycker Enter: finns taggen redan läggs den på (eller tas bort),
// annars skapas den. Backspace i tomt fält tar bort den sista taggen.
const props = defineProps<{
	open: boolean;
	/** Titeln på målet taggarna sätts på. */
	targetTitle: string;
	/** Taggar som redan sitter på målet. */
	selected: string[];
	/** Alla taggar användaren har. */
	available: Tag[];
}>();

const emit = defineEmits<{
	/** Användaren har ändrat tagglistan. */
	"update:selected": [tags: string[]];
	/** Användaren vill stänga väljaren. */
	close: [];
}>();

const query = ref("");
const selectedIndex = ref(0);
const searchInput = ref<HTMLInputElement | null>(null);

// Alla taggar som visas: användarens egna plus de som redan sitter på
// målet (de kan saknas i listan om de just skapats).
const allTags = computed<Tag[]>(() => {
	const names = new Map<string, Tag>();
	for (const tag of props.available) {
		names.set(tag.name.toLowerCase(), tag);
	}
	for (const name of props.selected) {
		if (!names.has(name.toLowerCase())) {
			names.set(name.toLowerCase(), { name, goal_count: 0 });
		}
	}
	return [...names.values()].sort((a, b) => a.name.localeCompare(b.name, "sv"));
});

// Taggar som matchar sökningen.
const results = computed<Tag[]>(() => {
	const needle = query.value.trim().toLowerCase();
	if (!needle) return allTags.value;
	return allTags.value.filter((tag) => tag.name.toLowerCase().includes(needle));
});

function isSelected(name: string): boolean {
	return props.selected.some((tag) => tag.toLowerCase() === name.toLowerCase());
}

// Lägger till eller tar bort en tagg. listan skickas ut direkt - varje
// ändring sparas.
function toggleTag(name: string) {
	if (isSelected(name)) {
		emit(
			"update:selected",
			props.selected.filter((tag) => tag.toLowerCase() !== name.toLowerCase()),
		);
		return;
	}
	emit("update:selected", [...props.selected, name]);
}

function handleKeydown(event: KeyboardEvent) {
	if (event.key === "ArrowDown") {
		event.preventDefault();
		if (results.value.length > 0) {
			selectedIndex.value = Math.min(selectedIndex.value + 1, results.value.length - 1);
		}
		return;
	}

	if (event.key === "ArrowUp") {
		event.preventDefault();
		selectedIndex.value = Math.max(selectedIndex.value - 1, 0);
		return;
	}

	if (event.key === "Enter") {
		event.preventDefault();
		const trimmed = query.value.trim();
		const exact = allTags.value.find(
			(tag) => tag.name.toLowerCase() === trimmed.toLowerCase(),
		);

		if (exact) {
			// Taggen finns - lägg på eller ta bort.
			toggleTag(exact.name);
			query.value = "";
			return;
		}

		if (trimmed) {
			// Ny tagg - den skapas av API:t.
			emit("update:selected", [...props.selected, trimmed]);
			query.value = "";
		}
		// Tomt fält gör ingenting - en oavsiktlig tangent tryck ska inte
		// lägga till en tagg. Taggar tas bort genom att skriva taggnamnet
		// och trycka Enter igen (eller genom krysset på taggen).
		return;
	}

	if (event.key === "Backspace" && query.value === "" && props.selected.length > 0) {
		event.preventDefault();
		emit("update:selected", props.selected.slice(0, -1));
		return;
	}

	if (event.key === "Escape") {
		event.preventDefault();
		emit("close");
	}
}

watch(
	() => props.open,
	async (open) => {
		if (!open) return;
		query.value = "";
		selectedIndex.value = 0;
		await nextTick();
		searchInput.value?.focus();
	},
);

// En ny sökning börjar alltid överst i listan.
watch(query, () => {
	selectedIndex.value = 0;
});

// Håll markeringen inom listan när sökresultatet ändras.
watch(results, () => {
	if (selectedIndex.value >= results.value.length) {
		selectedIndex.value = Math.max(results.value.length - 1, 0);
	}
});
</script>

<template>
	<div
		v-if="open"
		class="fixed inset-0 z-[70] flex items-start justify-center bg-black/50 pt-[12vh]"
		@click.self="emit('close')"
	>
		<div
			class="w-full max-w-lg mx-4 bg-surface rounded-2xl border border-line shadow-xl overflow-hidden"
			@click.stop
		>
			<div class="px-4 py-3 border-b border-line">
				<h2 class="text-sm font-medium text-ink">Taggar för "{{ targetTitle }}"</h2>
				<p class="text-xs text-ink-dim mt-0.5">
					Skriv ett taggnamn och tryck Enter. Finns taggen läggs den på,
					annars skapas den. Esc stänger.
				</p>
			</div>

			<div v-if="selected.length > 0" class="px-3 pt-3">
				<h3 class="px-1 mb-1 text-xs font-medium uppercase tracking-wide text-ink-dim">
					På målet
				</h3>
				<div class="flex flex-wrap gap-1.5">
					<span
						v-for="tag in selected"
						:key="tag"
						class="inline-flex items-center gap-1 rounded-full bg-accent/15 px-2 py-0.5 text-xs text-accent"
					>
						{{ tag }}
						<button
							type="button"
							class="rounded-full p-0.5 transition-colors hover:bg-accent/25"
							:title="`Ta bort ${tag}`"
							@click.stop="toggleTag(tag)"
						>
							<Icon name="lucide:x" class="w-3 h-3" />
						</button>
					</span>
				</div>
			</div>

			<div class="p-3">
				<input
					ref="searchInput"
					v-model="query"
					type="text"
					placeholder="Tagg, t.ex. ute..."
					class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full text-sm"
					@keydown="handleKeydown"
				/>
			</div>

			<ul v-if="results.length > 0" class="max-h-[40vh] overflow-y-auto px-2 pb-2 space-y-1">
				<li v-for="(tag, index) in results" :key="tag.name">
					<button
						type="button"
						@click.prevent.stop="toggleTag(tag.name)"
						@mouseenter="selectedIndex = index"
						class="w-full text-left px-3 py-2 rounded-lg transition-colors flex items-center gap-2 border"
						:class="index === selectedIndex ? 'bg-elevated border-accent' : 'border-transparent hover:bg-elevated'"
					>
						<Icon
							name="lucide:tag"
							class="w-4 h-4 flex-shrink-0"
							:class="isSelected(tag.name) ? 'text-accent' : 'text-ink-dim'"
						/>
						<span class="text-ink text-sm truncate">{{ tag.name }}</span>
						<span v-if="tag.goal_count > 0" class="text-xs text-ink-dim ml-auto flex-shrink-0">
							{{ tag.goal_count }} mål
						</span>
					</button>
				</li>
			</ul>

			<div v-else class="px-4 pb-4 text-sm text-ink-dim">
				Inga matchande taggar - tryck Enter för att skapa "{{ query.trim() }}"
			</div>
		</div>
	</div>
</template>

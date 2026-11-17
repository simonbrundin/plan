<script setup lang="ts">
import { EMOJI_CATALOGUE, searchEmojis } from "~/utils/emojis";

interface Props {
	modelValue?: string;
	open?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	modelValue: "",
	open: false,
});

const emit = defineEmits<{
	"update:modelValue": [value: string];
	"update:open": [value: boolean];
}>();

type Tab = "emoji" | "icon";

const tab = ref<Tab>("emoji");
const emojiQuery = ref("");
const searchInput = ref<HTMLInputElement | null>(null);

const { searchQuery, searchResults, isSearching, searchError, reset } = useIconSearch();

const displayOpen = computed({
	get: () => props.open,
	set: (value: boolean) => emit("update:open", value),
});

const emojiResults = computed(() => {
	if (!emojiQuery.value.trim()) return EMOJI_CATALOGUE;
	return searchEmojis(emojiQuery.value);
});

watch(displayOpen, async (newValue) => {
	if (newValue) {
		await nextTick();
		searchInput.value?.focus();
	} else {
		emojiQuery.value = "";
		reset();
	}
});

function selectEmoji(emoji: string) {
	emit("update:modelValue", emoji);
	close();
}

function selectIcon(full: string) {
	emit("update:modelValue", full);
	close();
}

function clearIcon() {
	emit("update:modelValue", "");
	close();
}

function close() {
	displayOpen.value = false;
	emojiQuery.value = "";
	reset();
}
</script>

<template>
	<div class="relative inline-block">
		<div v-if="displayOpen" class="fixed inset-0 bg-black/50 z-40" @click="close" />

		<div
			v-if="displayOpen"
			class="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-2xl mx-auto"
		>
			<div class="bg-surface-2 rounded-lg shadow-lg border border-line p-4">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-xl font-bold text-ink">Välj ikon</h2>
					<button
						type="button"
						class="text-ink-muted hover:text-ink p-1 rounded hover:bg-elevated"
						aria-label="Stäng"
						@click="close"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-6 w-6"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>

				<div class="flex gap-1 mb-4" role="tablist" aria-label="Typ av ikon">
					<button
						type="button"
						role="tab"
						:aria-selected="tab === 'emoji'"
						class="rounded-lg px-3 py-1.5 text-sm"
						:class="tab === 'emoji' ? 'bg-accent-strong text-white' : 'bg-elevated text-ink-muted'"
						@click="tab = 'emoji'"
					>
						Emoji
					</button>
					<button
						type="button"
						role="tab"
						:aria-selected="tab === 'icon'"
						class="rounded-lg px-3 py-1.5 text-sm"
						:class="tab === 'icon' ? 'bg-accent-strong text-white' : 'bg-elevated text-ink-muted'"
						@click="tab = 'icon'"
					>
						Ikoner
					</button>
				</div>

				<div v-if="tab === 'emoji'">
					<input
						ref="searchInput"
						v-model="emojiQuery"
						type="text"
						placeholder="Sök emoji, t.ex. träning"
						class="w-full px-4 py-2 bg-surface border border-line rounded-lg text-ink placeholder-ink-dim focus:outline-none focus:ring-2 focus:ring-accent mb-4"
					/>

					<div
						v-if="emojiResults.length > 0"
						class="grid grid-cols-8 gap-2 max-h-96 overflow-y-auto"
					>
						<button
							v-for="entry in emojiResults"
							:key="entry.emoji"
							type="button"
							class="aspect-square flex items-center justify-center rounded-lg p-2 text-2xl hover:bg-elevated transition-colors"
							:class="modelValue === entry.emoji ? 'bg-accent-strong' : 'bg-elevated'"
							:title="entry.keywords.join(', ')"
							@click="selectEmoji(entry.emoji)"
						>
							{{ entry.emoji }}
						</button>
					</div>

					<div v-else class="text-center py-8 text-ink-muted">
						Ingen emoji matchade "{{ emojiQuery }}"
					</div>
				</div>

				<div v-else>
					<input
						v-model="searchQuery"
						type="text"
						placeholder="Sök ikoner..."
						class="w-full px-4 py-2 bg-surface border border-line rounded-lg text-ink placeholder-ink-dim focus:outline-none focus:ring-2 focus:ring-accent mb-4"
					/>

					<div
						v-if="!searchQuery || searchQuery.trim().length < 2"
						class="text-center py-8 text-ink-muted"
					>
						Skriv minst 2 tecken för att söka bland alla ikoner
					</div>

					<div v-else-if="isSearching" class="text-center py-8 text-ink-muted">
						Söker ikoner...
					</div>

					<div
						v-else-if="searchResults.length > 0"
						class="grid grid-cols-6 gap-2 max-h-96 overflow-y-auto"
					>
						<button
							v-for="icon in searchResults"
							:key="icon.full"
							type="button"
							class="aspect-square flex items-center justify-center rounded-lg p-2 hover:bg-elevated transition-colors"
							:class="modelValue === icon.full ? 'bg-accent-strong' : 'bg-elevated'"
							@click="selectIcon(icon.full)"
						>
							<Icon :name="icon.full" class="w-8 h-8 text-white" />
						</button>
					</div>

					<div v-else-if="searchError" class="text-center py-8 text-red-400">
						{{ searchError }}
					</div>

					<div v-else class="text-center py-8 text-ink-muted">
						Ingen ikon matchade "{{ searchQuery }}"
					</div>
				</div>

				<div class="flex justify-between items-center mt-4 pt-3 border-t border-line">
					<button
						type="button"
						class="text-sm text-ink-muted hover:text-ink"
						@click="clearIcon"
					>
						Ingen ikon
					</button>
					<button type="button" class="text-sm text-ink-muted hover:text-ink" @click="close">
						Avbryt
					</button>
				</div>
			</div>
		</div>
	</div>
</template>

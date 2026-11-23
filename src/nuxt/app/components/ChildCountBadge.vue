<script setup lang="ts">
// Antalet undermål som ett mål har.
//
// Antalet räknas ur målgraphens relationer, så att det är aktuellt på
// alla sidor utan att varje rad behöver fråga servern.
const props = defineProps<{
	/** Id:t på målet som antalet ska visas för. */
	goalId: number;
	/** Storlek på märket. */
	size?: "sm" | "md";
}>();

const goalsStore = useGoalsStore();

const stats = computed(
	() => goalsStore.childStats.get(props.goalId) ?? { total: 0, finished: 0 },
);

const iconClass = computed(() => (props.size === "md" ? "w-4 h-4" : "w-3.5 h-3.5"));
const textClass = computed(() => (props.size === "md" ? "text-sm" : "text-xs"));

const tooltip = computed(() => {
	if (stats.value.finished > 0) {
		return `${stats.value.total} undermål (${stats.value.finished} klara)`;
	}
	return `${stats.value.total} undermål`;
});
</script>

<template>
	<span
		v-if="stats.total > 0"
		class="inline-flex items-center gap-0.5 flex-shrink-0 text-ink-dim"
		:title="tooltip"
		:aria-label="tooltip">
		<Icon name="lucide:git-fork" :class="iconClass" />
		<span class="tabular-nums" :class="textClass">{{ stats.total }}</span>
	</span>
</template>

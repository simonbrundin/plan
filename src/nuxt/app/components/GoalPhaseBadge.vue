<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { GOAL_PHASES, getGoalPhasePresentation } from "~/utils/goalPhase";

const props = defineProps<{
	goalId: number;
	statusId: number;
}>();

const emit = defineEmits<{
	updated: [statusId: number];
}>();

const goalsStore = useGoalsStore();
const { updateGoalStatus } = useGoalApi();
const selectElement = ref<HTMLSelectElement | null>(null);
const isSaving = ref(false);
const errorMessage = ref("");

const currentStatus = computed(() =>
	goalsStore.statuses.find((status) => status.id === props.statusId),
);
const presentation = computed(() =>
	getGoalPhasePresentation(currentStatus.value?.name),
);
const options = computed(() =>
	GOAL_PHASES.flatMap((phase) => {
		const status = goalsStore.statuses.find((entry) => entry.name === phase);
		return status ? [{ id: status.id, ...getGoalPhasePresentation(phase) }] : [];
	}),
);

onMounted(() => {
	void goalsStore.loadStatuses().catch((error: unknown) => {
		console.error("Kunde inte hämta målfaser:", error);
	});
});

async function changePhase(event: Event): Promise<void> {
	const select = event.target as HTMLSelectElement;
	const nextStatusId = Number(select.value);
	if (!Number.isInteger(nextStatusId) || nextStatusId === props.statusId) return;

	isSaving.value = true;
	errorMessage.value = "";
	try {
		await updateGoalStatus(props.goalId, nextStatusId);
		emit("updated", nextStatusId);
	} catch (error) {
		select.value = String(props.statusId);
		errorMessage.value =
			error instanceof Error ? error.message : "Kunde inte ändra målets fas";
	} finally {
		isSaving.value = false;
	}
}
</script>

<template>
	<span
		class="relative inline-flex h-6 w-6 flex-shrink-0 cursor-pointer items-center justify-center rounded transition-colors hover:bg-surface-2 focus-within:ring-2 focus-within:ring-accent/50"
		:class="presentation.colorClass"
		:title="errorMessage || `Fas: ${presentation.label}`"
		:aria-busy="isSaving"
	>
		<Icon :name="presentation.icon" class="h-4 w-4" aria-hidden="true" />
		<select
			ref="selectElement"
			:value="statusId"
			:aria-label="`Ändra målfas: ${presentation.label}`"
			:disabled="isSaving || options.length !== GOAL_PHASES.length"
			class="absolute inset-0 h-full w-full cursor-pointer opacity-0 disabled:cursor-wait"
			@click.stop
			@change.stop="changePhase"
		>
			<option v-for="option in options" :key="option.name" :value="option.id">
				{{ option.label }}
			</option>
		</select>
		<span v-if="errorMessage" class="sr-only" role="alert">{{ errorMessage }}</span>
	</span>
</template>

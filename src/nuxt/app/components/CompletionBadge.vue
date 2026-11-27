<script setup lang="ts">
import { describeCompletion } from '~/utils/goalVerification';

/**
 * Visar hur ett mål avslutades - "Verifierat" eller "Deklarerat".
 *
 * Skillnaden är det enda som skiljer "jag tror det är klart" från
 * "det här är bevisat", så den ska gå att se direkt i listan.
 */
const props = defineProps<{
	/** Målets completion_kind. */
	completionKind?: string | null;
}>();

const completion = computed(() => describeCompletion(props.completionKind));
</script>

<template>
	<span
		v-if="completion"
		class="inline-flex items-center gap-1 rounded-full border border-line px-2 py-0.5 text-[11px] text-ink-dim"
		:title="completion.description"
		data-testid="completion-badge"
	>
		<Icon
			:name="completion.label === 'Verifierat' ? 'lucide:badge-check' : 'lucide:flag'"
			class="h-3 w-3"
		/>
		{{ completion.label }}
	</span>
</template>

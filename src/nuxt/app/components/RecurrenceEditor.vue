<script setup lang="ts">
import type { RecurrenceMode } from "~/utils/recurrence";
import { describeGoalRecurrence } from "~/utils/recurrence";
import { parseNaturalLanguage } from "~/utils/naturalLanguage";

const props = defineProps<{
	/** Den nuvarande regeln, som RRULE-sträng. */
	rule: string | null;
	/** Den nuvarande schemaläggningsstrategin. */
	mode: RecurrenceMode;
}>();

const emit = defineEmits<{
	/** Användaren sparar en ny regel och strategi. */
	save: [rule: string, mode: RecurrenceMode];
	/** Användaren tar bort upprepningen. */
	remove: [];
	/** Användaren avbryter. */
	cancel: [];
}>();

const ruleText = ref(props.rule ?? "");
const mode = ref<RecurrenceMode>(props.mode);
const error = ref("");

watch(
	() => props.rule,
	(value) => {
		ruleText.value = value ?? "";
	},
);

// Regeln kan skrivas antingen som naturligt språk ("varannan dag") eller
// som en RRULE-sträng ("FREQ=DAILY;INTERVAL=2").
const parsedRule = computed<string | null>(() => {
	const text = ruleText.value.trim();
	if (!text) return null;

	if (/^(RRULE:)?FREQ=/i.test(text)) {
		return text.replace(/^RRULE:/i, "");
	}

	return parseNaturalLanguage(text).recurrenceRule;
});

const preview = computed(() =>
	parsedRule.value ? describeGoalRecurrence(parsedRule.value) : "",
);

function save() {
	if (!parsedRule.value) {
		error.value =
			'Kunde inte tolka upprepningen. Prova "varannan dag" eller "FREQ=DAILY;INTERVAL=2".';
		return;
	}

	error.value = "";
	emit("save", parsedRule.value, mode.value);
}
</script>

<template>
	<div class="bg-surface-2 border border-accent rounded p-3 shadow-lg w-80">
		<input
			v-model="ruleText"
			type="text"
			placeholder="varannan dag"
			class="w-full rounded border border-line bg-surface px-3 py-1.5 text-sm text-ink placeholder-ink-dim focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
			@keydown.enter.prevent="save"
			@keydown.esc.prevent="emit('cancel')"
		/>

		<p v-if="preview" class="text-xs text-ink-dim mt-2">
			Upprepas {{ preview }}
		</p>
		<p v-else-if="ruleText.trim()" class="text-xs text-ink-dim mt-2">
			Kunde inte tolka upprepningen
		</p>

		<div class="mt-3 space-y-1">
			<label
				class="flex items-start gap-2 px-2 py-1 rounded cursor-pointer hover:bg-surface-3"
			>
				<input
					v-model="mode"
					type="radio"
					value="schedule"
					class="accent-accent mt-0.5"
				/>
				<span class="text-sm">
					Håll schemat
					<span class="block text-xs text-ink-dim">
						Förekomsterna ligger fast, förfallna hoppas över
					</span>
				</span>
			</label>
			<label
				class="flex items-start gap-2 px-2 py-1 rounded cursor-pointer hover:bg-surface-3"
			>
				<input
					v-model="mode"
					type="radio"
					value="push"
					class="accent-accent mt-0.5"
				/>
				<span class="text-sm">
					Skjuta fram
					<span class="block text-xs text-ink-dim">
						Nästa förekomst räknas från när du slutförde den senaste
					</span>
				</span>
			</label>
		</div>

		<p v-if="error" class="text-xs text-red-400 mt-2">{{ error }}</p>

		<div class="flex gap-2 mt-3">
			<button
				class="px-3 py-1 text-sm rounded bg-accent text-white hover:opacity-90"
				@click="save"
			>
				Spara
			</button>
			<button
				v-if="rule"
				class="px-3 py-1 text-sm rounded text-red-400 hover:text-red-300"
				@click="emit('remove')"
			>
				Ta bort
			</button>
			<button
				class="px-3 py-1 text-sm rounded text-ink-dim hover:text-ink"
				@click="emit('cancel')"
			>
				Avbryt
			</button>
		</div>
	</div>
</template>

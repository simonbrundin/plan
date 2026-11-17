<script setup lang="ts">
import { isIconifyIcon } from "~/utils/icons";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
	defineProps<{
		/** Emoji (t.ex. "🏃") eller Iconify-namn (t.ex. "heroicons:star"). */
		icon?: string | null;
	}>(),
	{ icon: "" },
);

const attrs = useAttrs();
</script>

<template>
	<Icon
		v-if="isIconifyIcon(props.icon)"
		:name="props.icon"
		v-bind="attrs"
		aria-hidden="true"
	/>
	<span v-else v-bind="attrs" aria-hidden="true">{{ props.icon }}</span>
</template>

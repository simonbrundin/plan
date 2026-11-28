<script setup lang="ts">
import type { DocMeta } from "~/utils/docs";

/**
 * Sidomeny för dokumentationen. Visar alla undersidor, grupperade.
 * Den aktuella sidan markeras, så att man ser var man befinner sig.
 */
const props = defineProps<{
  groups: [string, DocMeta[]][];
  current: string;
}>();

function isActive(slug: string): boolean {
  return props.current === slug;
}
</script>

<template>
  <nav class="space-y-6" aria-label="Dokumentationssidor">
    <div v-for="[group, docs] in groups" :key="group">
      <p class="mb-2 text-xs font-semibold uppercase tracking-widest text-ink-dim">
        {{ group }}
      </p>
      <ul class="space-y-0.5">
        <li v-for="doc in docs" :key="doc.slug">
          <NuxtLink
            :to="`/dokumentation/${doc.slug}`"
            class="flex items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm transition-colors"
            :class="
              isActive(doc.slug)
                ? 'bg-accent/10 font-medium text-accent-light'
                : 'text-ink-muted hover:bg-surface-2 hover:text-ink'
            "
          >
            <UIcon :name="doc.icon" class="size-4 shrink-0" />
            <span class="truncate">{{ doc.title }}</span>
          </NuxtLink>
        </li>
      </ul>
    </div>
  </nav>
</template>

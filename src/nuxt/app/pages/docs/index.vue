<script setup lang="ts">
import type { DocMeta } from "~/utils/docs";

definePageMeta({
  ssr: false,
});

const { list, groups } = useDocs();

useHead({
  title: "Dokumentation - Plan",
});

const hero = {
  title: "Dokumentation",
  description:
    "Hur Plan är byggt, hur målen och agentkörningarna fungerar, och hur du utvecklar och driftsätter systemet.",
};

function pageHref(doc: DocMeta): string {
  return `/docs/${doc.slug}`;
}
</script>

<template>
  <div class="mx-auto w-full max-w-5xl px-6 py-10 space-y-10">
    <header class="space-y-3">
      <UBadge
        icon="i-lucide-book-open"
        color="primary"
        variant="subtle"
        size="sm"
        class="rounded-full"
      >
        Dokumentation
      </UBadge>
      <h1 class="text-3xl font-bold tracking-tight text-ink sm:text-4xl">
        {{ hero.title }}
      </h1>
      <p class="max-w-2xl text-sm leading-relaxed text-ink-muted">
        {{ hero.description }}
      </p>
    </header>

    <section
      v-for="[group, docs] in groups"
      :key="group"
      class="space-y-4"
    >
      <h2 class="text-xs font-semibold uppercase tracking-widest text-ink-dim">
        {{ group }}
      </h2>

      <div class="grid gap-4 sm:grid-cols-2">
        <NuxtLink
          v-for="doc in docs"
          :key="doc.slug"
          :to="pageHref(doc)"
          class="group rounded-2xl border border-line bg-surface p-5 transition-colors hover:border-line-strong hover:bg-surface-2"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex size-9 shrink-0 items-center justify-center rounded-xl border border-line bg-surface-2 text-accent transition-colors group-hover:border-accent/40"
            >
              <UIcon :name="doc.icon" class="size-4.5" />
            </div>

            <div class="min-w-0 flex-1">
              <h3
                class="flex items-center gap-1.5 text-sm font-semibold text-ink"
              >
                {{ doc.title }}
                <UIcon
                  name="i-lucide-arrow-right"
                  class="size-3.5 -translate-x-1 text-ink-dim opacity-0 transition-all group-hover:translate-x-0 group-hover:opacity-100"
                />
              </h3>
              <p class="mt-1.5 text-sm leading-relaxed text-ink-muted">
                {{ doc.description }}
              </p>
            </div>
          </div>
        </NuxtLink>
      </div>
    </section>

    <section class="rounded-2xl border border-line bg-surface p-6">
      <h2 class="text-sm font-semibold text-ink">Snabbvägar</h2>
      <div class="mt-4 grid gap-4 sm:grid-cols-3">
        <div>
          <p class="text-xs font-medium uppercase tracking-wide text-ink-dim">
            Tangenter i appen
          </p>
          <p class="mt-1.5 text-sm text-ink-muted">
            Tryck <kbd class="rounded border border-line bg-surface-2 px-1.5 py-0.5 font-mono text-xs text-ink">?</kbd>
            i appen för att se hela listan.
          </p>
        </div>
        <div>
          <p class="text-xs font-medium uppercase tracking-wide text-ink-dim">
            Utveckling
          </p>
          <p class="mt-1.5 text-sm text-ink-muted">
            <NuxtLink to="/docs/utveckling" class="underline underline-offset-4">
              Kör appen lokalt
            </NuxtLink>
            med Tilt eller Bun.
          </p>
        </div>
        <div>
          <p class="text-xs font-medium uppercase tracking-wide text-ink-dim">
            Agenten
          </p>
          <p class="mt-1.5 text-sm text-ink-muted">
            <NuxtLink to="/docs/agent-systemet" class="underline underline-offset-4">
              Så arbetar agenten
            </NuxtLink>
            med dina mål.
          </p>
        </div>
      </div>
    </section>
  </div>
</template>

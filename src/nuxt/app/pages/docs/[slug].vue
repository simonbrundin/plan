<script setup lang="ts">
definePageMeta({
  ssr: false,
});

const route = useRoute();
const { list, groups, find } = useDocs();

const slug = computed(() => String(route.params.slug ?? ""));
const page = computed(() => find(slug.value));

if (!page.value) {
  throw createError({
    statusCode: 404,
    statusMessage: "Sidan finns inte",
    fatal: true,
  });
}

useHead(() => ({
  title: page.value ? `${page.value.title} - Plan` : "Dokumentation",
}));

// Ny sida ska börja högst upp - annars hamnar man kvar på samma
// scrollposition när man byter undersida.
watch(slug, () => {
  if (import.meta.client) window.scrollTo({ top: 0 });
});

// Föregående/nästa sida gör att man kan läsa dokumentationen i ordning.
const neighbours = computed(() => {
  const index = list.value.findIndex((doc) => doc.slug === slug.value);
  return {
    previous: index > 0 ? list.value[index - 1] : undefined,
    next: index >= 0 && index < list.value.length - 1 ? list.value[index + 1] : undefined,
  };
});

const showToc = computed(() => (page.value?.headings.length ?? 0) > 2);
</script>

<template>
  <div v-if="page" class="mx-auto w-full max-w-7xl px-6 py-10">
    <div class="mb-8 flex items-center gap-2 text-sm text-ink-dim">
      <NuxtLink to="/docs" class="transition-colors hover:text-ink">
        Dokumentation
      </NuxtLink>
      <UIcon name="i-lucide-chevron-right" class="size-3.5" />
      <span class="text-ink-muted">{{ page.title }}</span>
    </div>

    <div class="grid gap-10 lg:grid-cols-[200px_minmax(0,1fr)] xl:grid-cols-[200px_minmax(0,1fr)_200px]">
      <!-- Sidomeny -->
      <aside class="hidden lg:block">
        <div class="sticky top-6">
          <DocsNav :groups="groups" :current="page.slug" />
        </div>
      </aside>

      <!-- Innehåll -->
      <main class="min-w-0">
        <header class="mb-8 space-y-3">
          <div class="flex items-center gap-2.5">
            <div
              class="flex size-10 items-center justify-center rounded-xl border border-line bg-surface text-accent"
            >
              <UIcon :name="page.icon" class="size-5" />
            </div>
            <h1 class="text-2xl font-bold tracking-tight text-ink sm:text-3xl">
              {{ page.title }}
            </h1>
          </div>
          <p class="text-sm leading-relaxed text-ink-muted">
            {{ page.description }}
          </p>
          <p v-if="page.updated" class="text-xs text-ink-dim">
            Senast uppdaterad {{ page.updated }}
          </p>
        </header>

        <article class="docs-content" v-html="page.html" />

        <!-- Föregående / nästa -->
        <nav
          class="mt-12 grid gap-3 border-t border-line pt-6 sm:grid-cols-2"
          aria-label="Dokumentationsnavigation"
        >
          <NuxtLink
            v-if="neighbours.previous"
            :to="`/docs/${neighbours.previous.slug}`"
            class="group rounded-xl border border-line bg-surface p-4 transition-colors hover:border-line-strong hover:bg-surface-2"
          >
            <span class="flex items-center gap-1 text-xs text-ink-dim">
              <UIcon name="i-lucide-arrow-left" class="size-3" />
              Föregående
            </span>
            <span class="mt-1 block text-sm font-medium text-ink">
              {{ neighbours.previous.title }}
            </span>
          </NuxtLink>
          <div v-else />

          <NuxtLink
            v-if="neighbours.next"
            :to="`/docs/${neighbours.next.slug}`"
            class="group rounded-xl border border-line bg-surface p-4 text-right transition-colors hover:border-line-strong hover:bg-surface-2 sm:col-start-2"
          >
            <span class="flex items-center justify-end gap-1 text-xs text-ink-dim">
              Nästa
              <UIcon name="i-lucide-arrow-right" class="size-3" />
            </span>
            <span class="mt-1 block text-sm font-medium text-ink">
              {{ neighbours.next.title }}
            </span>
          </NuxtLink>
        </nav>
      </main>

      <!-- Innehållsförteckning -->
      <aside v-if="showToc" class="hidden xl:block">
        <div class="sticky top-6">
          <p class="mb-3 text-xs font-semibold uppercase tracking-widest text-ink-dim">
            På den här sidan
          </p>
          <ul class="space-y-1.5 border-l border-line">
            <li v-for="heading in page.headings" :key="heading.id">
              <a
                :href="`#${heading.id}`"
                class="block border-l border-transparent py-0.5 text-sm leading-snug text-ink-muted transition-colors hover:border-accent hover:text-ink"
                :class="heading.depth === 3 ? 'pl-6' : 'pl-3'"
              >
                {{ heading.text }}
              </a>
            </li>
          </ul>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  ssr: false,
});

const { settings, update } = useCalendarSettings();

const maxInstances = computed({
  get: () => settings.value.maxInstances,
  set: (value: number | undefined) => {
    if (typeof value !== "number" || !Number.isFinite(value)) return;

    update({ maxInstances: Math.min(100, Math.max(1, Math.round(value))) });
  },
});

const hideRecurring = computed({
  get: () => settings.value.hideRecurring,
  set: (value: boolean) => update({ hideRecurring: value }),
});
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-8 space-y-6">
    <header>
      <h1 class="text-2xl font-bold tracking-tight text-ink">Inställningar</h1>
      <p class="text-sm text-ink-muted mt-1">
        Inställningarna sparas i webbläsaren och gäller på den här enheten.
      </p>
    </header>

    <section class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-6 space-y-6">
      <div>
        <h2 class="text-lg font-semibold text-ink">Kalender</h2>
        <p class="text-sm text-ink-muted mt-1">
          Tidslinjen visar varje händelse som en rad. En återkommande serie
          visas som vanligt så länge den ryms. När den har fler förekomster i
          vyn än talet nedan, döljs hela serien.
        </p>
      </div>

      <UFormField
        label="Max antal instanser per serie"
        description="En återkommande serie visas med högst så här många förekomster i vyn. Har den fler, döljs hela serien."
        :ui="{ description: 'text-ink-muted' }"
      >
        <UInputNumber
          v-model="maxInstances"
          :min="1"
          :max="100"
          :step="1"
          size="lg"
          class="w-32"
        />
      </UFormField>

      <USwitch
        v-model="hideRecurring"
        label="Dölj återkommande serier"
        description="När avstängd visas alla serier, och tidslinjen kan bli full."
        :ui="{ label: 'text-ink', description: 'text-ink-muted' }"
      />

      <p class="text-xs text-ink-muted">
        En dold serie syns inte i tidslinjen. Zooma in på den, så syns den
        som vanligt igen.
      </p>
    </section>
  </div>
</template>

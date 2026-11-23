<script setup lang="ts">
// Bakåtkompatibel route för gamla bokmärken och länkar.
definePageMeta({ ssr: false })
navigateTo('/now', { redirectCode: 301 })
</script>

<template>
  <div class="flex items-center justify-center min-h-[50vh] text-ink-muted">
    Går till Gör nu…
  </div>
</template>

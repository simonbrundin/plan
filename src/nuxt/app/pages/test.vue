<script setup lang="ts">
console.log('[SSR] test.vue setup')
const start = Date.now()

// This page has no components, no API calls, no cookies
// Just simple text rendering to test SSR performance
</script>

<template>
  <div class="p-6">
    <h1>Test Page</h1>
    <p>Simple page for SSR testing</p>
    <p>Time: {{ new Date().toISOString() }}</p>
  </div>
</template>

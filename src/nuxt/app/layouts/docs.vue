<script setup lang="ts">
// Dokumentationslayouten: samma skal som standardsidan, men utan
// kalendern till höger - dokumentationens texter får hela sidbredden.
</script>

<template>
  <AppShell :calendar="false">
    <slot />
  </AppShell>
</template>

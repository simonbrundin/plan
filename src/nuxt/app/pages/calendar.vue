<script setup lang="ts">
definePageMeta({
  ssr: false,
})
</script>

<template>
  <!-- På dator visas kalendern redan som layoutens permanenta högerspalt.
       På mobil behöver sidan fortfarande visa kalendern som huvudinnehåll. -->
  <div class="lg:hidden">
    <VerticalCalendar />
  </div>
</template>

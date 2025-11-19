<script setup>
const { loggedIn } = useUserSession();
</script>

<template>
  <div class="max-w-2xl mx-auto p-6">
    <div class="flex items-center justify-center min-h-[50vh]">
      <div v-if="loggedIn" class="text-center">
        <h1 class="text-3xl font-bold text-gray-900 mb-6">Välkommen tillbaka!</h1>
        <NuxtLink to="/goal/1">
          <UButton size="lg" icon="material-symbols:arrow-forward-rounded">
            Gå till mål
          </UButton>
        </NuxtLink>
      </div>
      <div v-else class="text-center">
        <h1 class="text-3xl font-bold text-gray-900 mb-6">Välkommen till Plan</h1>
        <NuxtLink to="/login">
          <UButton size="lg" icon="material-symbols:login-rounded">
            Logga in
          </UButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

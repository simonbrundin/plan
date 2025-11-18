<script setup>
const { loggedIn, user, session, fetch, clear, openInPopup } = useUserSession();
</script>

<template>
  <div v-if="loggedIn">
    <h1 class="text-3xl font-bold text-gray-300 mb-2" v-if="user">Inloggad</h1>
    <h1 class="text-3xl font-bold text-gray-300 mb-2" v-else>Utloggad</h1>

    <p v-if="user" class="mb-2">
      User ID: <span class="font-mono text-sm">{{ user.id }}</span>
    </p>

    <p v-if="user" class="mb-2">
      Sub: <span class="font-mono text-sm">{{ user.sub }}</span>
    </p>

    <p v-if="session?.loggedInAt">
      Inloggad sedan
      {{
        new Date(session.loggedInAt).toLocaleString("en-US", {
          month: "short",
          day: "2-digit",
          year: "numeric",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
          hour12: false,
        })
      }}
    </p>
    <UButton @click="clear" icon="solar:logout-2-line-duotone">Logout</UButton>
  </div>
  <div v-else>
    <h1>Not logged in</h1>
    <a href="/auth/authentik">
      <UButton>Logga in</UButton>
    </a>
  </div>
  <NuxtLink to="/">
    <UButton color="neutral" icon="material-symbols-light:home-outline-rounded"
      >Hem</UButton
    >
  </NuxtLink>
</template>
<style scoped></style>

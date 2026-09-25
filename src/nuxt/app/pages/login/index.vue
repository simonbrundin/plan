<script setup lang="ts">
const { loggedIn, user, session, clear } = useUserSession()

const handleLogout = async () => {
  await clear()
  window.location.reload()
}
</script>

<template>
  <div v-if="loggedIn">
    <h1 class="text-3xl font-bold text-gray-300 mb-2" v-if="user">Inloggad</h1>
    <h1 class="text-3xl font-bold text-gray-300 mb-2" v-else>Utloggad</h1>

    <p v-if="user" class="mb-2">
      User ID: <span class="font-mono text-sm">{{ user.id }}</span>
    </p>

    <p v-if="user?.email" class="mb-2">
      Email: <span class="font-mono text-sm">{{ user.email }}</span>
    </p>

    <p v-if="user?.name" class="mb-2">
      Namn: <span class="font-mono text-sm">{{ user.name }}</span>
    </p>

    <p v-if="session?.loggedInAt">
      Inloggad sedan
      {{
        new Date(session.loggedInAt).toLocaleString("sv-SE", {
          year: "numeric",
          month: "short",
          day: "2-digit",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        })
      }}
    </p>
    <UButton @click="handleLogout" icon="solar:logout-2-line-duotone">Logout</UButton>
  </div>
  <div v-else>
    <h1>Not logged in</h1>
    <a href="/auth/zitadel">
      <UButton>Logga in med Zitadel</UButton>
    </a>
  </div>
  <NuxtLink to="/">
    <UButton color="neutral" icon="material-symbols-light:home-outline-rounded"
      >Hem</UButton
    >
  </NuxtLink>
</template>
<style scoped></style>

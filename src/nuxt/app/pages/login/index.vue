<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface SessionData {
  userId: string;
  email?: string;
  name?: string;
  sessionToken: string;
  loggedInAt: number;
}

const session = ref<SessionData | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    session.value = await $fetch<SessionData>('/api/auth/session')
  } catch {
    session.value = null
  } finally {
    loading.value = false
  }
})

const handleLogout = async () => {
  await $fetch('/api/auth/logout', { method: 'POST' })
  window.location.reload()
}
</script>

<template>
  <div>
    <div v-if="loading" class="text-gray-400">
      Laddar...
    </div>
    <div v-else-if="session">
      <h1 class="text-3xl font-bold text-gray-300 mb-2">Inloggad</h1>

      <p v-if="session.userId" class="mb-2">
        User ID: <span class="font-mono text-sm">{{ session.userId }}</span>
      </p>

      <p v-if="session.name" class="mb-2">
        Namn: <span class="font-mono text-sm">{{ session.name }}</span>
      </p>

      <p v-if="session.email" class="mb-2">
        Email: <span class="font-mono text-sm">{{ session.email }}</span>
      </p>

      <p v-if="session.loggedInAt">
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
      <a href="/api/auth/zitadel">
        <UButton>Logga in med Zitadel</UButton>
      </a>
    </div>
    <NuxtLink to="/">
      <UButton color="neutral" icon="material-symbols-light:home-outline-rounded"
        >Hem</UButton
      >
    </NuxtLink>
  </div>
</template>
<style scoped></style>

  <script setup lang="ts">
  import type { NavigationMenuItem } from "@nuxt/ui";

  const route = useRoute();
  const { loggedIn } = useUserSession();

  const open = ref(false);

  const links = [
    [
      {
        label: "Inbox",
        icon: "i-lucide-inbox",
        to: "/",
        onSelect: () => {
          open.value = false;
        },
      },
      {
        label: "Goals",
        icon: "i-lucide-target",
        to: "/goal/1",
        onSelect: () => {
          open.value = false;
        },
      },
      {
        label: "Login",
        icon: "i-lucide-log-in",
        to: "/login",
        onSelect: () => {
          open.value = false;
        },
      },
    ],
  ] satisfies NavigationMenuItem[][];

  // Load goals data on client-side only to avoid SSR issues
  const goalsData = ref(null);

  // Load goals data after component is mounted (client-side only)
  onMounted(async () => {
    try {
      const query = `
      query GetAllGoals {
        goals(order_by: { created: desc }) {
          id
          title
          created
          finished
        }
      }
    `;

      const response = await $fetch('http://localhost:8080/v1/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-hasura-admin-secret': config.public.hasuraAdminSecret
        },
        body: JSON.stringify({ query })
      });

      if (response.data) {
        goalsData.value = response.data;
      }
    } catch (error) {
      console.warn("Failed to load goals for navigation:", error);
    }
  });

  const groups = computed(() => {
    const goalItems =
      goalsData.value?.goals?.map((goal) => ({
        id: `goal-${goal.id}`,
        label: goal.title,
        icon: goal.finished ? "i-lucide-check-circle" : "i-lucide-circle",
        to: `/goal/${goal.id}`,
        suffix: goal.finished ? "Klar" : undefined,
      })) || [];

    return [
      {
        id: "goals",
        label: "Mål",
        items: goalItems,
      },
      {
        id: "links",
        label: "Navigering",
        items: links.flat(),
      },
    ];
  });
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar id="default" v-model:open="open" collapsible resizable class="bg-elevated/25"
      :ui="{ footer: 'lg:border-t lg:border-default' }">
      <template #default="{ collapsed }">
        <UDashboardSearchButton :collapsed="collapsed" class="bg-transparent ring-default" />

        <UNavigationMenu :collapsed="collapsed" :items="links[0]" orientation="vertical" tooltip popover />
      </template>

      <template #footer="{ collapsed }">
        <UserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <UDashboardSearch :groups="groups" />

    <UDashboardPanel class="!overflow-y-auto !h-screen">
      <UDashboardNavbar class="hidden lg:flex">
        <template #toggle>
          <UDashboardSidebarToggle />
        </template>

        <template #right>
          <UDashboardSearchButton collapsed />
        </template>
      </UDashboardNavbar>

      <div class="pb-20 lg:pb-0">
        <slot />
      </div>

      <!-- Mobile bottom navigation -->
      <div
        class="fixed bottom-0 left-0 right-0 z-50 flex items-center justify-around gap-2 border-t border-default bg-elevated/95 backdrop-blur-sm p-4 lg:hidden">
        <UDashboardSidebarToggle />
        <UDashboardSearchButton collapsed />
      </div>
    </UDashboardPanel>
  </UDashboardGroup>
</template>

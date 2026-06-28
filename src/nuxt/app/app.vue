<script setup lang="ts">
const colorMode = useColorMode();

const color = computed(() =>
  colorMode.value === "dark" ? "#18181b" : "white",
);

useSeoMeta({
  titleTemplate: "%s - Plan",
  ogImage: "https://plan.example.com/og-image.png",
  twitterCard: "summary_large_image",
});

useHead({
  htmlAttrs: {
    lang: "sv",
  },
  link: [
    {
      rel: "icon",
      href: "/favicon.ico",
    },
  ],
  meta: [
    {
      name: "theme-color",
      content: color,
    },
  ],
});
</script>

<template>
  <UApp>
    <NuxtLoadingIndicator />

    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>

// Plugin to configure Pinia to skip hydration for all stores
export default defineNuxtPlugin({
  name: 'pinia-hydration-config',
  enforce: 'pre',
  setup(nuxtApp) {
    const pinia = nuxtApp.$pinia

    if (!pinia) return

    // Add a plugin to Pinia that marks all stores to skip hydration
    pinia.use(({ store }) => {
      // Mark all stores to skip hydration to prevent serialization issues
      if (store.$id && !store.$state._skipHydrate) {
        // @ts-expect-error - Adding internal property
        store._skipHydrate = true
      }
    })
  },
})

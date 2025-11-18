// Server-side plugin to fix Pinia payload serialization issues
export default defineNuxtPlugin({
  name: 'pinia-server-fix',
  enforce: 'pre',
  setup(nuxtApp) {
    if (!process.server) return

    // Override Object.prototype.hasOwnProperty for the global context
    // This prevents errors when Pinia tries to check properties on objects
    // that were created with Object.create(null)
    const originalHasOwnProperty = Object.prototype.hasOwnProperty

    // Patch the hasOwnProperty method to handle null-prototype objects
    Object.defineProperty(Object.prototype, 'hasOwnProperty', {
      value: function (prop: string) {
        try {
          return originalHasOwnProperty.call(this, prop)
        } catch (e) {
          // If this fails, use Object.hasOwn or fallback
          return Object.hasOwn ? Object.hasOwn(this, prop) : prop in this
        }
      },
      writable: true,
      enumerable: false,
      configurable: true,
    })

    // Also handle Pinia's payload plugin directly
    nuxtApp.hook('app:rendered', () => {
      // After rendering, clean up any problematic state
      if (nuxtApp.payload?.state) {
        try {
          // Ensure all state objects have proper prototypes
          nuxtApp.payload.state = JSON.parse(JSON.stringify(nuxtApp.payload.state))
        } catch (e) {
          console.warn('Failed to clean payload state:', e)
        }
      }
    })
  },
})

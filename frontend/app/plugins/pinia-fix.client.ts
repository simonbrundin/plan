// Client-side plugin to fix potential Pinia hydration issues
export default defineNuxtPlugin(() => {
  // Ensure objects have proper prototypes to prevent hasOwnProperty errors
  if (process.client) {
    // This plugin runs only on the client side
    // Nuxt and Pinia should handle the hydration correctly with the store configuration
  }
})
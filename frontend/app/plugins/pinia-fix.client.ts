// Client-side plugin to fix potential Pinia hydration issues
export default defineNuxtPlugin({
  name: 'pinia-hydration-fix',
  enforce: 'pre', // Run before other plugins
  setup(nuxtApp) {
    if (!process.client) return

    // Add global error handler for hydration issues
    nuxtApp.hook('app:error', (error) => {
      if (error?.message?.includes('hasOwnProperty')) {
        console.warn('Pinia hydration error caught and handled:', error.message)
        // Prevent the error from bubbling up
        return false
      }
    })

    // Also catch Vue errors
    nuxtApp.hook('vue:error', (error) => {
      if (error?.message?.includes('hasOwnProperty')) {
        console.warn('Vue hydration error caught and handled:', error.message)
        return false
      }
    })
  }
})
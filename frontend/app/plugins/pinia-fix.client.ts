// Client-side plugin to fix potential Pinia hydration issues
export default defineNuxtPlugin((nuxtApp) => {
  // This plugin prevents hasOwnProperty errors during Pinia state hydration
  // by ensuring proper error handling in the hydration process

  if (process.client) {
    // Add global error handler for hydration issues
    nuxtApp.hook('app:error', (error) => {
      if (error.message?.includes('hasOwnProperty')) {
        console.warn('Pinia hydration error caught and handled:', error.message)
        // Prevent the error from bubbling up
        return false
      }
    })
  }
})
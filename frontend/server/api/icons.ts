export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const searchQuery = query.q as string

  // If no search query, return info about available collections
  if (!searchQuery || searchQuery.trim().length < 2) {
    return {
      results: [],
      message: 'Enter at least 2 characters to search',
    }
  }

  try {
    // Search using Iconify API
    const response = await fetch(
      `https://api.iconify.design/search?query=${encodeURIComponent(searchQuery)}`
    )

    if (!response.ok) {
      throw new Error(`Iconify API error: ${response.statusText}`)
    }

    const data = await response.json()

    // Transform Iconify API response
    // API returns array of strings like "mdi:test-tube"
    const results = (data.icons || []).map((iconString: string) => {
      const parts = iconString.split(':')
      return {
        collection: parts[0] || 'unknown',
        name: parts.slice(1).join(':') || iconString,
        full: iconString,
      }
    })

    return {
      results,
      total: results.length,
    }
  } catch (error) {
    console.error('Failed to search icons:', error)
    return {
      results: [],
      error: 'Failed to search icons',
    }
  }
})

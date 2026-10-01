export default defineEventHandler((event) => {
  const start = Date.now()
  const path = event.path
  
  // Log timing when response finishes
  event.node.res.on('finish', () => {
    const duration = Date.now() - start
    console.log(`[TIMING] ${path} - ${duration}ms`)
  })
  
  // Log timing when response finishes (alternative)
  event.node.res.on('close', () => {
    const duration = Date.now() - start
    if (duration > 100) { // Only log slow requests
      console.log(`[TIMING-SLOW] ${path} - ${duration}ms`)
    }
  })
})

// Nitro plugin to fix Pinia hasOwnProperty errors
// This runs before any Nuxt/Pinia code, ensuring the fix is in place early
export default defineNitroPlugin((nitroApp) => {
  // Globally patch hasOwnProperty to handle null-prototype objects
  // This is necessary because Pinia's shouldHydrate function calls
  // obj.hasOwnProperty() on objects that may not have this method

  const safeHasOwnProperty = function (this: any, prop: PropertyKey): boolean {
    // First, check if 'this' is null or undefined
    if (this == null) {
      return false
    }

    try {
      // Try the standard approach first
      return Object.prototype.hasOwnProperty.call(this, prop)
    } catch (e) {
      // If that fails, use Object.hasOwn or fallback
      if (typeof Object.hasOwn === 'function') {
        return Object.hasOwn(this, prop)
      }
      // Last resort fallback
      return prop in Object(this)
    }
  }

  // Override hasOwnProperty on Object.prototype
  Object.defineProperty(Object.prototype, 'hasOwnProperty', {
    value: safeHasOwnProperty,
    writable: true,
    enumerable: false,
    configurable: true,
  })

  // Also patch objects created with Object.create(null)
  const originalCreate = Object.create
  Object.create = function (proto: any, propertiesObject?: PropertyDescriptorMap) {
    const obj = originalCreate.call(Object, proto, propertiesObject)

    // If creating a null-prototype object, add hasOwnProperty to it
    if (proto === null && obj) {
      Object.defineProperty(obj, 'hasOwnProperty', {
        value: safeHasOwnProperty,
        writable: true,
        enumerable: false,
        configurable: true,
      })
    }

    return obj
  }

  console.log('[Nitro] Pinia hasOwnProperty fix installed')
})

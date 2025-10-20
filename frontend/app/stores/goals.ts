  import { defineStore } from 'pinia'

interface Goal {
  id: number
  title: string
  created: string
  finished: string | null
}

interface GoalRelation {
  child_id: number
  parent_id: number
}

export const useGoalsStore = defineStore('goals', {
  state: () => ({
    goals: [] as Goal[],
    relations: [] as GoalRelation[],
    isLoaded: false,
    error: null as string | null,
  }),

  // Add hydration strategy to prevent serialization issues
  hydrate: (state, nuxtApp) => {
    // Ensure all objects are plain objects to prevent prototype issues during hydration
    if (Array.isArray(state.goals)) {
      state.goals = state.goals.map(goal => {
        // Create a new plain object to avoid prototype issues
        return { ...goal } as Goal;
      });
    }

    if (Array.isArray(state.relations)) {
      state.relations = state.relations.map(relation => {
        // Create a new plain object to avoid prototype issues
        return { ...relation } as GoalRelation;
      });
    }
  },

  getters: {
    // Hitta mål baserat på ID
    getGoalById: (state) => (id: number) => {
      return state.goals.find(g => g.id === id)
    },

    // Sök mål baserat på text
    searchGoals: (state) => (query: string) => {
      if (!query.trim()) return []
      const lowerQuery = query.toLowerCase()
      return state.goals.filter(g =>
        g.title.toLowerCase().includes(lowerQuery)
      )
    },

    // Hämta föräldrar för ett mål
    getParents: (state) => (goalId: number) => {
      const parentIds = state.relations
        .filter(r => r.child_id === goalId)
        .map(r => r.parent_id)

      return state.goals.filter(g => parentIds.includes(g.id))
    },

    // Hämta barn för ett mål
    getChildren: (state) => (goalId: number) => {
      const childIds = state.relations
        .filter(r => r.parent_id === goalId)
        .map(r => r.child_id)

      return state.goals.filter(g => childIds.includes(g.id))
    },

    // Hämta grundmål (mål utan föräldrar)
    getRootGoals: (state) => {
      const childIds = new Set(state.relations.map(r => r.child_id))
      return state.goals.filter(g => !childIds.has(g.id))
    },
  },

  actions: {
    // Ladda alla mål och relationer
    async loadGoals() {
      // This method is now handled in components
      // Keeping for compatibility
      this.isLoaded = true
    },

    // Lägg till ett nytt mål
    addGoal(goal: Goal) {
      this.goals.push(goal)
    },

    // Uppdatera ett mål
    updateGoal(id: number, updates: Partial<Goal>) {
      const index = this.goals.findIndex(g => g.id === id)
      if (index !== -1) {
        this.goals[index] = { ...this.goals[index], ...updates }
      }
    },

    // Ta bort ett mål
    removeGoal(id: number) {
      this.goals = this.goals.filter(g => g.id !== id)
      // Ta även bort relaterade relationer
      this.relations = this.relations.filter(
        r => r.child_id !== id && r.parent_id !== id
      )
    },

    // Lägg till en relation
    addRelation(childId: number, parentId: number) {
      // Kolla om relationen redan finns
      const exists = this.relations.some(
        r => r.child_id === childId && r.parent_id === parentId
      )

      if (!exists) {
        this.relations.push({ child_id: childId, parent_id: parentId })
      }
    },

    // Ta bort en relation
    removeRelation(childId: number, parentId: number) {
      this.relations = this.relations.filter(
        r => !(r.child_id === childId && r.parent_id === parentId)
      )
    },

    // Clear error state
    clearError() {
      this.error = null
    },

    // Reset store state (useful for error recovery)
    reset() {
      this.goals = []
      this.relations = []
      this.isLoaded = false
      this.error = null
    },
  },
})

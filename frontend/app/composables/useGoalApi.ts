import type { Goal, GoalWithRelations, GetGoalResponse } from '~/types/goal'

export function useGoalApi() {
  const config = useRuntimeConfig()

  const fetchGoalData = async (goalId: number): Promise<GetGoalResponse> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          query GetGoal($id: Int!) {
            goal: goals_by_pk(id: $id) {
              id
              title
              icon
              created
              finished

              # Barn - relationer där detta mål är parent
              childRelations: goalRelationsByParentId(order_by: { order: asc }) {
                child_id
                order
                weight
              }

              # Förälder - relationer där detta mål är child
              parentRelations: goal_relations {
                goalByParentId {
                  id
                  title
                  icon
                }
              }
            }

            # Hämta alla mål för att kunna matcha children
            allGoals: goals {
              id
              title
              icon
              finished
              created
            }
          }
        `,
        variables: { id: goalId },
      }),
    })

    const result = await response.json()

    if (result.errors) {
      throw new Error(result.errors[0].message)
    }

    return result.data
  }

  const updateGoalTitle = async (goalId: number, title: string): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalTitle($id: Int!, $title: String!) {
            update_goals_by_pk(pk_columns: { id: $id }, _set: { title: $title }) {
              id
              title
            }
          }
        `,
        variables: { id: goalId, title },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const updateGoalIcon = async (goalId: number, icon: string): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalIcon($id: Int!, $icon: String!) {
            update_goals_by_pk(pk_columns: { id: $id }, _set: { icon: $icon }) {
              id
              icon
            }
          }
        `,
        variables: { id: goalId, icon },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const toggleGoalFinished = async (goalId: number, finished: string | null): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalFinished($id: Int!, $finished: timestamptz) {
            update_goals_by_pk(pk_columns: { id: $id }, _set: { finished: $finished }) {
              id
              title
              created
              finished
            }
          }
        `,
        variables: { id: goalId, finished },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const deleteGoal = async (goalId: number): Promise<void> => {
    const deleteGoalQuery = `
      mutation DeleteGoal($id: Int!) {
        delete_goal_relations(where: { _or: [{ child_id: { _eq: $id } }, { parent_id: { _eq: $id } }] }) {
          affected_rows
        }
        delete_user_goals(where: { goal_id: { _eq: $id } }) {
          affected_rows
        }
        delete_goals_by_pk(id: $id) {
          id
        }
      }
    `

    const response = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: deleteGoalQuery,
        variables: { id: goalId },
      }),
    })

    if ((response as any).errors) {
      throw new Error((response as any).errors[0].message)
    }
  }

  const addParentRelation = async (childId: number, parentId: number): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation AddParentRelation($childId: Int!, $parentId: Int!) {
            insert_goal_relations_one(
              object: { child_id: $childId, parent_id: $parentId }
            ) {
              child_id
              parent_id
            }
          }
        `,
        variables: { childId, parentId },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const removeParentRelation = async (childId: number, parentId: number): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation RemoveParentRelation($childId: Int!, $parentId: Int!) {
            delete_goal_relations(
              where: {
                child_id: {_eq: $childId}
                parent_id: {_eq: $parentId}
              }
            ) {
              affected_rows
            }
          }
        `,
        variables: { childId, parentId },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const addChildRelation = async (childId: number, parentId: number, order: number): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation AddParentRelation($childId: Int!, $parentId: Int!, $order: Int!) {
            insert_goal_relations_one(
              object: { child_id: $childId, parent_id: $parentId, order: $order }
            ) {
              child_id
              parent_id
              order
            }
          }
        `,
        variables: { childId, parentId, order },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const updateGoalOrder = async (parentId: number, childId: number, order: number): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalOrder($parent_id: Int!, $child_id: Int!, $order: Int!) {
            update_goal_relations_by_pk(
              pk_columns: { parent_id: $parent_id, child_id: $child_id }
              _set: { order: $order }
            ) {
              parent_id
              child_id
              order
            }
          }
        `,
        variables: { parent_id: parentId, child_id: childId, order },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const updateGoalWeight = async (parentId: number, childId: number, weight: number): Promise<void> => {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalWeight($parent_id: Int!, $child_id: Int!, $weight: Int!) {
            update_goal_relations_by_pk(
              pk_columns: { parent_id: $parent_id, child_id: $child_id }
              _set: { weight: $weight }
            ) {
              parent_id
              child_id
              weight
            }
          }
        `,
        variables: { parent_id: parentId, child_id: childId, weight },
      }),
    })

    const result = await response.json()
    if (result.errors) {
      throw new Error(result.errors[0].message)
    }
  }

  const createGoal = async (title: string, userId: number): Promise<Goal> => {
    const createGoalQuery = `
      mutation CreateGoal($title: String!) {
        insert_goals_one(object: { title: $title, icon: "roentgen:default" }) {
          id
          title
          icon
          created
          finished
        }
      }
    `

    const goalResponse = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: createGoalQuery,
        variables: { title },
      }),
    })

    if ((goalResponse as any).errors) {
      throw new Error((goalResponse as any).errors[0].message)
    }

    const newGoal = (goalResponse as any).data?.insert_goals_one

    // Create user_goals relation
    const createUserGoalQuery = `
      mutation CreateUserGoal($userId: Int!, $goalId: Int!) {
        insert_user_goals_one(object: { user_id: $userId, goal_id: $goalId }) {
          user_id
          goal_id
        }
      }
    `

    const userGoalResponse = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: createUserGoalQuery,
        variables: { userId, goalId: newGoal.id },
      }),
    })

    if ((userGoalResponse as any).errors) {
      throw new Error((userGoalResponse as any).errors[0].message)
    }

    return newGoal
  }

  return {
    fetchGoalData,
    updateGoalTitle,
    updateGoalIcon,
    toggleGoalFinished,
    deleteGoal,
    addParentRelation,
    removeParentRelation,
    addChildRelation,
    updateGoalOrder,
    updateGoalWeight,
    createGoal,
  }
}

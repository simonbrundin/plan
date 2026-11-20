import type { GoalWithWeight } from '~/types/goal'

export type GoalWeight = { childId: number; weight: number }

/** Gör om äldre relativa vikter till heltalsprocent som summerar till 100. */
export function normalizeGoalWeights(children: GoalWithWeight[]): GoalWeight[] {
  if (children.length === 0) return []

  const total = children.reduce((sum, child) => sum + Math.max(0, child.weight), 0)
  const raw = children.map((child) =>
    total > 0 ? (Math.max(0, child.weight) / total) * 100 : 100 / children.length,
  )
  const weights = raw.map(Math.floor)
  let remainder = 100 - weights.reduce((sum, weight) => sum + weight, 0)

  raw
    .map((value, index) => ({ index, fraction: value - weights[index] }))
    .sort((a, b) => b.fraction - a.fraction)
    .forEach(({ index }) => {
      if (remainder <= 0) return
      weights[index] += 1
      remainder -= 1
    })

  return children.map((child, index) => ({ childId: child.id, weight: weights[index] }))
}

/** Ändrar ett mål och fördelar resten proportionellt mellan de övriga. */
export function redistributeGoalWeights(
  current: GoalWeight[],
  childId: number,
  targetWeight: number,
): GoalWeight[] {
  const target = Math.max(0, Math.min(100, Math.round(targetWeight)))
  const currentTarget = current.find((item) => item.childId === childId)
  const others = current.filter((item) => item.childId !== childId)
  if (!currentTarget || others.length === 0) return current

  const othersTotal = others.reduce((sum, item) => sum + Math.max(0, item.weight), 0)
  const remaining = 100 - target
  const raw = others.map((item) =>
    othersTotal > 0 ? (Math.max(0, item.weight) / othersTotal) * remaining : remaining / others.length,
  )
  const weights = raw.map(Math.floor)
  let remainder = remaining - weights.reduce((sum, weight) => sum + weight, 0)

  raw
    .map((value, index) => ({ index, fraction: value - weights[index] }))
    .sort((a, b) => b.fraction - a.fraction)
    .forEach(({ index }) => {
      if (remainder <= 0) return
      weights[index] += 1
      remainder -= 1
    })

  return [
    { childId, weight: target },
    ...others.map((item, index) => ({ childId: item.childId, weight: weights[index] })),
  ]
}

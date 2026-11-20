import type { GoalWithWeight } from '~/types/goal'

export type GoalWeight = { childId: number; weight: number }

function allocate(total: number, proportions: number[], minimum: number) {
  const baseTotal = minimum * proportions.length
  const distributable = Math.max(0, total - baseTotal)
  const proportionTotal = proportions.reduce((sum, value) => sum + Math.max(0, value), 0)
  const raw = proportions.map((value) =>
    proportionTotal > 0
      ? (Math.max(0, value) / proportionTotal) * distributable
      : distributable / proportions.length,
  )
  const weights = raw.map((value) => minimum + Math.floor(value))
  let remainder = total - weights.reduce((sum, weight) => sum + weight, 0)

  raw
    .map((value, index) => ({ index, fraction: value - Math.floor(value) }))
    .sort((a, b) => b.fraction - a.fraction)
    .forEach(({ index }) => {
      if (remainder <= 0) return
      weights[index] += 1
      remainder -= 1
    })

  return weights
}

/** Gör om äldre relativa vikter till heltalsprocent som summerar till 100. */
export function normalizeGoalWeights(children: GoalWithWeight[]): GoalWeight[] {
  if (children.length === 0) return []

  const total = children.reduce((sum, child) => sum + Math.max(0, child.weight), 0)
  const proportions = children.map((child) => (total > 0 ? Math.max(0, child.weight) : 1))
  const minimum = children.length <= 100 ? 1 : 0
  const weights = allocate(100, proportions, minimum)

  return children.map((child, index) => ({ childId: child.id, weight: weights[index] }))
}

/** Ändrar ett mål och fördelar resten proportionellt mellan de övriga. */
export function redistributeGoalWeights(
  current: GoalWeight[],
  childId: number,
  targetWeight: number,
): GoalWeight[] {
  const currentTarget = current.find((item) => item.childId === childId)
  const others = current.filter((item) => item.childId !== childId)
  if (!currentTarget || others.length === 0) return current

  const minimum = current.length <= 100 ? 1 : 0
  const maxTarget = 100 - minimum * others.length
  const target = Math.max(minimum, Math.min(maxTarget, Math.round(targetWeight)))
  const remaining = 100 - target
  const weights = allocate(
    remaining,
    others.map((item) => Math.max(0, item.weight)),
    minimum,
  )

  return [
    { childId, weight: target },
    ...others.map((item, index) => ({ childId: item.childId, weight: weights[index] })),
  ]
}

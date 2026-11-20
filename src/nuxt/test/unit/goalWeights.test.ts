import { describe, expect, it } from 'vitest'
import { redistributeGoalWeights } from '../../app/utils/goalWeights'

describe('redistributeGoalWeights', () => {
  it('fördelar om resten proportionellt när ett mål flyttas', () => {
    expect(redistributeGoalWeights([
      { childId: 1, weight: 20 },
      { childId: 2, weight: 20 },
      { childId: 3, weight: 20 },
      { childId: 4, weight: 20 },
      { childId: 5, weight: 20 },
    ], 1, 60)).toEqual([
      { childId: 1, weight: 60 },
      { childId: 2, weight: 10 },
      { childId: 3, weight: 10 },
      { childId: 4, weight: 10 },
      { childId: 5, weight: 10 },
    ])
  })

  it('summerar alltid till 100 även vid avrundning', () => {
    const result = redistributeGoalWeights([
      { childId: 1, weight: 60 },
      { childId: 2, weight: 10 },
      { childId: 3, weight: 10 },
      { childId: 4, weight: 10 },
      { childId: 5, weight: 10 },
    ], 1, 1)

    expect(result.reduce((sum, item) => sum + item.weight, 0)).toBe(100)
  })
})

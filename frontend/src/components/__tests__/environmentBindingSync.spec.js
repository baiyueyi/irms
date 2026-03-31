import { describe, expect, it } from 'vitest'
import { diffEnvironmentIds } from '../environmentBindingSync'

describe('diffEnvironmentIds', () => {
  it('returns added and removed ids', () => {
    const result = diffEnvironmentIds([1, 2], [2, 3])
    expect(result).toEqual({ added: [3], removed: [1] })
  })

  it('deduplicates and normalizes values', () => {
    const result = diffEnvironmentIds(['1', 1, 2], [2, '3', 3])
    expect(result).toEqual({ added: [3], removed: [1] })
  })
})

import { describe, expect, it } from 'vitest'
import { apiQueryKeys } from '@nomnomvault/api-client'

describe('api client root import', () => {
  it('builds recipe query keys from the package root', () => {
    expect(apiQueryKeys.recipes.detail('recipe_sample')).toEqual([
      'recipes',
      'detail',
      'recipe_sample',
    ])
  })
})

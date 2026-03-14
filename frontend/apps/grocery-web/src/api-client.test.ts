import { describe, expect, it } from 'vitest'
import { apiQueryKeys } from '@nomnomvault/api-client'

describe('api client root import', () => {
  it('builds auth query keys from the package root', () => {
    expect(apiQueryKeys.auth.session()).toEqual(['auth', 'session'])
  })
})

import type { components } from '../generated/schema'

export type ApiValidationError = components['schemas']['apicontract.ValidationError']
export type ApiErrorPayload = components['schemas']['apicontract.ErrorResponse']

export class ApiError extends Error {
  status: number
  code: string
  details?: ApiValidationError[]

  constructor(payload: ApiErrorPayload) {
    super(payload.message)
    this.name = 'ApiError'
    this.status = payload.status
    this.code = payload.code
    this.details = payload.details
  }
}

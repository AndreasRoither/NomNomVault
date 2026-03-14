import type { components } from '../generated/schema'

type ApiValidationError = components['schemas']['apicontract.ValidationError']
export type ApiErrorPayload = components['schemas']['apicontract.ErrorResponse']

export class ApiError extends Error {
  status: number
  code: string
  details?: ApiValidationError[]

  constructor(payload: ApiErrorPayload) {
    super(payload.message ?? 'Request failed.')
    this.name = 'ApiError'
    this.status = payload.status ?? 500
    this.code = payload.code ?? 'unknown_error'
    this.details = payload.details
  }
}

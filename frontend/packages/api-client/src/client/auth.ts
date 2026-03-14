import type { components } from '../generated/schema'
import type { RequestContext } from './http'

export type AuthRegisterRequest = components['schemas']['httpapi.AuthRegisterRequest']
export type AuthLoginRequest = components['schemas']['httpapi.AuthLoginRequest']
export type AuthSessionResponse = components['schemas']['httpapi.AuthSessionResponse']

export function createAuthClient(context: RequestContext) {
  return {
    register: (body: AuthRegisterRequest) =>
      context.request<AuthSessionResponse, AuthRegisterRequest>({
        path: '/auth/register',
        method: 'POST',
        body,
      }),
    login: (body: AuthLoginRequest) =>
      context.request<AuthSessionResponse, AuthLoginRequest>({
        path: '/auth/login',
        method: 'POST',
        body,
      }),
    logout: () =>
      context.request<void>({
        path: '/auth/logout',
        method: 'POST',
      }),
    refresh: () =>
      context.request<AuthSessionResponse>({
        path: '/auth/refresh',
        method: 'POST',
      }),
    session: () =>
      context.request<AuthSessionResponse>({
        path: '/auth/session',
      }),
  }
}

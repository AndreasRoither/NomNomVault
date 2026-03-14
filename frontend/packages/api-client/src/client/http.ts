import { ApiError, type ApiErrorPayload } from './errors'
import { createAuthClient } from './auth'
import { createRecipesClient } from './recipes'

export type ApiClientOptions = {
  baseUrl?: string
  fetch?: typeof fetch
  getCsrfToken?: () => string | undefined
  headers?: HeadersInit
}

export type RequestOptions<TBody> = {
  path: string
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: TBody
  search?: Record<string, string | number | boolean | undefined>
  headers?: HeadersInit
}

export type RequestContext = {
  request: <TResponse, TBody = undefined>(options: RequestOptions<TBody>) => Promise<TResponse>
}

export type ApiClient = ReturnType<typeof createApiClient>

function buildUrl(baseUrl: string, path: string, search: RequestOptions<unknown>['search']) {
  const normalizedBase = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const url = new URL(`${normalizedBase}${normalizedPath}`, 'http://nomnomvault.local')

  if (!baseUrl.startsWith('http://') && !baseUrl.startsWith('https://') && !baseUrl.startsWith('/')) {
    url.pathname = `${normalizedBase}${normalizedPath}`
  }

  if (search) {
    for (const [key, value] of Object.entries(search)) {
      if (value !== undefined) {
        url.searchParams.set(key, String(value))
      }
    }
  }

  if (baseUrl.startsWith('/')) {
    return `${url.pathname}${url.search}`
  }

  return url.toString()
}

export function createApiClient(options: ApiClientOptions = {}) {
  const request = async <TResponse, TBody = undefined>({
    path,
    method = 'GET',
    body,
    search,
    headers,
  }: RequestOptions<TBody>): Promise<TResponse> => {
    const fetcher = options.fetch ?? fetch
    const requestHeaders = new Headers(options.headers)

    if (headers) {
      new Headers(headers).forEach((value, key) => requestHeaders.set(key, value))
    }

    if (body !== undefined) {
      requestHeaders.set('Content-Type', 'application/json')
    }

    if (method !== 'GET') {
      const csrfToken = options.getCsrfToken?.()
      if (csrfToken) {
        requestHeaders.set('X-CSRF-Token', csrfToken)
      }
    }

    const response = await fetcher(buildUrl(options.baseUrl ?? '/api/v1', path, search), {
      method,
      headers: requestHeaders,
      body: body === undefined ? undefined : JSON.stringify(body),
      credentials: 'include',
    })

    if (response.status === 204) {
      return undefined as TResponse
    }

    const contentType = response.headers.get('content-type') ?? ''
    const payload = contentType.includes('application/json') ? await response.json() : undefined

    if (!response.ok) {
      if (payload) {
        throw new ApiError(payload as ApiErrorPayload)
      }

      throw new Error(`Request failed with status ${response.status}`)
    }

    return payload as TResponse
  }

  const context: RequestContext = { request }

  return {
    auth: createAuthClient(context),
    recipes: createRecipesClient(context),
  }
}

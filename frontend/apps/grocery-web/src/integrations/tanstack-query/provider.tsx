import { QueryClient } from '@tanstack/solid-query'

export type AppRouterContext = {
  queryClient: QueryClient
}

export function getContext(): AppRouterContext {
  const queryClient = new QueryClient()
  return {
    queryClient,
  }
}

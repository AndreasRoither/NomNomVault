import type { components } from '../generated/schema'
import type { RequestContext } from './http'

export type RecipeListParams = {
  cursor?: string
  limit?: number
  q?: string
  mealType?: string
  region?: string
  sort?: string
}

export type RecipeListResponse = components['schemas']['httpapi.RecipeListResponse']
export type RecipeDetailResponse = components['schemas']['httpapi.RecipeDetailResponse']

export function createRecipesClient(context: RequestContext) {
  return {
    list: (params: RecipeListParams = {}) =>
      context.request<RecipeListResponse>({
        path: '/recipes',
        search: params,
      }),
    detail: (recipeId: string) =>
      context.request<RecipeDetailResponse>({
        path: `/recipes/${recipeId}`,
      }),
  }
}

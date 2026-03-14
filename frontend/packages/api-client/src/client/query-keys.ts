export type RecipeListParams = {
  cursor?: string
  limit?: number
}

export const apiQueryKeys = {
  auth: {
    session: () => ['auth', 'session'] as const,
  },
  recipes: {
    list: (params: RecipeListParams = {}) => ['recipes', 'list', params] as const,
    detail: (recipeId: string) => ['recipes', 'detail', recipeId] as const,
  },
}

import {
	type ApiClient,
	apiQueryKeys,
	type RecipeListParams,
	type RecipeListResponse,
} from "@nomnomvault/api-client";

export function recipeListQueryOptions(
	api: ApiClient,
	params: RecipeListParams = {},
) {
	return {
		queryKey: apiQueryKeys.recipes.list(params),
		staleTime: 30_000,
		queryFn: async (): Promise<RecipeListResponse> => api.recipes.list(params),
	};
}

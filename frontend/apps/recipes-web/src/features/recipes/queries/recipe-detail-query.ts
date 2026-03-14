import {
	type ApiClient,
	apiQueryKeys,
	type RecipeDetailResponse,
} from "@nomnomvault/api-client";

export function recipeDetailQueryOptions(api: ApiClient, recipeId: string) {
	return {
		queryKey: apiQueryKeys.recipes.detail(recipeId),
		staleTime: 30_000,
		queryFn: async (): Promise<RecipeDetailResponse> =>
			api.recipes.detail(recipeId),
	};
}

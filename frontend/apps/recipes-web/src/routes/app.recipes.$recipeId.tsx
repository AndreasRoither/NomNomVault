import { ApiError } from "@nomnomvault/api-client";
import { createFileRoute, notFound } from "@tanstack/solid-router";
import { onMount } from "solid-js";

import { RecipeDetailPage } from "../features/recipes/detail/RecipeDetailPage";
import { recipeDetailQueryOptions } from "../features/recipes/queries/recipe-detail-query";
import { writeRecentRecipeId } from "../features/recipes/shared/recent-recipes";
import { mapRecipeDetailToVM } from "../features/recipes/view-models/map-recipe-detail";

export const Route = createFileRoute("/app/recipes/$recipeId")({
	loader: async ({ context, params }) => {
		if (context.backendAvailable === false) {
			return null;
		}

		try {
			const response = await context.queryClient.ensureQueryData(
				recipeDetailQueryOptions(context.apiClient, params.recipeId),
			);

			return mapRecipeDetailToVM(response);
		} catch (error) {
			if (error instanceof ApiError && error.status === 404) {
				throw notFound();
			}

			throw error;
		}
	},
	pendingComponent: RecipeDetailPending,
	errorComponent: RecipeDetailError,
	component: RecipeDetailRouteComponent,
});

function RecipeDetailRouteComponent() {
	const recipe = Route.useLoaderData();

	if (recipe() == null) {
		return <RecipeDetailError />;
	}

	onMount(() => {
		writeRecentRecipeId(recipe().id);
	});

	return <RecipeDetailPage recipe={recipe()} />;
}

function RecipeDetailPending() {
	return (
		<main class="mx-auto grid w-[min(1180px,100%)] gap-5 py-2 pb-8">
			<section class="grid justify-items-center gap-4 px-5 py-8 text-center">
				<div class="grid max-w-xl gap-2">
					<h2>Loading recipe…</h2>
					<p>Fetching the recipe details, ingredients, and cooking steps.</p>
				</div>
			</section>
		</main>
	);
}

function RecipeDetailError() {
	return (
		<main class="mx-auto grid w-[min(1180px,100%)] gap-5 py-2 pb-8">
			<section class="grid justify-items-center gap-4 px-5 py-8 text-center">
				<div class="grid max-w-xl gap-2">
					<h2>The recipe could not be loaded.</h2>
					<p>Refresh the page or try again after the data service recovers.</p>
				</div>
			</section>
		</main>
	);
}

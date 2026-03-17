import { createFileRoute } from "@tanstack/solid-router";
import { z } from "zod";

import { RECIPE_SORT_VALUES } from "../features/recipes/config/recipe-filters";
import { createRecipesHomeController } from "../features/recipes/home/create-recipes-home-controller";
import { RecipesHomePage } from "../features/recipes/home/RecipesHomePage";
import { recipeListQueryOptions } from "../features/recipes/queries/recipe-list-query";
import { mapRecipeListToCards } from "../features/recipes/view-models/map-recipe-list";

const recipesSearchSchema = z.object({
	q: z.string().optional(),
	mealType: z.string().optional(),
	region: z.string().optional(),
	sort: z.enum(RECIPE_SORT_VALUES).optional().catch("recent"),
});

export const Route = createFileRoute("/app/recipes/")({
	validateSearch: recipesSearchSchema,
	loaderDeps: ({ search }) => ({
		q: search.q,
		mealType: search.mealType,
		region: search.region,
		sort: search.sort ?? "recent",
	}),
	loader: async ({ context, deps }) => {
		if (context.backendAvailable === false) {
			return {
				cards: [],
				search: deps,
			};
		}

		const response = await context.queryClient.ensureQueryData(
			recipeListQueryOptions(context.apiClient, {
				q: deps.q,
				mealType: deps.mealType,
				region: deps.region,
				sort: deps.sort,
			}),
		);

		return {
			cards: mapRecipeListToCards(response),
			search: deps,
		};
	},
	pendingComponent: RecipesIndexPending,
	errorComponent: RecipesIndexError,
	component: RecipesIndexPage,
});

function RecipesIndexPage() {
	const context = Route.useRouteContext();
	const navigate = Route.useNavigate();
	const data = Route.useLoaderData();

	const controller = createRecipesHomeController({
		cards: () => data().cards,
		search: () => data().search,
		navigate: (patch) =>
			void navigate({
				to: "/app/recipes",
				search: (previous) => ({
					...previous,
					...patch,
				}),
				replace: true,
			}),
	});

	return (
		<RecipesHomePage
			query={controller.query()}
			sort={data().search.sort}
			mealType={data().search.mealType}
			region={data().search.region}
			collections={controller.visibleCollections()}
			hasActiveFilters={controller.hasActiveFilters()}
			onQueryInput={controller.setQuery}
			onSortChange={controller.setSort}
			onMealTypeChange={controller.setMealType}
			onRegionChange={controller.setRegion}
		/>
	);
}

function RecipesIndexPending() {
	return (
		<main class="grid gap-4 py-2 pb-8">
			<section class="grid justify-items-center gap-4 px-5 py-8 text-center">
				<div class="grid max-w-xl gap-2">
					<h2>Loading recipes…</h2>
					<p>Preparing the workspace and fetching the latest recipe list.</p>
				</div>
			</section>
		</main>
	);
}

function RecipesIndexError() {
	return (
		<main class="grid gap-4 py-2 pb-8">
			<section class="grid justify-items-center gap-4 px-5 py-8 text-center">
				<div class="grid max-w-xl gap-2">
					<h2>The recipe workspace could not be loaded.</h2>
					<p>Try refreshing the page once the API is available again.</p>
				</div>
			</section>
		</main>
	);
}

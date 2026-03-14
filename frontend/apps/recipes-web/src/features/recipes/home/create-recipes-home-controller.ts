import { createEffect, createSignal, onMount } from "solid-js";
import type { RecipeSortOption } from "../config/recipe-filters";
import { readRecentRecipeIds } from "../shared/recent-recipes";
import { deriveRecipeCollections } from "../view-models/derive-recipe-collections";
import type { RecipeCardVM, RecipeCollectionVM } from "../view-models/types";
import { createQuerySync } from "./create-query-sync";
import { resolveRecentCards } from "./resolve-recent-cards";

type RecipesHomeSearch = {
	q?: string;
	mealType?: string;
	region?: string;
	sort: RecipeSortOption;
};

type CreateRecipesHomeControllerArgs = {
	cards: () => RecipeCardVM[];
	search: () => RecipesHomeSearch;
	navigate: (patch: Partial<RecipesHomeSearch>) => void;
};

export function createRecipesHomeController(
	args: CreateRecipesHomeControllerArgs,
) {
	const [hasHydrated, setHasHydrated] = createSignal(false);
	const [recentCards, setRecentCards] = createSignal<RecipeCardVM[]>([]);
	const querySync = createQuerySync({
		value: () => args.search().q ?? "",
		onCommit: (value) => {
			args.navigate({
				q: value || undefined,
			});
		},
	});

	onMount(() => {
		setHasHydrated(true);
		setRecentCards(resolveRecentCards(readRecentRecipeIds(), args.cards()));
	});

	createEffect(() => {
		if (!hasHydrated()) {
			return;
		}

		setRecentCards(resolveRecentCards(readRecentRecipeIds(), args.cards()));
	});

	const collections = (): RecipeCollectionVM[] =>
		deriveRecipeCollections({
			cards: args.cards(),
			recentlyViewed: recentCards(),
		});

	const visibleCollections = () =>
		collections().filter((collection) => collection.recipes.length > 0);

	const hasActiveFilters = () =>
		Boolean(args.search().q || args.search().mealType || args.search().region);

	return {
		query: querySync.query,
		visibleCollections,
		hasActiveFilters,
		setQuery: querySync.setQuery,
		setSort: (value: RecipeSortOption) => args.navigate({ sort: value }),
		setMealType: (value?: string) => args.navigate({ mealType: value }),
		setRegion: (value?: string) => args.navigate({ region: value }),
	};
}

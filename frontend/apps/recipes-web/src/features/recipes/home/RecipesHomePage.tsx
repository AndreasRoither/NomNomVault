import { For } from "solid-js";
import {
	RECIPE_MEAL_TYPE_OPTIONS,
	RECIPE_REGION_OPTIONS,
	type RecipeSortOption,
} from "../config/recipe-filters";
import type { RecipeCollectionVM } from "../view-models/types";
import { PantryPromptCard } from "./PantryPromptCard";
import { QuickSearchBar } from "./QuickSearchBar";
import { RecipeCollectionRail } from "./RecipeCollectionRail";
import { RecipeEmptyState } from "./RecipeEmptyState";
import { RecipeFiltersPanel } from "./RecipeFiltersPanel";

type RecipesHomePageProps = {
	query?: string;
	sort: RecipeSortOption;
	mealType?: string;
	region?: string;
	collections: RecipeCollectionVM[];
	hasActiveFilters: boolean;
	onQueryInput: (value: string) => void;
	onSortChange: (value: RecipeSortOption) => void;
	onMealTypeChange: (value?: string) => void;
	onRegionChange: (value?: string) => void;
};

export function RecipesHomePage(props: RecipesHomePageProps) {
	return (
		<main class="grid gap-4 py-2 pb-8">
			<QuickSearchBar
				query={props.query}
				sort={props.sort}
				onQueryInput={props.onQueryInput}
				onSortChange={props.onSortChange}
				renderMobileFilters={() => (
					<RecipeFiltersPanel
						showSort
						sort={props.sort}
						mealType={props.mealType}
						region={props.region}
						mealTypes={RECIPE_MEAL_TYPE_OPTIONS}
						regions={RECIPE_REGION_OPTIONS}
						onSortChange={props.onSortChange}
						onMealTypeChange={props.onMealTypeChange}
						onRegionChange={props.onRegionChange}
					/>
				)}
			/>
			<RecipeFiltersPanel
				sort={props.sort}
				mealType={props.mealType}
				region={props.region}
				mealTypes={RECIPE_MEAL_TYPE_OPTIONS}
				regions={RECIPE_REGION_OPTIONS}
				onSortChange={props.onSortChange}
				onMealTypeChange={props.onMealTypeChange}
				onRegionChange={props.onRegionChange}
			/>
			<PantryPromptCard />
			{props.collections.length > 0 ? (
				<For each={props.collections}>
					{(collection) => <RecipeCollectionRail collection={collection} />}
				</For>
			) : (
				<RecipeEmptyState hasActiveFilters={props.hasActiveFilters} />
			)}
		</main>
	);
}

import type { RecipeSortOption } from "../config/recipe-filters";

import { FilterChipRow } from "./FilterChipRow";
import { RecipeSortControl } from "./RecipeSortControl";

type FilterOption = {
	value: string;
	label: string;
};

type RecipeFiltersPanelProps = {
	sort: RecipeSortOption;
	mealType?: string;
	region?: string;
	mealTypes: FilterOption[];
	regions: FilterOption[];
	showSort?: boolean;
	onSortChange: (value: RecipeSortOption) => void;
	onMealTypeChange: (value?: string) => void;
	onRegionChange: (value?: string) => void;
};

export function RecipeFiltersPanel(props: RecipeFiltersPanelProps) {
	return (
		<>
			{props.showSort ? (
				<RecipeSortControl
					sort={props.sort}
					label="Sort"
					onSortChange={props.onSortChange}
				/>
			) : null}
			<FilterChipRow
				title="Categories"
				items={props.mealTypes}
				selected={props.mealType}
				onSelect={props.onMealTypeChange}
			/>
			<FilterChipRow
				title="Regions"
				items={props.regions}
				selected={props.region}
				onSelect={props.onRegionChange}
			/>
		</>
	);
}

import { Select, SelectItem, SelectTrigger } from "@nomnomvault/ui";
import {
	isRecipeSortOption,
	RECIPE_SORT_OPTIONS,
	type RecipeSortOption,
} from "../config/recipe-filters";

type RecipeSortControlProps = {
	sort: RecipeSortOption;
	onSortChange: (value: RecipeSortOption) => void;
	class?: string;
	label?: string;
};

export function RecipeSortControl(props: RecipeSortControlProps) {
	return (
		<section class={props.label ? "grid gap-2" : undefined}>
			{props.label ? (
				<div class="text-sm font-bold text-[var(--nnv-text-muted)]">
					{props.label}
				</div>
			) : null}
			<Select
				class={props.class}
				value={props.sort}
				onValueChange={(value) => {
					if (isRecipeSortOption(value)) {
						props.onSortChange(value);
					}
				}}
			>
				<SelectTrigger aria-label="Sort recipes">
					{RECIPE_SORT_OPTIONS.map((option) => (
						<SelectItem value={option.value}>{option.label}</SelectItem>
					))}
				</SelectTrigger>
			</Select>
		</section>
	);
}

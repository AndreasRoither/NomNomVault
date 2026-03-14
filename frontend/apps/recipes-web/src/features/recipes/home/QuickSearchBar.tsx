import {
	Button,
	ClearIcon,
	FilterIcon,
	SearchIcon,
	Sheet,
	SheetContent,
	SheetDescription,
	SheetHeader,
	SheetTitle,
	SheetTrigger,
	TextField,
	TextFieldInput,
} from "@nomnomvault/ui";
import type { JSX } from "solid-js";
import type { RecipeSortOption } from "../config/recipe-filters";
import { RecipeSortControl } from "./RecipeSortControl";

type QuickSearchBarProps = {
	query?: string;
	sort: RecipeSortOption;
	onQueryInput: (value: string) => void;
	onSortChange: (value: RecipeSortOption) => void;
	renderMobileFilters?: () => JSX.Element;
};

export function QuickSearchBar(props: QuickSearchBarProps) {
	return (
		<section class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
			<TextField>
				<SearchIcon size="sm" />
				<TextFieldInput
					aria-label="Search recipes"
					placeholder="Search recipes, ingredients, cuisines"
					value={props.query ?? ""}
					onInput={(event) => props.onQueryInput(event.currentTarget.value)}
				/>
				{props.query ? (
					<Button
						variant="quiet"
						size="sm"
						aria-label="Clear search"
						onClick={() => props.onQueryInput("")}
					>
						<ClearIcon size="sm" />
					</Button>
				) : null}
			</TextField>
			<div class="flex items-center gap-2.5">
				<RecipeSortControl
					class="hidden min-w-44 md:block [&_select]:min-h-12"
					sort={props.sort}
					onSortChange={props.onSortChange}
				/>
				<Sheet>
					<SheetTrigger class="inline-flex w-full items-center justify-center gap-2 rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] px-3.5 py-3 font-bold text-[var(--nnv-text-strong)] md:hidden">
						<FilterIcon size="md" />
						<span>Filters</span>
					</SheetTrigger>
					<SheetContent>
						<SheetHeader>
							<SheetTitle>Refine recipes</SheetTitle>
							<SheetDescription>
								Adjust sort, category, and region filters.
							</SheetDescription>
						</SheetHeader>
						<div class="grid gap-4">{props.renderMobileFilters?.()}</div>
					</SheetContent>
				</Sheet>
			</div>
		</section>
	);
}

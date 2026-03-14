import { formatEnumLabel } from "../view-models/formatters";

export const RECIPE_SORT_OPTIONS = [
	{ value: "recent", label: "Recent" },
	{ value: "quick", label: "Quick" },
	{ value: "popular", label: "Popular" },
	{ value: "newest", label: "Newest" },
] as const;

export const RECIPE_SORT_VALUES = RECIPE_SORT_OPTIONS.map(
	(option) => option.value,
) as [
	(typeof RECIPE_SORT_OPTIONS)[number]["value"],
	...(typeof RECIPE_SORT_OPTIONS)[number]["value"][],
];

export type RecipeSortOption = (typeof RECIPE_SORT_OPTIONS)[number]["value"];

export function isRecipeSortOption(value: string): value is RecipeSortOption {
	return RECIPE_SORT_VALUES.includes(value as RecipeSortOption);
}

// todo: remove after load from backend is in place
const MEAL_TYPE_VALUES = [
	"breakfast",
	"lunch",
	"dinner",
	"snack",
	"dessert",
	"appetizer",
	"side_dish",
	"beverage",
] as const;

const REGION_VALUES = [
	"europe",
	"asia",
	"middle_east",
	"north_america",
	"south_america",
	"africa",
	"oceania",
] as const;

export const RECIPE_MEAL_TYPE_OPTIONS = MEAL_TYPE_VALUES.map((value) => ({
	value,
	label: formatEnumLabel(value) ?? value,
}));

export const RECIPE_REGION_OPTIONS = REGION_VALUES.map((value) => ({
	value,
	label: formatEnumLabel(value) ?? value,
}));

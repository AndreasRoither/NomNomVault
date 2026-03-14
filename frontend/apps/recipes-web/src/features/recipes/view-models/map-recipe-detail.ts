import type { RecipeDetailResponse } from "@nomnomvault/api-client";

import {
	formatCalories,
	formatDateLabel,
	formatEnumLabel,
	formatMinutes,
	formatQuantity,
	formatServings,
} from "./formatters";
import type {
	AllergenStateVM,
	NutrientRowVM,
	RecipeDetailVM,
	RecipeFactsVM,
	RecipeStatVM,
} from "./types";

function sortBySortOrder<T extends { sortOrder?: number }>(
	items: T[] | undefined,
) {
	return (items ?? [])
		.slice()
		.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
}

function buildAllergenState(items?: string[]): AllergenStateVM {
	if (!items) {
		return { kind: "unknown" };
	}
	if (items.length === 0) {
		return { kind: "none-recorded" };
	}
	return { kind: "present", items };
}

function buildStats(response: RecipeDetailResponse): RecipeStatVM[] {
	const recipe = response.recipe;
	const prep = recipe?.prepMinutes;
	const cook = recipe?.cookMinutes;
	const total = (prep ?? 0) + (cook ?? 0);
	const stats: RecipeStatVM[] = [];

	if (prep) {
		stats.push({
			key: "prep",
			label: "Prep",
			value: formatMinutes(prep) ?? `${prep} min`,
		});
	}
	if (cook) {
		stats.push({
			key: "cook",
			label: "Cook",
			value: formatMinutes(cook) ?? `${cook} min`,
		});
	}
	if (total > 0) {
		stats.push({
			key: "total",
			label: "Total",
			value: formatMinutes(total) ?? `${total} min`,
		});
	}
	if (recipe?.servings) {
		stats.push({
			key: "serves",
			label: "Serves",
			value: formatServings(recipe.servings) ?? `${recipe.servings}`,
		});
	}
	if (recipe?.caloriesPerServe) {
		stats.push({
			key: "calories",
			label: "Calories",
			value:
				formatCalories(recipe.caloriesPerServe) ?? `${recipe.caloriesPerServe}`,
		});
	}

	return stats;
}

function buildNutrients(response: RecipeDetailResponse) {
	const entry = response.nutritionEntries?.[0];
	if (!entry) {
		return null;
	}

	const rows: NutrientRowVM[] = [];
	const pushRow = (
		key: string,
		label: string,
		value?: number,
		suffix = "g",
	) => {
		if (value == null) {
			return;
		}

		rows.push({
			key,
			label,
			value: suffix ? `${value}${suffix}` : String(value),
		});
	};

	pushRow("calories", "Calories", entry.energyKcal, " kcal");
	pushRow("protein", "Protein", entry.protein);
	pushRow("carbohydrates", "Carbs", entry.carbohydrates);
	pushRow("fat", "Fat", entry.fat);
	pushRow("saturatedFat", "Sat fat", entry.saturatedFat);
	pushRow("fiber", "Fiber", entry.fiber);
	pushRow("sugars", "Sugars", entry.sugars);
	pushRow("sodium", "Sodium", entry.sodium);
	pushRow("salt", "Salt", entry.salt);

	return {
		referenceLabel: entry.referenceQuantity,
		rows,
	};
}

function buildGallery(response: RecipeDetailResponse) {
	const fallbackAlt = response.recipe?.title ?? "Recipe image";

	return sortBySortOrder(response.mediaAssets).map((asset) => ({
		id: asset.id ?? "media",
		url: asset.url ?? asset.thumbnailUrl ?? "",
		thumbnailUrl: asset.thumbnailUrl,
		alt: asset.altText ?? fallbackAlt,
	}));
}

function buildFacts(recipe: RecipeDetailResponse["recipe"]): RecipeFactsVM {
	return {
		region: formatEnumLabel(recipe?.region),
		cuisine: formatEnumLabel(recipe?.cuisine),
		difficulty: formatEnumLabel(recipe?.difficulty),
		mealType: formatEnumLabel(recipe?.mealType),
	};
}

function buildIngredients(response: RecipeDetailResponse) {
	return sortBySortOrder(response.ingredients).map((ingredient) => ({
		id: ingredient.id ?? "ingredient",
		quantity: formatQuantity(ingredient.quantity, ingredient.unit),
		name: ingredient.name ?? "Ingredient",
		preparation: ingredient.preparation,
	}));
}

function buildMethod(response: RecipeDetailResponse) {
	return sortBySortOrder(response.steps).map((step, index) => ({
		id: step.id ?? `step-${index + 1}`,
		stepNumber: index + 1,
		instruction: step.instruction ?? "",
		durationLabel: formatMinutes(step.durationMinutes),
		tip: step.tip,
	}));
}

function buildSource(recipe: RecipeDetailResponse["recipe"]) {
	return {
		url: recipe?.sourceUrl,
		capturedAtLabel: formatDateLabel(recipe?.sourceCapturedAt),
		versionLabel: recipe?.version ? `Version ${recipe.version}` : undefined,
	};
}

export function mapRecipeDetailToVM(
	response: RecipeDetailResponse,
): RecipeDetailVM {
	const recipe = response.recipe;
	const gallery = buildGallery(response);

	return {
		id: recipe?.id ?? "recipe",
		title: recipe?.title ?? "Untitled recipe",
		summary: recipe?.description,
		breadcrumbs: [
			{ label: "Recipes", to: "/app/recipes" },
			{ label: recipe?.title ?? "Recipe" },
		],
		gallery,
		stats: buildStats(response),
		allergens: buildAllergenState(recipe?.allergens),
		facts: buildFacts(recipe),
		nutrients: buildNutrients(response),
		ingredients: buildIngredients(response),
		method: buildMethod(response),
		pantryMatch: { kind: "unavailable" },
		source: buildSource(recipe),
	};
}

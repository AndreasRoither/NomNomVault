import type { RecipeListResponse } from "@nomnomvault/api-client";

import {
	formatCalories,
	formatEnumLabel,
	formatMinutes,
	formatServings,
} from "./formatters";
import type { RecipeCardVM } from "./types";

function getTotalMinutes(prep?: number, cook?: number) {
	const total = (prep ?? 0) + (cook ?? 0);
	return total > 0 ? total : undefined;
}

export function mapRecipeListToCards(
	response: RecipeListResponse,
): RecipeCardVM[] {
	return (response.data ?? []).map((recipe) => {
		const totalMinutes = getTotalMinutes(
			recipe.prepMinutes,
			recipe.cookMinutes,
		);

		return {
			id: recipe.id ?? "recipe",
			title: recipe.title ?? "Untitled recipe",
			summary: recipe.description ?? "No summary available yet.",
			metrics: {
				prepMinutes: recipe.prepMinutes ?? undefined,
				cookMinutes: recipe.cookMinutes ?? undefined,
				totalMinutes,
				servings: recipe.servings ?? undefined,
				caloriesPerServe: recipe.caloriesPerServe ?? undefined,
			},
			taxonomy: {
				mealType: recipe.mealType ?? undefined,
				region: recipe.region ?? undefined,
				difficulty: recipe.difficulty ?? undefined,
			},
			display: {
				prepLabel: formatMinutes(recipe.prepMinutes),
				cookLabel: formatMinutes(recipe.cookMinutes),
				totalLabel: formatMinutes(totalMinutes),
				servingsLabel: formatServings(recipe.servings),
				categoryLabel: formatEnumLabel(recipe.mealType),
				regionLabel: formatEnumLabel(recipe.region),
				difficultyLabel: formatEnumLabel(recipe.difficulty),
				caloriesLabel: formatCalories(recipe.caloriesPerServe),
			},
		};
	});
}

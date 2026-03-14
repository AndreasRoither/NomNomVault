import type { RecipeCardVM } from "../view-models/types";

export function resolveRecentCards(
	recentIds: string[],
	cards: RecipeCardVM[],
): RecipeCardVM[] {
	return recentIds
		.map((id) => cards.find((card) => card.id === id))
		.filter((card): card is RecipeCardVM => Boolean(card));
}

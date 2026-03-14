import type { RecipeCardVM, RecipeCollectionVM } from "./types";

type DeriveCollectionsArgs = {
	cards: RecipeCardVM[];
	recentlyViewed?: RecipeCardVM[];
};

export function deriveRecipeCollections(
	args: DeriveCollectionsArgs,
): RecipeCollectionVM[] {
	const quickWeeknight = args.cards.filter((card) => {
		const minutes = card.metrics.totalMinutes;
		return Boolean(minutes && minutes > 0 && minutes <= 35);
	});

	const collections: RecipeCollectionVM[] = [
		{
			id: "quick-weeknight",
			title: "Quick weeknight",
			description: "Fast, practical picks for busy evenings.",
			recipes:
				quickWeeknight.length > 0 ? quickWeeknight : args.cards.slice(0, 3),
		},
		{
			id: "newly-added",
			title: "Newly added",
			description: "Fresh additions to the household recipe box.",
			recipes: args.cards.slice(0, 3),
		},
	];

	if ((args.recentlyViewed?.length ?? 0) > 0) {
		collections.push({
			id: "recently-viewed",
			title: "Recently viewed",
			description: "Jump back into recipes you opened most recently.",
			recipes: args.recentlyViewed ?? [],
		});
	}

	return collections;
}

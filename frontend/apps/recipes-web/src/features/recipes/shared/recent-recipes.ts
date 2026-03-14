const STORAGE_KEY = "nnv-recent-recipes";
const MAX_ITEMS = 6;

export function readRecentRecipeIds() {
	if (typeof window === "undefined") {
		return [] as string[];
	}

	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		if (!raw) {
			return [];
		}

		const parsed = JSON.parse(raw);
		return Array.isArray(parsed)
			? parsed.filter((value): value is string => typeof value === "string")
			: [];
	} catch {
		return [];
	}
}

export function writeRecentRecipeId(recipeId: string) {
	if (typeof window === "undefined") {
		return;
	}

	const deduped = [
		recipeId,
		...readRecentRecipeIds().filter((value) => value !== recipeId),
	];
	window.localStorage.setItem(
		STORAGE_KEY,
		JSON.stringify(deduped.slice(0, MAX_ITEMS)),
	);
}

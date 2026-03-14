export type RecipeCardVM = {
	id: string;
	title: string;
	summary: string;
	metrics: {
		prepMinutes?: number;
		cookMinutes?: number;
		totalMinutes?: number;
		servings?: number;
		caloriesPerServe?: number;
	};
	taxonomy: {
		mealType?: string;
		region?: string;
		difficulty?: string;
	};
	display: {
		prepLabel?: string;
		cookLabel?: string;
		totalLabel?: string;
		servingsLabel?: string;
		categoryLabel?: string;
		regionLabel?: string;
		difficultyLabel?: string;
		caloriesLabel?: string;
	};
};

export type RecipeStatVM = {
	key: "prep" | "cook" | "total" | "serves" | "calories";
	label: string;
	value: string;
};

export type AllergenStateVM =
	| { kind: "present"; items: string[] }
	| { kind: "none-recorded" }
	| { kind: "unknown" };

export type NutrientRowVM = {
	key: string;
	label: string;
	value: string;
};

export type RecipeFactsVM = {
	region?: string;
	cuisine?: string;
	difficulty?: string;
	mealType?: string;
};

export type PantryMatchVM =
	| { kind: "unavailable" }
	| { kind: "empty-pantry" }
	| { kind: "partial"; matched: number; total: number }
	| { kind: "strong"; matched: number; total: number };

export type RecipeDetailVM = {
	id: string;
	title: string;
	summary?: string;
	breadcrumbs: { label: string; to?: string }[];
	gallery: { id: string; url: string; alt: string; thumbnailUrl?: string }[];
	stats: RecipeStatVM[];
	allergens: AllergenStateVM;
	facts: RecipeFactsVM;
	nutrients: { referenceLabel?: string; rows: NutrientRowVM[] } | null;
	ingredients: {
		id: string;
		quantity?: string;
		name: string;
		preparation?: string;
	}[];
	method: {
		id: string;
		stepNumber: number;
		instruction: string;
		durationLabel?: string;
		tip?: string;
	}[];
	pantryMatch: PantryMatchVM;
	source?: { url?: string; capturedAtLabel?: string; versionLabel?: string };
};

export type RecipeCollectionVM = {
	id: string;
	title: string;
	description: string;
	recipes: RecipeCardVM[];
};

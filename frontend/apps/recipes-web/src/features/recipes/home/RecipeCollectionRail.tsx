import { For } from "solid-js";

import { RecipeCard } from "../shared/RecipeCard";
import type { RecipeCollectionVM } from "../view-models/types";

type RecipeCollectionRailProps = {
	collection: RecipeCollectionVM;
};

export function RecipeCollectionRail(props: RecipeCollectionRailProps) {
	return (
		<section class="grid gap-3 pt-1">
			<div class="grid gap-1">
				<div>
					<h2 class="m-0 text-[1.45rem]">{props.collection.title}</h2>
					<p class="m-0 text-[var(--nnv-text-muted)]">
						{props.collection.description}
					</p>
				</div>
			</div>
			<div class="grid grid-flow-col auto-cols-[minmax(16.5rem,84%)] gap-4 overflow-x-auto pb-1 md:grid-flow-row md:grid-cols-2 md:auto-cols-auto md:overflow-visible xl:grid-cols-3">
				<For each={props.collection.recipes}>
					{(recipe) => <RecipeCard recipe={recipe} />}
				</For>
			</div>
		</section>
	);
}

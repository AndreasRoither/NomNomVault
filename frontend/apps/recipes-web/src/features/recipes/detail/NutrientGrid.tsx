import { Card, CardContent, CardHeader, CardTitle } from "@nomnomvault/ui";
import { For, Show } from "solid-js";

import type { RecipeDetailVM } from "../view-models/types";

type NutrientGridProps = {
	nutrients: RecipeDetailVM["nutrients"];
};

export function NutrientGrid(props: NutrientGridProps) {
	return (
		<Card>
			<CardHeader>
				<p class="nnv-eyebrow">Nutrition</p>
				<CardTitle>Per-serving nutrients</CardTitle>
			</CardHeader>
			<CardContent>
				<Show
					when={props.nutrients}
					fallback={
						<p class="m-0 text-[var(--nnv-text-muted)]">
							No nutrition data is available for this recipe yet.
						</p>
					}
				>
					<table class="nnv-table">
						<Show when={props.nutrients?.referenceLabel}>
							<caption>{props.nutrients?.referenceLabel}</caption>
						</Show>
						<thead>
							<tr>
								<th scope="col">Nutrient</th>
								<th scope="col">Value</th>
							</tr>
						</thead>
						<tbody>
							<For each={props.nutrients?.rows ?? []}>
								{(row) => (
									<tr>
										<td>{row.label}</td>
										<td>{row.value}</td>
									</tr>
								)}
							</For>
						</tbody>
					</table>
				</Show>
			</CardContent>
		</Card>
	);
}

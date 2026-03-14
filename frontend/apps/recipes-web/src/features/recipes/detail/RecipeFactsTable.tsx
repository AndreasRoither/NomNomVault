import { Card, CardContent, CardHeader, CardTitle } from "@nomnomvault/ui";
import { For } from "solid-js";
import type { RecipeFactsVM } from "../view-models/types";

type RecipeFactsTableProps = {
	facts: RecipeFactsVM;
};

export function RecipeFactsTable(props: RecipeFactsTableProps) {
	const rows = () =>
		[
			{ label: "Region", value: props.facts.region },
			{ label: "Cuisine", value: props.facts.cuisine },
			{ label: "Difficulty", value: props.facts.difficulty },
			{ label: "Meal type", value: props.facts.mealType },
		].filter((row): row is { label: string; value: string } =>
			Boolean(row.value),
		);

	return (
		<Card>
			<CardHeader>
				<p class="nnv-eyebrow">Recipe facts</p>
				<CardTitle>At a glance</CardTitle>
			</CardHeader>
			<CardContent>
				<table class="nnv-table">
					<thead>
						<tr>
							<th scope="col">Fact</th>
							<th scope="col">Value</th>
						</tr>
					</thead>
					<tbody>
						<For each={rows()}>
							{(fact) => (
								<tr>
									<td>{fact.label}</td>
									<td>{fact.value}</td>
								</tr>
							)}
						</For>
					</tbody>
				</table>
			</CardContent>
		</Card>
	);
}

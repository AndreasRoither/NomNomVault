import {
	Badge,
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from "@nomnomvault/ui";
import { For, Match, Switch } from "solid-js";

import type { AllergenStateVM } from "../view-models/types";

type AllergenPanelProps = {
	state: AllergenStateVM;
};

export function AllergenPanel(props: AllergenPanelProps) {
	return (
		<Card>
			<CardHeader>
				<p class="nnv-eyebrow">Safety first</p>
				<CardTitle>Allergens</CardTitle>
			</CardHeader>
			<CardContent>
				<Switch>
					<Match when={props.state.kind === "present"}>
						<div class="flex flex-wrap gap-2">
							<For
								each={props.state.kind === "present" ? props.state.items : []}
							>
								{(item) => <Badge tone="danger">{item}</Badge>}
							</For>
						</div>
					</Match>
					<Match when={props.state.kind === "none-recorded"}>
						<p class="m-0 text-[var(--nnv-text-muted)]">
							No recorded allergens for this recipe.
						</p>
					</Match>
					<Match when={props.state.kind === "unknown"}>
						<p class="m-0 text-[var(--nnv-text-muted)]">
							Allergen data is unavailable for this recipe.
						</p>
					</Match>
				</Switch>
			</CardContent>
		</Card>
	);
}

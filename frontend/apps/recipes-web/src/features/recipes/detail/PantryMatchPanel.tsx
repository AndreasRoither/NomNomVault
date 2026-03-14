import { Card, CardContent, CardHeader, CardTitle } from "@nomnomvault/ui";
import { Match, Switch } from "solid-js";

import type { PantryMatchVM } from "../view-models/types";

type PantryMatchPanelProps = {
	match: PantryMatchVM;
};

export function PantryMatchPanel(props: PantryMatchPanelProps) {
	return (
		<Card>
			<CardHeader>
				<p class="nnv-eyebrow">Pantry relevance</p>
				<CardTitle>Pantry match</CardTitle>
			</CardHeader>
			<CardContent>
				<Switch>
					<Match when={props.match.kind === "strong"}>
						<p>
							Strong match:{" "}
							{props.match.kind === "strong" ? props.match.matched : 0}/
							{props.match.kind === "strong" ? props.match.total : 0}{" "}
							ingredients ready.
						</p>
					</Match>
					<Match when={props.match.kind === "partial"}>
						<p>
							Almost there:{" "}
							{props.match.kind === "partial" ? props.match.matched : 0}/
							{props.match.kind === "partial" ? props.match.total : 0}{" "}
							ingredients ready.
						</p>
					</Match>
					<Match when={props.match.kind === "empty-pantry"}>
						<p>Your pantry is empty, so matching is unavailable.</p>
					</Match>
					<Match when={props.match.kind === "unavailable"}>
						<p>Pantry matching is not connected yet in this build.</p>
					</Match>
				</Switch>
			</CardContent>
		</Card>
	);
}

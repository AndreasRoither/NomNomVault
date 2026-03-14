import { For } from "solid-js";

import type { RecipeStatVM } from "../view-models/types";

type RecipeStatPillsProps = {
	stats: RecipeStatVM[];
};

export function RecipeStatPills(props: RecipeStatPillsProps) {
	return (
		<div class="flex flex-wrap gap-2.5">
			<For each={props.stats}>
				{(stat) => (
					<span class="inline-flex min-h-[2.6rem] items-center gap-1.5 rounded-[var(--nnv-radius-sm)] border border-[color:var(--nnv-line)] bg-transparent px-4 py-1.5 text-sm font-semibold text-[var(--nnv-text-strong)]">
						<strong>{stat.label}</strong>
						<span>{stat.value}</span>
					</span>
				)}
			</For>
		</div>
	);
}

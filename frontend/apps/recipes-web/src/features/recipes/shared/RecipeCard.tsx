import {
	Badge,
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@nomnomvault/ui";

import type { RecipeCardVM } from "../view-models/types";

type RecipeCardProps = {
	recipe: RecipeCardVM;
};

export function RecipeCard(props: RecipeCardProps) {
	return (
		<a
			href={`/app/recipes/${props.recipe.id}`}
			class="snap-start text-inherit no-underline"
		>
			<Card class="h-full">
				<div class="relative mx-3 mt-3 aspect-[4/3] overflow-hidden rounded-t-[var(--nnv-radius-lg)] bg-[linear-gradient(135deg,rgba(193,138,85,0.54),rgba(120,143,102,0.52))]">
					<div class="absolute inset-0 bg-[radial-gradient(circle_at_30%_30%,rgba(255,255,255,0.35),transparent_38%),linear-gradient(180deg,transparent_20%,rgba(0,0,0,0.16)_100%)]" />
				</div>
				<CardHeader>
					<div class="flex flex-wrap gap-2">
						{props.recipe.display.categoryLabel ? (
							<Badge>{props.recipe.display.categoryLabel}</Badge>
						) : null}
						{props.recipe.display.regionLabel ? (
							<Badge tone="accent">{props.recipe.display.regionLabel}</Badge>
						) : null}
					</div>
					<CardTitle class="line-clamp-2 text-[1.35rem]">
						{props.recipe.title}
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="m-0 line-clamp-3 text-[var(--nnv-text-muted)]">
						{props.recipe.summary}
					</p>
				</CardContent>
				<CardFooter class="gap-2 text-[0.88rem]">
					{props.recipe.display.totalLabel ? (
						<span>{props.recipe.display.totalLabel}</span>
					) : null}
					{props.recipe.display.servingsLabel ? (
						<span>{props.recipe.display.servingsLabel}</span>
					) : null}
					{props.recipe.display.caloriesLabel ? (
						<span>{props.recipe.display.caloriesLabel}</span>
					) : null}
				</CardFooter>
			</Card>
		</a>
	);
}

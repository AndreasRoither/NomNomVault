import {
	Breadcrumb,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbSeparator,
	Button,
	Card,
	CardContent,
	CardHeader,
	CardTitle,
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
	MoreIcon,
} from "@nomnomvault/ui";
import { For, Show } from "solid-js";
import type { RecipeDetailVM } from "../view-models/types";
import { AllergenPanel } from "./AllergenPanel";
import { NutrientGrid } from "./NutrientGrid";
import { PantryMatchPanel } from "./PantryMatchPanel";
import { RecipeFactsTable } from "./RecipeFactsTable";
import { RecipeHeroGallery } from "./RecipeHeroGallery";
import { RecipeStatPills } from "./RecipeStatPills";
import { SourceAttributionCard } from "./SourceAttributionCard";

type RecipeDetailPageProps = {
	recipe: RecipeDetailVM;
};

export function RecipeDetailPage(props: RecipeDetailPageProps) {
	return (
		<main class="mx-auto grid w-[min(1180px,100%)] gap-5 py-2 pb-8">
			<header class="grid gap-4">
				<div class="grid gap-4 xl:grid-cols-[minmax(0,820px)_280px] xl:justify-center xl:items-start">
					<div class="grid gap-4">
						<Breadcrumb>
							<BreadcrumbList>
								<For each={props.recipe.breadcrumbs}>
									{(crumb, index) => (
										<>
											<BreadcrumbItem>
												{crumb.to ? (
													<BreadcrumbLink href={crumb.to}>
														{crumb.label}
													</BreadcrumbLink>
												) : (
													<span>{crumb.label}</span>
												)}
											</BreadcrumbItem>
											<Show
												when={index() < props.recipe.breadcrumbs.length - 1}
											>
												<BreadcrumbSeparator>/</BreadcrumbSeparator>
											</Show>
										</>
									)}
								</For>
							</BreadcrumbList>
						</Breadcrumb>
						<div class="grid gap-4 md:items-start">
							<div>
								<h1 class="display-title m-0 text-[clamp(2.5rem,6vw,4.8rem)] leading-[0.96]">
									{props.recipe.title}
								</h1>
								<Show when={props.recipe.summary}>
									<p class="m-0 text-[var(--nnv-text-muted)]">
										{props.recipe.summary}
									</p>
								</Show>
							</div>
						</div>
						<RecipeStatPills stats={props.recipe.stats} />
					</div>

					<div class="grid justify-self-end xl:self-start">
						<div class="flex flex-wrap items-center justify-end gap-3">
							<Button variant="secondary">Save later</Button>
							<DropdownMenu>
								<DropdownMenuTrigger class="inline-flex h-[2.625rem] w-[2.625rem] items-center justify-center rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] p-0">
									<MoreIcon size="md" />
								</DropdownMenuTrigger>
								<DropdownMenuContent>
									<DropdownMenuItem>Share</DropdownMenuItem>
									<DropdownMenuItem>Print</DropdownMenuItem>
								</DropdownMenuContent>
							</DropdownMenu>
						</div>
					</div>
				</div>
			</header>

			<section class="grid gap-4 xl:grid-cols-[minmax(0,820px)_280px] xl:justify-center xl:items-start">
				<section class="grid gap-4">
					<RecipeHeroGallery
						title={props.recipe.title}
						items={props.recipe.gallery}
					/>

					<Card>
						<CardHeader>
							<p class="nnv-eyebrow">Ingredients</p>
							<CardTitle>What you need</CardTitle>
						</CardHeader>
						<CardContent class="grid gap-4">
							<For each={props.recipe.ingredients}>
								{(ingredient) => (
									<div class="grid gap-1 md:grid-cols-[minmax(5.5rem,7rem)_minmax(0,1fr)] md:items-start">
										<strong>{ingredient.quantity ?? "To taste"}</strong>
										<div>
											<div>{ingredient.name}</div>
											{ingredient.preparation ? (
												<span class="text-[var(--nnv-text-muted)]">
													{ingredient.preparation}
												</span>
											) : null}
										</div>
									</div>
								)}
							</For>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<p class="nnv-eyebrow">Method</p>
							<CardTitle>How to cook it</CardTitle>
						</CardHeader>
						<CardContent class="grid gap-4">
							<For each={props.recipe.method}>
								{(step) => (
									<article class="grid grid-cols-[auto_minmax(0,1fr)] gap-4">
										<div class="grid h-[2.3rem] w-[2.3rem] place-items-center rounded-[var(--nnv-radius-sm)] bg-[var(--nnv-chip-active)] font-extrabold">
											{step.stepNumber}
										</div>
										<div class="grid gap-2">
											<p class="m-0">{step.instruction}</p>
											{step.durationLabel ? (
												<span class="text-[var(--nnv-text-muted)]">
													{step.durationLabel}
												</span>
											) : null}
											{step.tip ? (
												<small class="m-0 text-[var(--nnv-text-muted)]">
													{step.tip}
												</small>
											) : null}
										</div>
									</article>
								)}
							</For>
						</CardContent>
					</Card>

					<SourceAttributionCard source={props.recipe.source} />
				</section>

				<aside class="grid gap-4 xl:sticky xl:top-[5.75rem]">
					<RecipeFactsTable facts={props.recipe.facts} />
					<AllergenPanel state={props.recipe.allergens} />
					<NutrientGrid nutrients={props.recipe.nutrients} />
					<PantryMatchPanel match={props.recipe.pantryMatch} />
				</aside>
			</section>
		</main>
	);
}

import { RecipesIcon } from "@nomnomvault/ui";

type RecipeEmptyStateProps = {
	hasActiveFilters: boolean;
};

export function RecipeEmptyState(props: RecipeEmptyStateProps) {
	return (
		<section class="grid justify-items-center gap-4 px-5 pt-9 pb-2 text-center">
			<div
				class="w-[min(100%,20rem)]"
				aria-hidden="true"
				data-slot="empty-illustration"
			>
				<div class="grid justify-items-center gap-4 rounded-[var(--nnv-radius-lg)] border border-dashed border-[color:color-mix(in_srgb,var(--nnv-line)_80%,transparent)] bg-[linear-gradient(180deg,color-mix(in_srgb,var(--nnv-surface-2)_72%,transparent),color-mix(in_srgb,var(--nnv-surface-1)_92%,transparent))] px-5 py-6">
					<div class="grid h-[4.75rem] w-[4.75rem] place-items-center rounded-[1.4rem] border border-[color:color-mix(in_srgb,var(--nnv-line)_82%,transparent)] bg-[color:color-mix(in_srgb,var(--nnv-surface-2)_78%,transparent)] text-[var(--nnv-accent-strong)]">
						<RecipesIcon class="h-8 w-8" size="lg" />
					</div>
					<div class="rounded-full border border-[color:color-mix(in_srgb,var(--nnv-line)_70%,transparent)] bg-[color:color-mix(in_srgb,var(--nnv-surface-1)_94%,transparent)] px-3 py-2 text-[0.82rem] font-semibold uppercase tracking-[0.12em] text-[var(--nnv-text-muted)]">
						{props.hasActiveFilters ? "No matches" : "No recipes yet"}
					</div>
				</div>
			</div>
			<div class="grid max-w-[32rem] gap-2">
				<h2 class="m-0 text-[1.55rem] font-bold">
					{props.hasActiveFilters
						? "No recipes match this setup yet."
						: "No recipes are available yet."}
				</h2>
				<p class="m-0 leading-[1.7] text-[var(--nnv-text-muted)]">
					{props.hasActiveFilters
						? "Try another search, clear a filter, or broaden the region or category."
						: "Add a few recipes and this chest will have something good to hold."}
				</p>
			</div>
		</section>
	);
}

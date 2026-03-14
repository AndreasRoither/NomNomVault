export function HomePage() {
	return (
		<main class="page-wrap grid min-h-[calc(100vh-4rem)] items-center py-8 pb-12">
			<section class="grid gap-5 rounded-[var(--nnv-radius-lg)] border border-[color:var(--nnv-line)] bg-[linear-gradient(180deg,var(--nnv-surface-1)_0%,var(--nnv-surface-2)_100%)] p-8 shadow-[var(--nnv-shadow-lg)]">
				<p class="nnv-eyebrow">Cooking workspace</p>
				<h1 class="display-title m-0 text-[clamp(2.5rem,6vw,4.8rem)] leading-[0.96]">
					Cook with a calmer recipe shell.
				</h1>
				<p class="m-0 max-w-[42rem] text-[1.05rem] leading-[1.8] text-[var(--nnv-text-muted)]">
					Search first, filter by meal and region, keep safety details visible,
					and move straight from home to a practical detail page.
				</p>
				<div class="flex flex-wrap gap-3">
					<a
						href="/app/recipes"
						class="inline-flex min-h-[2.9rem] items-center justify-center gap-2 rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-accent-strong)] bg-[var(--nnv-accent-strong)] px-[1.15rem] py-3 font-bold text-[#fffaf0] no-underline transition-transform duration-150 hover:-translate-y-px"
					>
						Open recipes
					</a>
					<a
						href="/app/about"
						class="inline-flex min-h-[2.9rem] items-center justify-center gap-2 rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] px-[1.15rem] py-3 font-bold text-[var(--nnv-text-strong)] no-underline transition-transform duration-150 hover:-translate-y-px"
					>
						About the app
					</a>
				</div>
			</section>
		</main>
	);
}

import { Card, CardContent, CardHeader, CardTitle } from "@nomnomvault/ui";
import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/app/about")({
	component: AboutPage,
});

function AboutPage() {
	return (
		<main class="grid gap-5 py-2 pb-8">
			<section class="grid gap-4">
				<p class="nnv-eyebrow">About</p>
			</section>

			<section class="grid grid-cols-[minmax(0,1fr)] gap-4">
				<Card>
					<CardHeader>
						<p class="nnv-eyebrow">Roadmap</p>
						<CardTitle>Next expansions</CardTitle>
					</CardHeader>
					<CardContent class="grid gap-3">
						<ul class="m-0 grid list-none gap-4 p-0">
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Auth and household foundation</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									Local sign-in, default household bootstrap, and stronger
									household-aware protection across the app.
								</span>
							</li>
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Core recipe vault</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									Create, edit, delete, and manage recipe details, ingredients,
									steps, tags, and media.
								</span>
							</li>
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Recipe import workflows</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									URL import, raw text import, and OCR for scanned or
									handwritten recipes.
								</span>
							</li>
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Pantry and planning</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									Pantry tracking with expiry awareness and weekly meal
									planning.
								</span>
							</li>
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Grocery workflow and offline support</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									Grocery list generation, mobile-friendly checklist flows, and
									offline shopping support.
								</span>
							</li>
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Household collaboration and release hardening</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									Invites, roles, collaboration tooling, plus security,
									observability, exports, and restore validation.
								</span>
							</li>
							<li class="grid gap-1 border-b border-[color:var(--nnv-line)] pb-4 last:border-b-0 last:pb-0">
								<strong>Post-v1 expansion</strong>
								<span class="leading-7 text-[var(--nnv-text-muted)]">
									Standalone grocery flows, cookbook/export improvements, and
									deferred platform features like OAuth/OIDC.
								</span>
							</li>
						</ul>
					</CardContent>
				</Card>
			</section>
		</main>
	);
}

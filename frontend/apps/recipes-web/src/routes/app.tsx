import { createFileRoute, Outlet, redirect } from "@tanstack/solid-router";

import { RecipesAppShell } from "../features/app-shell/RecipesAppShell";
import { sessionQueryOptions } from "../features/auth/session-query";

export const Route = createFileRoute("/app")({
	beforeLoad: async ({ context }) => {
		try {
			const session = await context.queryClient.ensureQueryData(
				sessionQueryOptions(context.apiClient),
			);

			if (session.authenticated !== true) {
				throw redirect({ to: "/" });
			}

			return { session, backendAvailable: true as const };
		} catch (error) {
			if (error instanceof Response) {
				throw error;
			}

			return { session: undefined, backendAvailable: false as const };
		}
	},
	component: AppLayout,
});

function AppLayout() {
	const context = Route.useRouteContext();

	return (
		<RecipesAppShell session={context().session}>
			<>
				{!context().backendAvailable ? <AppUnavailableBanner /> : null}
				<Outlet />
			</>
		</RecipesAppShell>
	);
}

function AppUnavailableBanner() {
	return (
		<section class="mb-4 grid gap-2 rounded-[var(--nnv-radius-lg)] border border-[color:var(--nnv-line)] bg-[linear-gradient(180deg,var(--nnv-surface-1)_0%,var(--nnv-surface-2)_100%)] p-5 shadow-[var(--nnv-shadow-sm)]">
			<p class="nnv-eyebrow">Offline</p>
			<p class="m-0 max-w-[42rem] text-[0.98rem] leading-[1.7] text-[var(--nnv-text-muted)]">
				The backend is offline.
			</p>
		</section>
	);
}

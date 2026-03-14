import { createFileRoute, Outlet, redirect } from "@tanstack/solid-router";

import { RecipesAppShell } from "../features/app-shell/RecipesAppShell";
import { sessionQueryOptions } from "../features/auth/session-query";

export const Route = createFileRoute("/app")({
	beforeLoad: async ({ context }) => {
		const session = await context.queryClient.ensureQueryData(
			sessionQueryOptions(context.apiClient),
		);

		if (session.authenticated !== true) {
			throw redirect({ to: "/" });
		}

		return { session };
	},
	component: AppLayout,
});

function AppLayout() {
	const context = Route.useRouteContext();

	return (
		<RecipesAppShell session={context().session}>
			<Outlet />
		</RecipesAppShell>
	);
}

import "@fontsource-variable/geist";
import { ThemeProvider, ThemeScript } from "@nomnomvault/app-shell";
import {
	createRootRouteWithContext,
	type ErrorComponentProps,
	HeadContent,
	Outlet,
	Scripts,
} from "@tanstack/solid-router";
import { Suspense } from "solid-js";
import { HydrationScript } from "solid-js/web";

import type { AppRouterContext } from "../integrations/tanstack-query/provider";

import styleCss from "../styles.css?url";

export const Route = createRootRouteWithContext<AppRouterContext>()({
	head: () => ({
		links: [{ rel: "stylesheet", href: styleCss }],
	}),
	shellComponent: RootComponent,
	errorComponent: RootErrorComponent,
});

function RootComponent() {
	return (
		<html lang="en">
			<head>
				<HeadContent />
				<ThemeScript />
				<HydrationScript />
			</head>
			<body>
				<ThemeProvider>
					<Suspense>
						<Outlet />
					</Suspense>
				</ThemeProvider>
				<Scripts />
			</body>
		</html>
	);
}

function RootErrorComponent(props: ErrorComponentProps) {
	const errorMessage = import.meta.env.DEV
		? props.error.message || String(props.error)
		: null;

	return (
		<main class="page-wrap py-12">
			<section class="grid gap-5 rounded-[var(--nnv-radius-lg)] border border-[color:var(--nnv-line)] bg-[linear-gradient(180deg,var(--nnv-surface-1)_0%,var(--nnv-surface-2)_100%)] p-8 shadow-[var(--nnv-shadow-lg)]">
				<p class="nnv-eyebrow">Application Error</p>
				<h1 class="display-title m-0 text-5xl">
					The page could not be rendered.
				</h1>
				<p class="m-0 max-w-[42rem] text-[1.05rem] leading-[1.8] text-[var(--nnv-text-muted)]">
					Refresh the page or try again after the current change finishes
					loading.
				</p>
				{errorMessage ? (
					<pre class="overflow-x-auto rounded-2xl border border-[var(--nnv-line)] bg-[var(--nnv-surface-2)] p-4 text-sm text-[var(--nnv-text-muted)]">
						<code>{errorMessage}</code>
					</pre>
				) : null}
			</section>
		</main>
	);
}

import { createFileRoute, redirect } from "@tanstack/solid-router";

export const Route = createFileRoute("/about")({
	beforeLoad: () => {
		throw redirect({ to: "/app/about" });
	},
});

import type { JSX } from "solid-js";

export type AppNavItem = {
	href: string;
	label: string;
	icon?: () => JSX.Element;
	active?: boolean;
	exact?: boolean;
};

import type { AppNavItem } from "@nomnomvault/app-shell";
import { InfoIcon, RecipesIcon } from "@nomnomvault/ui";

export function buildRecipeWorkspaceNav(pathname: string): {
	items: AppNavItem[];
	secondaryItems: AppNavItem[];
} {
	return {
		items: [
			{
				href: "/app/recipes",
				label: "Recipes",
				icon: () => <RecipesIcon size="sm" />,
				active: pathname.startsWith("/app/recipes"),
			},
		],
		secondaryItems: [
			{
				href: "/app/about",
				label: "About",
				icon: () => <InfoIcon size="sm" />,
				active: pathname === "/app/about",
			},
		],
	};
}

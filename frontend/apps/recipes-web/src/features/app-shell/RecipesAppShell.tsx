import type { AuthSessionResponse } from "@nomnomvault/api-client";
import {
	AppHeader,
	AppShell,
	AppSidebar,
	ThemeToggle,
} from "@nomnomvault/app-shell";
import { useLocation } from "@tanstack/solid-router";
import type { JSX } from "solid-js";
import { buildRecipeWorkspaceNav } from "../recipes/config/recipe-nav";

type RecipesAppShellProps = {
	session?: AuthSessionResponse;
	children: JSX.Element;
};

export function RecipesAppShell(props: RecipesAppShellProps) {
	const location = useLocation();
	const nav = () => buildRecipeWorkspaceNav(location().pathname);

	return (
		<AppShell
			header={
				<AppHeader
					brandHref="/"
					brandLabel="NomNomVault"
					items={nav().items}
					householdName={props.session?.activeHousehold?.name}
					secondaryItems={nav().secondaryItems}
					accountName={props.session?.user?.displayName}
					accountEmail={props.session?.user?.email}
					actions={<ThemeToggle />}
				/>
			}
			sidebar={
				<AppSidebar
					title="Workspace"
					items={nav().items}
					secondaryItems={nav().secondaryItems}
				/>
			}
		>
			{props.children}
		</AppShell>
	);
}

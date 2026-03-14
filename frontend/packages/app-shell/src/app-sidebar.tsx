import { For, Show } from "solid-js";

import { InfoIcon } from "@nomnomvault/ui";
import type { AppNavItem } from "./types";

type AppSidebarProps = {
	title: string;
	eyebrow?: string;
	items: AppNavItem[];
	secondaryItems?: AppNavItem[];
};

export function AppSidebar(props: AppSidebarProps) {
	return (
		<aside class="hidden gap-5 lg:grid lg:sticky lg:top-[5.75rem] lg:py-3">
			<div class="grid gap-1">
				<Show when={props.eyebrow}>
					<p class="nnv-eyebrow text-[0.72rem] tracking-[0.12em] text-[var(--nnv-text-muted)]">
						{props.eyebrow}
					</p>
				</Show>
				<h2 class="m-0 text-[1.2rem]">{props.title}</h2>
			</div>

			<nav class="grid gap-1.5" aria-label="Workspace navigation">
				<For each={props.items}>
					{(item) => (
						<a
							href={item.href}
							class="inline-flex min-h-[2.6rem] items-center gap-3 rounded-[var(--nnv-radius-sm)] border border-transparent px-4 py-2 font-bold text-[var(--nnv-text-muted)] no-underline transition-colors hover:bg-[var(--nnv-surface-2)] hover:text-[var(--nnv-text-strong)] data-[active=true]:border-[color:var(--nnv-line)] data-[active=true]:bg-[color:color-mix(in_oklab,var(--nnv-chip-active)_78%,var(--nnv-surface-1))] data-[active=true]:text-[var(--nnv-text-strong)]"
							aria-current={item.active ? "page" : undefined}
							data-active={item.active ? "true" : "false"}
						>
							{item.icon?.()}
							<span>{item.label}</span>
						</a>
					)}
				</For>
			</nav>

			<div class="grid gap-1.5">
				<For
					each={
						props.secondaryItems ?? [
							{
								href: "/app/about",
								label: "About",
								icon: () => <InfoIcon size="sm" />,
							},
						]
					}
				>
					{(item) => (
						<a
							href={item.href}
							class="inline-flex min-h-[2.6rem] items-center gap-3 rounded-[var(--nnv-radius-sm)] border border-transparent px-4 py-2 font-bold text-[var(--nnv-text-muted)] no-underline transition-colors hover:bg-[var(--nnv-surface-2)] hover:text-[var(--nnv-text-strong)] data-[active=true]:border-[color:var(--nnv-line)] data-[active=true]:bg-[color:color-mix(in_oklab,var(--nnv-chip-active)_78%,var(--nnv-surface-1))] data-[active=true]:text-[var(--nnv-text-strong)]"
							aria-current={item.active ? "page" : undefined}
							data-active={item.active ? "true" : "false"}
						>
							{item.icon?.()}
							<span>{item.label}</span>
						</a>
					)}
				</For>
			</div>
		</aside>
	);
}

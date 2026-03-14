import { For, Show } from "solid-js";
import type { AppNavItem } from "./types";

type AppMobileNavProps = {
  items: AppNavItem[]
  secondaryItems?: AppNavItem[]
}

export function AppMobileNav(props: AppMobileNavProps) {
  return (
    <div class="grid gap-4">
      <div class="grid gap-1.5">
        <For each={props.items}>
          {(item) => (
            <a
              href={item.href}
              class="inline-flex min-h-[2.6rem] items-center gap-3 rounded-[var(--nnv-radius-sm)] border border-transparent px-4 py-2 font-bold text-[var(--nnv-text-muted)] no-underline transition-colors hover:bg-[var(--nnv-surface-2)] hover:text-[var(--nnv-text-strong)] data-[active=true]:border-[color:var(--nnv-line)] data-[active=true]:bg-[color:color-mix(in_oklab,var(--nnv-chip-active)_78%,var(--nnv-surface-1))] data-[active=true]:text-[var(--nnv-text-strong)]"
              data-active={item.active ? 'true' : 'false'}
            >
              {item.icon?.()}
              {item.label}
            </a>
          )}
        </For>
      </div>
      <Show when={(props.secondaryItems?.length ?? 0) > 0}>
        <div class="grid gap-1.5">
          <For each={props.secondaryItems ?? []}>
            {(item) => (
              <a
                href={item.href}
                class="inline-flex min-h-[2.6rem] items-center gap-3 rounded-[var(--nnv-radius-sm)] border border-transparent px-4 py-2 font-bold text-[var(--nnv-text-muted)] no-underline transition-colors hover:bg-[var(--nnv-surface-2)] hover:text-[var(--nnv-text-strong)] data-[active=true]:border-[color:var(--nnv-line)] data-[active=true]:bg-[color:color-mix(in_oklab,var(--nnv-chip-active)_78%,var(--nnv-surface-1))] data-[active=true]:text-[var(--nnv-text-strong)]"
                data-active={item.active ? 'true' : 'false'}
              >
                {item.icon?.()}
                {item.label}
              </a>
            )}
          </For>
        </div>
      </Show>
    </div>
  )
}

import type { JSX } from "solid-js";
import { Show } from "solid-js";

import {
  Badge,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  HouseholdIcon,
  MenuIcon,
  ProfileIcon,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@nomnomvault/ui";

import { AppMobileNav } from './app-mobile-nav'
import type { AppNavItem } from "./types";

type AppHeaderProps = {
  brandHref: string
  brandLabel: string
  items: AppNavItem[]
  secondaryItems?: AppNavItem[]
  householdName?: string
  actions?: JSX.Element
  accountEmail?: string
  accountName?: string
  accountAvatarUrl?: string
}

export function AppHeader(props: AppHeaderProps) {
  return (
    <header class="sticky top-0 z-[60] border-b border-[color:var(--nnv-line)] bg-[color:color-mix(in_oklab,var(--nnv-surface-1)_92%,transparent)] backdrop-blur-[10px]">
      <div class="mx-auto flex w-[min(1440px,calc(100%-2rem))] items-center justify-between gap-4 py-2.5">
        <div class="flex items-center gap-3">
          <a
            href={props.brandHref}
            class="inline-flex items-center gap-2.5 no-underline"
          >
            <span class="h-3 w-3 rounded-full bg-[linear-gradient(135deg,var(--nnv-accent-strong),#9e8a63)] shadow-[0_0_0_0.3rem_rgba(133,112,77,0.16)]" />
            <span class="grid gap-0.5">
              <strong class="text-[0.98rem] tracking-[0.01em]">
                {props.brandLabel}
              </strong>
            </span>
          </a>
          <Show when={props.householdName}>
            <Badge
              tone="accent"
              class="w-fit border-[color:var(--nnv-line)] bg-transparent text-[var(--nnv-text-strong)]"
            >
              <HouseholdIcon size="sm" />
              <span>{props.householdName}</span>
            </Badge>
          </Show>
        </div>

        <div class="flex items-center gap-3">
          {props.actions}
          <DropdownMenu>
            <DropdownMenuTrigger
              class="inline-flex h-9.5 w-9.5 items-center justify-center rounded-full border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] p-0 text-[var(--nnv-text-strong)]"
              aria-label="Open account menu"
            >
              <Show
                when={props.accountAvatarUrl}
                fallback={
                  <span class="inline-flex h-[calc(100%-6px)] w-[calc(100%-6px)] items-center justify-center rounded-full bg-[color:color-mix(in_oklab,var(--nnv-surface-3)_72%,var(--nnv-surface-1))]">
                    <ProfileIcon size="md" />
                  </span>
                }
              >
                {(avatarUrl) => (
                  <img
                    src={avatarUrl()}
                    alt={props.accountName ?? props.accountEmail ?? "Account"}
                    class="block h-[calc(100%-6px)] w-[calc(100%-6px)] rounded-full object-cover"
                  />
                )}
              </Show>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem disabled>
                Signed in as{" "}
                {props.accountName ?? props.accountEmail ?? "NomNom User"}
              </DropdownMenuItem>
              <DropdownMenuItem>Log out</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <div class="inline-flex md:hidden">
            <Sheet>
              <SheetTrigger class="inline-flex h-[2.625rem] w-[2.625rem] items-center justify-center rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] p-0">
                <MenuIcon size="md" />
              </SheetTrigger>
              <SheetContent>
                <SheetHeader>
                  <SheetTitle>NomNomVault</SheetTitle>
                  <SheetDescription>
                    Navigate the workspace.
                  </SheetDescription>
                </SheetHeader>
                <AppMobileNav
                  items={props.items}
                  secondaryItems={props.secondaryItems}
                />
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </header>
  )
}

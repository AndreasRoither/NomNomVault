import {
  type ErrorComponentProps,
  HeadContent,
  Outlet,
  Scripts,
  createRootRouteWithContext,
} from '@tanstack/solid-router'
import { TanStackRouterDevtools } from '@tanstack/solid-router-devtools'

import { HydrationScript } from 'solid-js/web'
import { Suspense } from 'solid-js'

import type { AppRouterContext } from '../integrations/tanstack-query/provider'

import styleCss from '../styles.css?url'

export const Route = createRootRouteWithContext<AppRouterContext>()({
  head: () => ({
    links: [{ rel: 'stylesheet', href: styleCss }],
  }),
  shellComponent: RootComponent,
  errorComponent: RootErrorComponent,
})

function RootComponent() {
  return (
    <html lang="en">
      <head>
        <HydrationScript />
      </head>
      <body>
        <HeadContent />
        <Suspense>
          <Outlet />
          <TanStackRouterDevtools />
        </Suspense>
        <Scripts />
      </body>
    </html>
  )
}

function RootErrorComponent(props: ErrorComponentProps) {
  const errorMessage = import.meta.env.DEV
    ? props.error.message || String(props.error)
    : null

  return (
    <main class="page-wrap px-4 py-12">
      <section class="island-shell rounded-2xl p-6 sm:p-8">
        <p class="island-kicker mb-2">Application Error</p>
        <h1 class="display-title mb-3 text-4xl font-bold text-[var(--sea-ink)] sm:text-5xl">
          The page could not be rendered.
        </h1>
        <p class="m-0 max-w-3xl text-base leading-8 text-[var(--sea-ink-soft)]">
          Refresh the page or try again after the current change finishes
          loading.
        </p>
        {errorMessage ? (
          <pre class="mt-4 overflow-x-auto rounded-xl border border-[var(--line)] bg-[var(--header-bg)] p-4 text-sm text-[var(--sea-ink-soft)]">
            <code>{errorMessage}</code>
          </pre>
        ) : null}
      </section>
    </main>
  )
}

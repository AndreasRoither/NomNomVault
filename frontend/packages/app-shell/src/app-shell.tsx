import type { JSX } from 'solid-js'

import { ShellFrame } from '@nomnomvault/ui'

type AppShellProps = {
  header: JSX.Element
  sidebar?: JSX.Element
  children: JSX.Element
}

export function AppShell(props: AppShellProps) {
  return (
    <ShellFrame class="min-h-screen" header={props.header}>
      <div class="mx-auto min-h-0 w-[min(1440px,calc(100%-2rem))] py-3 pb-9">
        <div class="grid items-start gap-7 lg:grid-cols-[15rem_minmax(0,1fr)]">
          {props.sidebar}
          <main class="min-w-0">{props.children}</main>
        </div>
      </div>
    </ShellFrame>
  )
}

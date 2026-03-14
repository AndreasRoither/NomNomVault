import type { JSX } from 'solid-js'

import { cn } from './utils'

type ShellFrameProps = {
  class?: string
  header?: JSX.Element
  nav?: JSX.Element
  aside?: JSX.Element
  footer?: JSX.Element
  children: JSX.Element
}

export function ShellFrame(props: ShellFrameProps) {
  return (
    <div class={cn('grid min-h-screen grid-rows-[auto_minmax(0,1fr)_auto]', props.class)}>
      {props.header}
      <div class="min-h-0">
        {props.nav}
        {props.children}
        {props.aside}
      </div>
      {props.footer}
    </div>
  )
}

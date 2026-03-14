import type { JSX } from 'solid-js'
import { splitProps } from 'solid-js'

import { cn } from './utils'

export function TextField(props: JSX.HTMLAttributes<HTMLDivElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <div
      class={cn(
        'flex min-h-12 items-center gap-3 rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-1)] px-4 shadow-[var(--nnv-shadow-md)]',
        local.class,
      )}
      {...rest}
    />
  )
}

export function TextFieldInput(
  props: JSX.InputHTMLAttributes<HTMLInputElement>,
) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <input
      class={cn(
        'w-full border-0 bg-transparent text-[var(--nnv-text-strong)] outline-none placeholder:text-[var(--nnv-text-muted)] focus:outline-none focus-visible:outline-none',
        local.class,
      )}
      {...rest}
    />
  )
}

import type { JSX } from 'solid-js'
import { splitProps } from 'solid-js'

import { cn } from './utils'

type BadgeProps = JSX.HTMLAttributes<HTMLSpanElement> & {
  tone?: 'default' | 'accent' | 'warning' | 'danger' | 'safe'
}

export function Badge(props: BadgeProps) {
  const [local, rest] = splitProps(props, ['class', 'tone'])

  const toneClass = {
    default:
      'border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] text-[var(--nnv-text-strong)]',
    accent:
      'border-[color:var(--nnv-line)] bg-[color:color-mix(in_oklab,var(--nnv-chip-active)_72%,var(--nnv-surface-1))] text-[var(--nnv-text-strong)]',
    warning:
      'border-transparent bg-[var(--nnv-warning-soft)] text-[var(--nnv-text-strong)]',
    danger:
      'border-transparent bg-[var(--nnv-danger-soft)] text-[var(--nnv-text-strong)]',
    safe:
      'border-transparent bg-[var(--nnv-safe-soft)] text-[var(--nnv-text-strong)]',
  }[local.tone ?? 'default']

  return (
    <span
      class={cn(
        'inline-flex min-h-9 items-center gap-1.5 rounded-[var(--nnv-radius-sm)] border px-3 py-1.5 text-sm font-semibold',
        toneClass,
        local.class,
      )}
      {...rest}
    />
  )
}

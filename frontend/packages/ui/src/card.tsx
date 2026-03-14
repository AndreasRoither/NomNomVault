import type { JSX } from 'solid-js'
import { splitProps } from 'solid-js'

import { cn } from './utils'

export function Card(props: JSX.HTMLAttributes<HTMLDivElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <div
      class={cn(
        'rounded-[var(--nnv-radius-lg)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-1)]',
        local.class,
      )}
      {...rest}
    />
  )
}

export function CardHeader(props: JSX.HTMLAttributes<HTMLDivElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return <div class={cn('grid gap-2 px-5 pt-5', local.class)} {...rest} />
}

export function CardTitle(props: JSX.HTMLAttributes<HTMLHeadingElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <h3
      class={cn(
        'text-[1.55rem] font-semibold tracking-[-0.02em] text-[var(--nnv-text-strong)]',
        local.class,
      )}
      {...rest}
    />
  )
}

export function CardContent(props: JSX.HTMLAttributes<HTMLDivElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return <div class={cn('px-5 pb-5', local.class)} {...rest} />
}

export function CardFooter(props: JSX.HTMLAttributes<HTMLDivElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <div
      class={cn(
        'flex flex-wrap items-center gap-3 px-5 pb-5 text-sm text-[var(--nnv-text-muted)]',
        local.class,
      )}
      {...rest}
    />
  )
}

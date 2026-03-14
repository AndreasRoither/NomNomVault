import type { JSX } from 'solid-js'
import { splitProps } from 'solid-js'

import { cn } from './utils'

export function Breadcrumb(props: JSX.HTMLAttributes<HTMLElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return <nav aria-label="Breadcrumbs" class={cn(local.class)} {...rest} />
}

export function BreadcrumbList(props: JSX.HTMLAttributes<HTMLOListElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <ol
      class={cn(
        'm-0 flex flex-wrap items-center gap-2 p-0 text-sm text-[var(--nnv-text-muted)]',
        local.class,
      )}
      {...rest}
    />
  )
}

export function BreadcrumbItem(props: JSX.HTMLAttributes<HTMLLIElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return <li class={cn('inline-flex items-center gap-2', local.class)} {...rest} />
}

export function BreadcrumbLink(props: JSX.AnchorHTMLAttributes<HTMLAnchorElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return (
    <a
      class={cn('transition-colors hover:text-[var(--nnv-text-strong)]', local.class)}
      {...rest}
    />
  )
}

export function BreadcrumbSeparator(props: JSX.HTMLAttributes<HTMLSpanElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return <span class={cn('text-[var(--nnv-text-muted)]', local.class)} {...rest} />
}

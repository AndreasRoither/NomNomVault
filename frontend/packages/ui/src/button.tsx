import type { JSX } from 'solid-js'
import { splitProps } from 'solid-js'

import { cn } from './utils'

type ButtonProps = JSX.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'quiet' | 'chip' | 'danger'
  size?: 'sm' | 'md' | 'lg' | 'icon'
}

export function Button(props: ButtonProps) {
  const [local, rest] = splitProps(props, ['class', 'variant', 'size'])

  const variantClass = {
    primary:
      'border-[color:var(--nnv-accent-strong)] bg-[var(--nnv-accent-strong)] text-[#fffaf0] hover:bg-[color:color-mix(in_oklab,var(--nnv-accent-strong)_92%,black)]',
    secondary:
      'border-[color:var(--nnv-line)] bg-[var(--nnv-surface-2)] text-[var(--nnv-text-strong)] hover:bg-[var(--nnv-surface-3)]',
    quiet:
      'border-transparent bg-transparent text-[var(--nnv-text-strong)] hover:bg-[var(--nnv-surface-2)]',
    chip:
      'rounded-[var(--nnv-radius-sm)] border-[color:var(--nnv-line)] bg-[color:color-mix(in_oklab,var(--nnv-surface-1)_78%,var(--nnv-surface-3))] text-[var(--nnv-text-strong)] hover:bg-[var(--nnv-chip-active)] data-[selected=true]:bg-[var(--nnv-chip-active)]',
    danger:
      'border-[#8f4e45] bg-[#8f4e45] text-[#fff7f5] hover:bg-[#7d433c]',
  }[local.variant ?? 'primary']

  const sizeClass = {
    sm: 'min-h-9 px-3 py-1.5 text-sm',
    md: 'min-h-[2.625rem] px-3.5 py-2',
    lg: 'min-h-[2.9rem] px-[1.15rem] py-3',
    icon: 'h-[2.625rem] w-[2.625rem] p-0',
  }[local.size ?? 'md']

  return (
    <button
      type="button"
      class={cn(
        'inline-flex items-center justify-center gap-2 rounded-[var(--nnv-radius-md)] border font-bold transition-colors transition-transform duration-150 ease-out focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--nnv-accent-strong)] disabled:cursor-not-allowed disabled:opacity-55 disabled:transform-none hover:-translate-y-px',
        variantClass,
        sizeClass,
        local.class,
      )}
      {...rest}
    />
  )
}

import type { JSX } from 'solid-js'
import { createContext, useContext, splitProps } from 'solid-js'

import { cn } from './utils'

type SelectContextValue = {
  value?: () => string | undefined
  onValueChange?: (value: string) => void
}

const SelectContext = createContext<SelectContextValue>()

type SelectProps = JSX.HTMLAttributes<HTMLDivElement> & {
  value?: string
  onValueChange?: (value: string) => void
}

export function Select(props: SelectProps) {
  const [local, rest] = splitProps(props, [
    'class',
    'children',
    'value',
    'onValueChange',
  ])

  return (
    <SelectContext.Provider
      value={{
        value: () => local.value,
        onValueChange: local.onValueChange,
      }}
    >
      <div class={cn('nnv-select', local.class)} {...rest}>
        {local.children}
      </div>
    </SelectContext.Provider>
  )
}

export function SelectTrigger(
  props: JSX.SelectHTMLAttributes<HTMLSelectElement>,
) {
  const context = useContext(SelectContext)
  const [local, rest] = splitProps(props, ['class', 'children'])

  return (
    <select
      class={cn(
        'min-h-[2.625rem] w-full rounded-[var(--nnv-radius-md)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-1)] px-4 text-[var(--nnv-text-strong)] shadow-[var(--nnv-shadow-md)] outline-none focus:outline-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--nnv-accent-strong)]',
        local.class,
      )}
      value={context?.value?.()}
      onInput={(event) => context?.onValueChange?.(event.currentTarget.value)}
      {...rest}
    >
      {local.children}
    </select>
  )
}

export function SelectItem(props: { value: string; children: JSX.Element }) {
  return <option value={props.value}>{props.children}</option>
}

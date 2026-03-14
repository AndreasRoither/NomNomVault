import type { JSX } from 'solid-js'
import {
  createContext,
  createUniqueId,
  createSignal,
  splitProps,
  useContext,
} from 'solid-js'
import { Show } from "solid-js";

import { createDismissibleLayer } from "./internal/create-dismissible-layer";
import { createEscapeDismiss } from "./internal/create-escape-dismiss";
import { createFocusRestore } from "./internal/create-focus-restore";
import { cn } from './utils'

type DropdownMenuContextValue = {
  open: () => boolean
  setOpen: (value: boolean) => void
  triggerRef: () => HTMLButtonElement | undefined
  setTriggerRef: (element: HTMLButtonElement | undefined) => void
  contentRef: () => HTMLDivElement | undefined
  setContentRef: (element: HTMLDivElement | undefined) => void
  contentId: () => string
}

const DropdownMenuContext = createContext<DropdownMenuContextValue>()

export function DropdownMenu(props: { children: JSX.Element; class?: string }) {
  const [open, setOpen] = createSignal(false)
  const [triggerRef, setTriggerRef] = createSignal<HTMLButtonElement>()
  const [contentRef, setContentRef] = createSignal<HTMLDivElement>()
  const contentId = createUniqueId()

  createDismissibleLayer({
    isOpen: open,
    refs: () => [triggerRef(), contentRef()],
    onDismiss: () => setOpen(false),
    preventOutsideInteraction: true,
  })

  createEscapeDismiss({
    isOpen: open,
    onDismiss: () => setOpen(false),
  })

  createFocusRestore({
    isOpen: open,
    trigger: triggerRef,
    content: contentRef,
    initialFocusSelector: '[role="menuitem"]:not([disabled])',
  })

  return (
    <DropdownMenuContext.Provider
      value={{
        open,
        setOpen,
        triggerRef,
        setTriggerRef,
        contentRef,
        setContentRef,
        contentId: () => contentId,
      }}
    >
      <div class={cn('relative', props.class)}>{props.children}</div>
    </DropdownMenuContext.Provider>
  )
}

export function DropdownMenuTrigger(
  props: JSX.ButtonHTMLAttributes<HTMLButtonElement>,
) {
  const context = useContext(DropdownMenuContext)
  const [local, rest] = splitProps(props, ['class', 'onClick', 'ref'])

  return (
    <button
      ref={(element) => {
        context?.setTriggerRef(element)
        const ref = local.ref
        if (typeof ref === 'function') {
          ref(element)
        }
      }}
      type="button"
      aria-expanded={context?.open() ?? false}
      aria-haspopup="menu"
      aria-controls={context?.contentId()}
      class={cn('inline-flex items-center justify-center', local.class)}
      {...rest}
      onClick={(event) => {
        const onClick = local.onClick as
          | JSX.EventHandler<HTMLButtonElement, MouseEvent>
          | undefined
        onClick?.(event)
        context?.setOpen(!context.open())
      }}
    />
  )
}

export function DropdownMenuContent(
  props: JSX.HTMLAttributes<HTMLDivElement>,
) {
  const context = useContext(DropdownMenuContext)
  const [local, rest] = splitProps(props, ['class', 'children', 'ref'])

  const focusItem = (direction: 1 | -1 | 0) => {
    const items = Array.from(
      context?.contentRef()?.querySelectorAll<HTMLElement>(
        '[role="menuitem"]:not([disabled])',
      ) ?? [],
    )

    if (items.length === 0) {
      return
    }

    const activeIndex = items.findIndex((item) => item === document.activeElement)

    if (direction === 0 || activeIndex === -1) {
      items[0]?.focus()
      return
    }

    const nextIndex = (activeIndex + direction + items.length) % items.length
    items[nextIndex]?.focus()
  }

  return (
    <Show when={context?.open()}>
      <div
        ref={(element) => {
          context?.setContentRef(element)
          const ref = local.ref
          if (typeof ref === 'function') {
            ref(element)
          }
        }}
        id={context?.contentId()}
        class={cn(
          'absolute right-0 top-[calc(100%+0.5rem)] z-[70] min-w-56 rounded-[var(--nnv-radius-lg)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-solid)] p-2 shadow-[var(--nnv-shadow-lg)]',
          local.class,
        )}
        role="menu"
        tabindex="-1"
        onKeyDown={(event) => {
          if (event.key === 'ArrowDown') {
            event.preventDefault()
            focusItem(1)
          } else if (event.key === 'ArrowUp') {
            event.preventDefault()
            focusItem(-1)
          } else if (event.key === 'Home') {
            event.preventDefault()
            focusItem(0)
          } else if (event.key === 'End') {
            event.preventDefault()
            const items = Array.from(
              context?.contentRef()?.querySelectorAll<HTMLElement>(
                '[role="menuitem"]:not([disabled])',
              ) ?? [],
            )
            items.at(-1)?.focus()
          } else if (event.key === 'Tab') {
            context?.setOpen(false)
          }
        }}
        {...rest}
      >
        {local.children}
      </div>
    </Show>
  )
}

type DropdownMenuItemProps = JSX.ButtonHTMLAttributes<HTMLButtonElement> & {
  inset?: boolean
}

export function DropdownMenuItem(props: DropdownMenuItemProps) {
  const context = useContext(DropdownMenuContext)
  const [local, rest] = splitProps(props, ['class', 'inset', 'onClick'])

  return (
    <button
      type="button"
      role="menuitem"
      class={cn(
        'flex w-full items-center rounded-[var(--nnv-radius-sm)] border-0 bg-transparent px-3.5 py-2.5 text-left text-[var(--nnv-text-strong)] transition-colors hover:bg-[var(--nnv-surface-2)] focus:bg-[var(--nnv-surface-2)] focus:outline-none disabled:cursor-not-allowed disabled:opacity-55',
        local.inset && 'pl-[1.15rem]',
        local.class,
      )}
      {...rest}
      onClick={(event) => {
        if (event.currentTarget.disabled) {
          return
        }
        const onClick = local.onClick as
          | JSX.EventHandler<HTMLButtonElement, MouseEvent>
          | undefined
        onClick?.(event)
        context?.setOpen(false)
      }}
    />
  )
}

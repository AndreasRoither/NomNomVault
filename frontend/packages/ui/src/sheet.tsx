import type { JSX } from 'solid-js'
import {
  Show,
  createUniqueId,
  createContext,
  createSignal,
  splitProps,
  useContext,
} from 'solid-js'
import { Portal } from 'solid-js/web'

import { createBodyScrollLock } from "./internal/create-body-scroll-lock";
import { createEscapeDismiss } from "./internal/create-escape-dismiss";
import { createFocusTrap } from "./internal/create-focus-trap";
import { createFocusRestore } from "./internal/create-focus-restore";
import { cn } from './utils'

type SheetContextValue = {
  open: () => boolean
  setOpen: (value: boolean) => void
  triggerRef: () => HTMLButtonElement | undefined
  setTriggerRef: (value: HTMLButtonElement | undefined) => void
  contentRef: () => HTMLDivElement | undefined
  setContentRef: (value: HTMLDivElement | undefined) => void
  titleId: () => string | undefined
  setTitleId: (value: string | undefined) => void
  descriptionId: () => string | undefined
  setDescriptionId: (value: string | undefined) => void
}

const SheetContext = createContext<SheetContextValue>()

export function Sheet(props: { children: JSX.Element }) {
  const [open, setOpen] = createSignal(false)
  const [triggerRef, setTriggerRef] = createSignal<HTMLButtonElement>()
  const [contentRef, setContentRef] = createSignal<HTMLDivElement>()
  const [titleId, setTitleId] = createSignal<string>()
  const [descriptionId, setDescriptionId] = createSignal<string>()

  createBodyScrollLock(open)
  createEscapeDismiss({
    isOpen: open,
    onDismiss: () => setOpen(false),
  })
  createFocusRestore({
    isOpen: open,
    trigger: triggerRef,
    content: contentRef,
  })
  createFocusTrap({
    isOpen: open,
    content: contentRef,
  })

  return (
    <SheetContext.Provider
      value={{
        open,
        setOpen,
        triggerRef,
        setTriggerRef,
        contentRef,
        setContentRef,
        titleId,
        setTitleId,
        descriptionId,
        setDescriptionId,
      }}
    >
      {props.children}
    </SheetContext.Provider>
  )
}

export function SheetTrigger(
  props: JSX.ButtonHTMLAttributes<HTMLButtonElement>,
) {
  const context = useContext(SheetContext)
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
      aria-haspopup="dialog"
      class={cn(local.class)}
      onClick={(event) => {
        const onClick = local.onClick as
          | JSX.EventHandler<HTMLButtonElement, MouseEvent>
          | undefined
        onClick?.(event)
        context?.setOpen(true)
      }}
      {...rest}
    />
  )
}

type SheetContentProps = JSX.HTMLAttributes<HTMLDivElement> & {
  side?: 'left' | 'right'
}

export function SheetContent(props: SheetContentProps) {
  const context = useContext(SheetContext)
  const [local, rest] = splitProps(props, ['class', 'children', 'side', 'ref'])

  return (
    <Show when={context?.open()}>
      <Portal>
        <div
          class="fixed inset-0 z-[65] bg-[rgba(17,16,13,0.42)] backdrop-blur-[3px]"
          onClick={() => context?.setOpen(false)}
          role="presentation"
        />
        <div
          ref={(element) => {
            context?.setContentRef(element)
            const ref = local.ref
            if (typeof ref === 'function') {
              ref(element)
            }
          }}
          class={cn(
            'fixed inset-y-0 z-[70] grid w-[min(28rem,calc(100vw-1rem))] gap-4 border-[color:var(--nnv-line)] bg-[var(--nnv-surface-solid)] p-5 text-[var(--nnv-text-strong)] shadow-[var(--nnv-shadow-lg)]',
            local.side === 'left'
              ? 'left-0 border-r'
              : 'right-0 border-l',
            local.class,
          )}
          role="dialog"
          aria-modal="true"
          aria-labelledby={context?.titleId()}
          aria-describedby={context?.descriptionId()}
          tabindex="-1"
          onClick={(event) => event.stopPropagation()}
          {...rest}
        >
          {local.children}
        </div>
      </Portal>
    </Show>
  )
}

export function SheetHeader(props: JSX.HTMLAttributes<HTMLDivElement>) {
  const [local, rest] = splitProps(props, ['class'])

  return <div class={cn('grid gap-2', local.class)} {...rest} />
}

export function SheetTitle(props: JSX.HTMLAttributes<HTMLHeadingElement>) {
  const context = useContext(SheetContext)
  const generatedId = createUniqueId()
  const [local, rest] = splitProps(props, ['class', 'id'])

  context?.setTitleId(local.id ?? generatedId)

  return (
    <h2
      id={local.id ?? generatedId}
      class={cn('text-xl font-semibold tracking-[-0.02em]', local.class)}
      {...rest}
    />
  )
}

export function SheetDescription(props: JSX.HTMLAttributes<HTMLParagraphElement>) {
  const context = useContext(SheetContext)
  const generatedId = createUniqueId()
  const [local, rest] = splitProps(props, ['class', 'id'])

  context?.setDescriptionId(local.id ?? generatedId)

  return (
    <p
      id={local.id ?? generatedId}
      class={cn('m-0 text-sm leading-6 text-[var(--nnv-text-muted)]', local.class)}
      {...rest}
    />
  )
}

import { createEffect, onCleanup, type Accessor } from "solid-js";

type CreateFocusTrapOptions = {
	isOpen: Accessor<boolean>;
	content: Accessor<HTMLElement | undefined>;
};

function getFocusableElements(content: HTMLElement | undefined) {
	if (!content) {
		return [];
	}

	return Array.from(
		content.querySelectorAll<HTMLElement>(
			'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
		),
	).filter((element) => !element.hasAttribute("disabled"));
}

export function createFocusTrap(options: CreateFocusTrapOptions) {
	createEffect(() => {
		if (!options.isOpen() || typeof document === "undefined") {
			return;
		}

		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key !== "Tab") {
				return;
			}

			const content = options.content();
			const focusable = getFocusableElements(content);

			if (!content || focusable.length === 0) {
				event.preventDefault();
				content?.focus();
				return;
			}

			const first = focusable[0];
			const last = focusable[focusable.length - 1];
			const active =
				document.activeElement instanceof HTMLElement
					? document.activeElement
					: undefined;

			if (!active || !content.contains(active)) {
				event.preventDefault();
				(event.shiftKey ? last : first)?.focus();
				return;
			}

			if (!event.shiftKey && active === last) {
				event.preventDefault();
				first?.focus();
			}

			if (event.shiftKey && active === first) {
				event.preventDefault();
				last?.focus();
			}
		};

		document.addEventListener("keydown", handleKeyDown);

		onCleanup(() => {
			document.removeEventListener("keydown", handleKeyDown);
		});
	});
}

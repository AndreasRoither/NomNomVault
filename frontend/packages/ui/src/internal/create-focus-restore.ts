import { createEffect, onCleanup, type Accessor } from "solid-js";

type CreateFocusRestoreOptions = {
	isOpen: Accessor<boolean>;
	trigger: Accessor<HTMLElement | undefined>;
	content: Accessor<HTMLElement | undefined>;
	initialFocusSelector?: string;
};

function findInitialFocusTarget(
	content: HTMLElement | undefined,
	selector?: string,
) {
	if (!content) {
		return undefined;
	}

	if (selector) {
		return content.querySelector<HTMLElement>(selector) ?? content;
	}

	return (
		content.querySelector<HTMLElement>(
			'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
		) ?? content
	);
}

export function createFocusRestore(options: CreateFocusRestoreOptions) {
	createEffect(() => {
		if (!options.isOpen() || typeof document === "undefined") {
			return;
		}

		const previousActive =
			document.activeElement instanceof HTMLElement
				? document.activeElement
				: undefined;

		queueMicrotask(() => {
			findInitialFocusTarget(
				options.content(),
				options.initialFocusSelector,
			)?.focus();
		});

		onCleanup(() => {
			queueMicrotask(() => {
				(options.trigger() ?? previousActive)?.focus?.();
			});
		});
	});
}

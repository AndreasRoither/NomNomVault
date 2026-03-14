import { createEffect, onCleanup, type Accessor } from "solid-js";

type CreateDismissibleLayerOptions = {
	isOpen: Accessor<boolean>;
	refs: Accessor<(HTMLElement | undefined)[]>;
	onDismiss: () => void;
	preventOutsideInteraction?: boolean;
};

function isTargetInside(
	target: EventTarget | null,
	elements: (HTMLElement | undefined)[],
) {
	if (!(target instanceof Node)) {
		return false;
	}

	return elements.some((element) => Boolean(element?.contains(target)));
}

export function createDismissibleLayer(
	options: CreateDismissibleLayerOptions,
) {
	createEffect(() => {
		if (!options.isOpen() || typeof document === "undefined") {
			return;
		}

		const handlePointerDown = (event: PointerEvent) => {
			const elements = options.refs().filter(
				(element): element is HTMLElement => Boolean(element),
			);

			if (isTargetInside(event.target, elements)) {
				return;
			}

			if (options.preventOutsideInteraction) {
				event.preventDefault();
				event.stopPropagation();
			}

			options.onDismiss();
		};

		document.addEventListener("pointerdown", handlePointerDown, true);

		onCleanup(() => {
			document.removeEventListener("pointerdown", handlePointerDown, true);
		});
	});
}

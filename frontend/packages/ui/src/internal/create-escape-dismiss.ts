import { createEffect, onCleanup, type Accessor } from "solid-js";

type CreateEscapeDismissOptions = {
	isOpen: Accessor<boolean>;
	onDismiss: () => void;
};

export function createEscapeDismiss(options: CreateEscapeDismissOptions) {
	createEffect(() => {
		if (!options.isOpen() || typeof window === "undefined") {
			return;
		}

		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				options.onDismiss();
			}
		};

		window.addEventListener("keydown", handleKeyDown);

		onCleanup(() => {
			window.removeEventListener("keydown", handleKeyDown);
		});
	});
}

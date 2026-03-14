import { createEffect, onCleanup, type Accessor } from "solid-js";

export function createBodyScrollLock(isOpen: Accessor<boolean>) {
	createEffect(() => {
		if (!isOpen() || typeof document === "undefined") {
			return;
		}

		const previous = document.body.style.overflow;
		document.body.style.overflow = "hidden";

		onCleanup(() => {
			document.body.style.overflow = previous;
		});
	});
}

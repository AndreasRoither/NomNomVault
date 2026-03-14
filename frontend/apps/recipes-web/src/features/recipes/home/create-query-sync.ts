import { createEffect, createSignal, onCleanup } from "solid-js";

type CreateQuerySyncOptions = {
	value: () => string;
	onCommit: (value: string) => void;
	delayMs?: number;
};

export function createQuerySync(options: CreateQuerySyncOptions) {
	const [query, setQuery] = createSignal(options.value());

	createEffect(() => {
		setQuery(options.value());
	});

	createEffect(() => {
		if (typeof window === "undefined") {
			return;
		}

		if (query() === options.value()) {
			return;
		}

		const handle = window.setTimeout(() => {
			options.onCommit(query());
		}, options.delayMs ?? 250);

		onCleanup(() => window.clearTimeout(handle));
	});

	return {
		query,
		setQuery,
	};
}

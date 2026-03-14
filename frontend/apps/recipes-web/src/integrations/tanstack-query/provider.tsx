import type { ApiClient } from "@nomnomvault/api-client";
import { QueryClient } from "@tanstack/solid-query";
import { apiClient } from "../api-client";

export type AppRouterContext = {
	queryClient: QueryClient;
	apiClient: ApiClient;
};

export function getContext(): AppRouterContext {
	const queryClient = new QueryClient();
	return {
		queryClient,
		apiClient,
	};
}

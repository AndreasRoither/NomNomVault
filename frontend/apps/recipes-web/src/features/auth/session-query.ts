import {
	type ApiClient,
	type AuthSessionResponse,
	apiQueryKeys,
} from "@nomnomvault/api-client";

export function sessionQueryOptions(api: ApiClient) {
	return {
		queryKey: apiQueryKeys.auth.session(),
		staleTime: 60_000,
		queryFn: async (): Promise<AuthSessionResponse> => api.auth.session(),
	};
}

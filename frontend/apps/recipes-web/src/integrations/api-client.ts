import { createApiClient } from "@nomnomvault/api-client";
import { env } from "../env";

const localApiBaseUrl = "http://127.0.0.1:8080/api/v1";
const composeApiBaseUrl = "http://api.localhost/api/v1";

function resolveApiBaseUrlForHost(hostname: string) {
	if (hostname === "localhost" || hostname === "127.0.0.1") {
		return localApiBaseUrl;
	}

	if (
		hostname === "recipes.localhost" ||
		hostname === "grocery.localhost" ||
		hostname === "api.localhost"
	) {
		return composeApiBaseUrl;
	}

	return "/api/v1";
}

function resolveBrowserBaseUrl() {
	if (env.VITE_API_BASE_URL) {
		return env.VITE_API_BASE_URL;
	}

	if (typeof window === "undefined") {
		return "/api/v1";
	}

	return resolveApiBaseUrlForHost(window.location.hostname);
}

async function resolveServerRelativeUrl(path: string) {
	if (env.VITE_API_BASE_URL) {
		return new URL(path, env.VITE_API_BASE_URL).toString();
	}

	if (env.SERVER_URL) {
		const serverUrl = new URL(env.SERVER_URL);
		const baseUrl = resolveApiBaseUrlForHost(serverUrl.hostname);

		if (baseUrl.startsWith("http://") || baseUrl.startsWith("https://")) {
			return new URL(path, baseUrl).toString();
		}

		return new URL(path, serverUrl.origin).toString();
	}

	return new URL(path, localApiBaseUrl).toString();
}

const ssrAwareFetch: typeof fetch = async (input, init) => {
	if (
		!import.meta.env.SSR ||
		typeof input !== "string" ||
		!input.startsWith("/")
	) {
		return fetch(input, init);
	}

	return fetch(await resolveServerRelativeUrl(input), init);
};

export const apiClient = createApiClient({
	baseUrl: resolveBrowserBaseUrl(),
	fetch: ssrAwareFetch,
});

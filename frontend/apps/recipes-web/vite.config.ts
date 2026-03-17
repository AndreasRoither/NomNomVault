import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/solid-start/plugin/vite";
import { nitro } from "nitro/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
	resolve: {
		tsconfigPaths: true,
	},
	plugins: [tailwindcss(), tanstackStart(), nitro(), solidPlugin({ ssr: true })],
	server: {
		port: 3000,
		strictPort: true,
	},
	preview: {
		port: 3000,
		strictPort: true,
	},
	nitro: {},
});

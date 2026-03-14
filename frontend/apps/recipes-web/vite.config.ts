import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/solid-start/plugin/vite";
import { nitro } from "nitro/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
import viteTsConfigPaths from "vite-tsconfig-paths";

export default defineConfig({
	plugins: [
		// this is the plugin that enables path aliases
		viteTsConfigPaths({
			projects: ["./tsconfig.json"],
		}),
		tailwindcss(),
		tanstackStart(),
		nitro(),
		solidPlugin({ ssr: true }),
	],
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

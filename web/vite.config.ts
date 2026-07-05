import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    // Forward /api to the Go server so the auth cookie is same-origin.
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});

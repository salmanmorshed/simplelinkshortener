import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
    plugins: [vue(), viteSingleFile()],
    build: {
        outDir: "../internal/routes/webroot",
        emptyOutDir: true,
        rollupOptions: {
            input: "./web.html",
        },
    },
    server: {
        host: "127.0.0.1",
    },
});

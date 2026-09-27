import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { readFileSync } from "node:fs";
export default defineConfig({
  plugins: [
    react(),
    {
      name: "licenses",
      generateBundle() {
        const files = [
          "LICENSE",
          "node_modules/react/LICENSE",
          "node_modules/react-dom/LICENSE",
          "node_modules/scheduler/LICENSE",
          "node_modules/lucide-react/LICENSE",
        ];
        this.emitFile({
          type: "asset",
          fileName: "licenses.txt",
          source: files
            .map((file) => `${file}\n${readFileSync(file, "utf8")}`)
            .join("\n\n"),
        });
      },
    },
  ],
  define: { "process.env.NODE_ENV": JSON.stringify("production") },
  build: {
    outDir: "../web/sayso",
    emptyOutDir: true,
    minify: true,
    lib: {
      entry: "src/main.tsx",
      formats: ["es"],
      fileName: () => "sayso.js",
      cssFileName: "sayso",
    },
  },
});

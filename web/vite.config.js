import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_API_PROXY_TARGET || "http://localhost:8082";
const waWebTarget = process.env.VITE_WA_WEB_PROXY_TARGET || "http://localhost:3001";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: apiTarget,
        changeOrigin: true,
      },
      "/webhook": {
        target: apiTarget,
        changeOrigin: true,
      },
      "/waweb": {
        target: waWebTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/waweb/, ""),
      },
    },
  },
});
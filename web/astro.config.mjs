import alpinejs from '@astrojs/alpinejs';
import { defineConfig } from 'astro/config';

export default defineConfig({
  output: 'static',
  // PocketBase serves the build from its pb_public directory.
  outDir: '../backend/pb_public',
  integrations: [alpinejs({ entrypoint: '/src/lib/alpine.ts' })],
  vite: {
    server: {
      // The client always talks to its own origin: in development Astro forwards the API to PocketBase.
      proxy: { '/api': 'http://127.0.0.1:8090' },
    },
  },
});

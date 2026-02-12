import path from 'path';
import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, '.', '');

    return {
      server: {
        port: 3500,
        host: '0.0.0.0',
        // Optional: Proxy API requests to avoid CORS in development
        // proxy: {
        //   '/api': {
        //     target: env.VITE_API_URL || 'http://localhost:8080',
        //     changeOrigin: true,
        //   },
        // },
      },
      plugins: [react()],
      resolve: {
        alias: {
          '@': path.resolve(__dirname, '.'),
        }
      },
      // Make env variables available in TypeScript
      define: {
        'import.meta.env.VITE_API_URL': JSON.stringify(env.VITE_API_URL || 'http://localhost:8080'),
        'import.meta.env.VITE_WS_URL': JSON.stringify(env.VITE_WS_URL || 'ws://localhost:8080/ws'),
      },
    };
});

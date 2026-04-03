import { defineConfig } from 'orval';

export default defineConfig({
  kraken: {
    input: '../internal/market/api.yaml',
    output: {
      mode: 'single',
      target: './src/client/api.ts',
      schemas: './src/client/models',
      client: 'axios',
      mock: false,
    },
  },
});

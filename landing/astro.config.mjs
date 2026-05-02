import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  site: 'https://mbiondo.github.io',
  base: '/a11ysentry',
  vite: {
    plugins: [tailwindcss()],
  },
});

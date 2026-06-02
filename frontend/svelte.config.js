import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  kit: {
    // Configure the static adapter for SSG
    adapter: adapter({
      pages: 'build',
      assets: 'build',
      fallback: 'index.html',
      precompress: false,
      strict: false
    }),
    
    // Set the base path. Must be empty for serving from root.
    paths: {
      base: '', // <-- This should resolve the CSS/JS routing issue.
    },

    // CSP: SvelteKit injects hashes for inline bootstrap scripts (see kit docs).
    // frame-ancestors belongs on HTTP response headers (reverse proxy), not <meta>.
    csp: {
      mode: 'hash',
      directives: {
        'default-src': ['self'],
        'script-src': ['self'],
        'style-src': ['self', 'unsafe-inline', 'https://fonts.googleapis.com'],
        'font-src': ['self', 'https://fonts.gstatic.com'],
        'img-src': ['self', 'blob:', 'data:', 'https:'],
        'connect-src': [
          'self',
          'http://localhost:8080',
          'http://127.0.0.1:8080',
          'https:',
        ],
        'base-uri': ['self'],
        'form-action': ['self'],
      },
    },

    // Prerendering configuration
    prerender: {
      entries: ['*'],
      handleUnseenRoutes: 'ignore',
      
      // Don't prerender dynamic routes like /vault/[vault_id]
      // These will be handled by client-side navigation
      handleHttpError: ({ status, path, referrer, referenceType }) => {
        if (status === 404) {
          console.warn(`Could not prerender ${path}`);
        }
      }
    }
  }
};

export default config;
/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	darkMode: 'class',
	theme: {
		extend: {
			colors: {
				border: 'var(--border)',
				input: 'var(--border)',
				ring: 'var(--muted)',
				background: 'var(--bg)',
				foreground: 'var(--text)',
				primary: {
					DEFAULT: 'var(--text)',
					foreground: 'var(--bg)'
				},
				secondary: {
					DEFAULT: 'var(--surface)',
					foreground: 'var(--text)'
				},
				muted: {
					DEFAULT: 'var(--surface)',
					foreground: 'var(--muted)'
				},
				accent: {
					DEFAULT: 'var(--surface)',
					foreground: 'var(--text)'
				},
				destructive: {
					DEFAULT: 'var(--color-conflict)',
					foreground: 'var(--text)'
				}
			},
			fontFamily: {
				sans: ['var(--nano-font-heading)', 'sans-serif'],
				mono: ['var(--nano-font-body)', 'monospace'],
				display: ['var(--nano-font-heading-display)', 'sans-serif'],
				logo: ['var(--nano-font-logo-ext)', 'sans-serif']
			}
		}
	},
	plugins: []
};

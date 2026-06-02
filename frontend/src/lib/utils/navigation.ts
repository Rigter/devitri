/** Full page navigation — bypasses SvelteKit client router on adapter-static. */
export function navigateTo(path: string): void {
	if (typeof window === 'undefined') return;

	const target = path.startsWith('/') ? path : `/${path}`;
	if (window.location.pathname === target) return;

	window.location.href = target;
}

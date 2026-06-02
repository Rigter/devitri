import { toast } from '$lib/stores/toast.svelte';

export async function copyToClipboard(
	text: string,
	successMessage = 'Copied to clipboard'
): Promise<boolean> {
	try {
		await navigator.clipboard.writeText(text);
		toast.success(successMessage);
		return true;
	} catch {
		toast.error('Failed to copy to clipboard');
		return false;
	}
}

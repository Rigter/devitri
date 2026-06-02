export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface ToastItem {
	id: string;
	type: ToastType;
	message: string;
}

const DEFAULT_DURATION_MS = 3200;

class ToastStore {
	items = $state<ToastItem[]>([]);

	push(type: ToastType, message: string, duration = DEFAULT_DURATION_MS): string {
		const id = crypto.randomUUID();
		this.items = [...this.items, { id, type, message }];

		if (duration > 0) {
			setTimeout(() => this.dismiss(id), duration);
		}

		return id;
	}

	dismiss(id: string): void {
		this.items = this.items.filter((item) => item.id !== id);
	}
}

export const toastStore = new ToastStore();

export const toast = {
	success: (message: string, duration?: number) => toastStore.push('success', message, duration),
	error: (message: string, duration?: number) => toastStore.push('error', message, duration ?? 5000),
	info: (message: string, duration?: number) => toastStore.push('info', message, duration),
	warning: (message: string, duration?: number) => toastStore.push('warning', message, duration),
	dismiss: (id: string) => toastStore.dismiss(id)
};

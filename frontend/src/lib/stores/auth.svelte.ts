import { ApiClient } from '../api/client';

const TOKEN_KEY = 'devitri_token';
const DEVICE_ID_KEY = 'devitri_device_id';

function getOrCreateDeviceId(): string {
	if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
		return 'devitri-web';
	}

	let deviceId = localStorage.getItem(DEVICE_ID_KEY);
	if (!deviceId) {
		deviceId = crypto.randomUUID();
		localStorage.setItem(DEVICE_ID_KEY, deviceId);
	}
	return deviceId;
}

class AuthStore {
	token = $state<string | null>(null);
	isAuthenticated = $state(false);
	isLoading = $state(true);

	constructor() {
		if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
			this.token = localStorage.getItem(TOKEN_KEY);
		}
	}

	private persistToken(token: string | null): void {
		this.token = token;
		if (typeof window === 'undefined' || typeof localStorage === 'undefined') return;
		if (token) {
			localStorage.setItem(TOKEN_KEY, token);
		} else {
			localStorage.removeItem(TOKEN_KEY);
		}
	}

	async init(): Promise<void> {
		if (!this.token) {
			this.isAuthenticated = false;
			this.isLoading = false;
			return;
		}

		const api = new ApiClient();
		try {
			await api.getSession();
			this.isAuthenticated = true;
		} catch {
			this.persistToken(null);
			this.isAuthenticated = false;
		} finally {
			this.isLoading = false;
		}
	}

	async login(password: string): Promise<void> {
		const api = new ApiClient();
		const response = await api.login(password, getOrCreateDeviceId(), 'Devitri Web Dashboard');
		this.persistToken(response.token);
		this.isAuthenticated = true;
	}

	async logout(): Promise<void> {
		if (this.token) {
			const api = new ApiClient();
			try {
				await api.logout();
			} catch {
				// Clear local session even if server logout fails
			}
		}
		this.persistToken(null);
		this.isAuthenticated = false;
	}
}

export const authStore = new AuthStore();

export async function initAuth(): Promise<void> {
	await authStore.init();
}

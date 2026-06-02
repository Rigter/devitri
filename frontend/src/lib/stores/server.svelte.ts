import { ApiClient, type SetupCheckResponse } from '../api/client';

class ServerStatusStore {
	ready = $state(false);
	missingFields = $state<string[]>([]);
	isLoading = $state(true);
	checked = $state(false);
	error = $state<string | null>(null);
}

export const serverStatus = new ServerStatusStore();

export async function refreshServerStatus(): Promise<SetupCheckResponse | null> {
	const api = new ApiClient();
	serverStatus.isLoading = true;
	try {
		const status = await api.checkSetup();
		serverStatus.ready = status.ready;
		serverStatus.missingFields = status.missing_fields;
		serverStatus.error = null;
		serverStatus.checked = true;
		return status;
	} catch (e) {
		serverStatus.error = e instanceof Error ? e.message : 'Unknown error';
		serverStatus.checked = true;
		return null;
	} finally {
		serverStatus.isLoading = false;
	}
}

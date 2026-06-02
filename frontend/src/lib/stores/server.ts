import { writable } from 'svelte/store';

export const isServerReady = writable<boolean | null>(null);
export const missingFields = writable<string[]>([]);

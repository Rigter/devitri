<script lang="ts">
	import { toastStore, type ToastType } from '$lib/stores/toast.svelte';
	import { CheckCircle2, AlertCircle, Info, AlertTriangle, X } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	const iconByType: Record<ToastType, typeof CheckCircle2> = {
		success: CheckCircle2,
		error: AlertCircle,
		info: Info,
		warning: AlertTriangle
	};

	const toneByType: Record<ToastType, string> = {
		success: 'border-[var(--color-success)] text-[var(--color-success)]',
		error: 'border-destructive text-destructive',
		info: 'border-border text-foreground',
		warning: 'border-[var(--color-conflict)] text-[var(--color-conflict)]'
	};
</script>

<div
	class="pointer-events-none fixed bottom-4 right-4 z-[100] flex w-full max-w-sm flex-col gap-2"
	aria-live="polite"
	aria-label="Notifications"
>
	{#each toastStore.items as item (item.id)}
		{@const Icon = iconByType[item.type]}
		<div
			role="status"
			class={cn(
				'pointer-events-auto flex items-start gap-3 rounded-lg border bg-background p-4 shadow-lg',
				toneByType[item.type]
			)}
		>
			<Icon size={18} class="mt-0.5 shrink-0" aria-hidden="true" />
			<p class="flex-1 text-sm text-foreground">{item.message}</p>
			<button
				type="button"
				class="shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
				aria-label="Dismiss notification"
				onclick={() => toastStore.dismiss(item.id)}
			>
				<X size={14} />
			</button>
		</div>
	{/each}
</div>

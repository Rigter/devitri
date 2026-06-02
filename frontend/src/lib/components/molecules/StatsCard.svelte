<script lang="ts">
  import { TrendingUp, TrendingDown } from 'lucide-svelte';
  import { cn } from '$lib/utils';

  interface Props {
    title?: string;
    value?: number | string;
    icon?: any; // Component type
    change?: number | null;
    isPositive?: boolean;
  }

  let {
    title = '',
    value = 0,
    icon: Icon = null,
    change = null,
    isPositive = true
  }: Props = $props();

  const changeText = $derived(
    change !== null ? `${isPositive ? '+' : '-'}${Math.abs(change)}%` : ''
  );
</script>

<div class="rounded-xl border bg-card p-6 shadow-sm transition-all hover:shadow-md">
  <div class="flex items-center justify-between space-y-0 pb-2">
    <h3 class="text-xs font-bold uppercase tracking-widest text-muted-foreground">
      {title}
    </h3>
    {#if Icon}
      <div class="text-muted-foreground/50">
        <Icon size={16} />
      </div>
    {/if}
  </div>
  <div class="flex items-baseline justify-between mt-2">
    <div class="text-3xl font-bold tracking-tight">{value}</div>
    {#if change !== null}
      <div class={cn(
        "flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full",
        isPositive ? "bg-emerald-500/10 text-emerald-500" : "bg-red-500/10 text-red-500"
      )}>
        {#if isPositive}
          <TrendingUp size={10} />
        {:else}
          <TrendingDown size={10} />
        {/if}
        {changeText}
      </div>
    {/if}
  </div>
</div>

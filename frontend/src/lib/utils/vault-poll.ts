export type VaultPollStop = () => void;

export interface VaultPollOptions {
  /** Poll interval when the tab is visible. Default 30s. */
  intervalMs?: number;
  /** Run tick once when polling starts. Default false. */
  runOnStart?: boolean;
  /** Called on interval and when the tab becomes visible again. */
  tick: () => void | Promise<void>;
}

/**
 * Polls while the document tab is visible. Runs tick immediately when visibility returns.
 */
export function startVaultPoll(options: VaultPollOptions): VaultPollStop {
  const intervalMs = options.intervalMs ?? 30_000;
  let timer: ReturnType<typeof setInterval> | null = null;

  const runTick = (): void => {
    if (typeof document !== 'undefined' && document.visibilityState === 'hidden') {
      return;
    }
    void options.tick();
  };

  const onVisibility = (): void => {
    if (document.visibilityState === 'visible') {
      runTick();
    }
  };

  timer = setInterval(runTick, intervalMs);

  if (options.runOnStart) {
    runTick();
  }

  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', onVisibility);
  }

  return () => {
    if (timer !== null) {
      clearInterval(timer);
      timer = null;
    }
    if (typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', onVisibility);
    }
  };
}

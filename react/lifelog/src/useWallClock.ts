import { useEffect, useState } from 'react';

import { currentTimeString, minutesOfDay } from './date.ts';

/** Re-renders every `intervalMs` and reports the wall clock. */
export function useWallClock(intervalMs = 60000): { time: string; minutes: number } {
  const [time, setTime] = useState(() => currentTimeString());
  useEffect(() => {
    const timer = setInterval(() => setTime(currentTimeString()), intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);
  return { time, minutes: minutesOfDay(time) };
}

const CYCLE_MS = 30 * 60 * 1000;

export interface DoingTimer {
  /** Countdown within the current 30 minute cycle, for example `12m30s/2`. */
  displayText: string;
  /** Remaining share of the cycle, 0–100, used as the progress overlay width. */
  percentage: number;
}

/**
 * Countdown timer for a task that is currently being worked on.
 *
 * Kept byte-for-byte compatible with lifelog's original timer so both hosts
 * show the same numbers for the same `startedAt`.
 */
export function useDoingTimer(startedAt: string | undefined, enabled: boolean): DoingTimer {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!enabled || !startedAt) return;
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [enabled, startedAt]);
  if (!enabled || !startedAt) return { displayText: '', percentage: 0 };

  const parsed = new Date(startedAt).getTime();
  if (Number.isNaN(parsed)) return { displayText: '', percentage: 0 };
  const elapsed = Math.max(0, now - parsed);
  const cycles = Math.floor(elapsed / CYCLE_MS);
  const remaining = CYCLE_MS - (elapsed % CYCLE_MS);
  const minutes = Math.floor(remaining / 60000);
  const seconds = Math.floor((remaining % 60000) / 1000);
  const text = `${minutes}m${seconds}s`;
  return {
    displayText: cycles > 0 ? `${text}/${cycles}` : text,
    percentage: (remaining / CYCLE_MS) * 100,
  };
}

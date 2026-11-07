/**
 * Pure wall-clock date helpers for the weekly schedule kernel.
 *
 * Everything is local-time and string based (`YYYY-MM-DD`), matching how both
 * hosts store schedule dates. No timezone conversion happens here.
 */

import type { WeeklyScheduleDay, WeeklyScheduleTask } from './types.ts';

const MS_PER_DAY = 86400000;

export const MONDAY_FIRST_WEEKDAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

export function pad2(value: number): string {
  return value < 10 ? `0${value}` : String(value);
}

export function formatISODate(date: Date): string {
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}`;
}

export function localToday(): string {
  return formatISODate(new Date());
}

export function parseISODate(value: string): Date {
  const [y, m, d] = value.split('-').map((part) => Number.parseInt(part, 10));
  return new Date(y, (m || 1) - 1, d || 1);
}

export function addDays(value: string, days: number): string {
  const date = parseISODate(value);
  date.setDate(date.getDate() + days);
  return formatISODate(date);
}

export function weekdayIndex(value: string): number {
  // 0 = Monday … 6 = Sunday
  return (parseISODate(value).getDay() + 6) % 7;
}

export function getWeekMonday(value: string = localToday()): string {
  return addDays(value, -weekdayIndex(value));
}

export function getWeekDays(weekStart: string, today: string): Array<{ date: string; dayNumber: number; isToday: boolean }> {
  return Array.from({ length: 7 }, (_, index) => {
    const date = addDays(weekStart, index);
    return { date, dayNumber: index + 1, isToday: date === today };
  });
}

/** Milliseconds-free day distance, used to keep tasks inside their week. */
export function daysBetween(from: string, to: string): number {
  return Math.round((parseISODate(to).getTime() - parseISODate(from).getTime()) / MS_PER_DAY);
}

export function minutesOfDay(value: string): number {
  const [hours, minutes] = value.split(':').map((part) => Number.parseInt(part, 10));
  if (Number.isNaN(hours) || Number.isNaN(minutes)) return 0;
  return hours * 60 + minutes;
}

export function currentTimeString(now: Date = new Date()): string {
  return `${pad2(now.getHours())}:${pad2(now.getMinutes())}`;
}

export function timeProgress(startTime: string, endTime: string, currentTime: string): number {
  const start = minutesOfDay(startTime);
  const end = minutesOfDay(endTime);
  const current = minutesOfDay(currentTime);
  if (end <= start) return 0;
  return Math.min(100, Math.max(0, ((current - start) / (end - start)) * 100));
}

/**
 * Buckets tasks into the seven days of `weekStart`.
 *
 * Tasks without a date, or outside the week, are dropped: the goal panel is
 * the only place an unscheduled goal shows up.
 */
export function groupTasksByWeek(
  tasks: WeeklyScheduleTask[],
  weekStart: string,
  dayLabels: (date: string, index: number) => string,
): WeeklyScheduleDay[] {
  const byDate = new Map<string, WeeklyScheduleTask[]>();
  for (const task of tasks) {
    if (!task.date) continue;
    const bucket = byDate.get(task.date);
    if (bucket) bucket.push(task);
    else byDate.set(task.date, [task]);
  }
  return getWeekDays(weekStart, weekStart).map(({ date }, index) => ({
    date,
    label: dayLabels(date, index),
    tasks: (byDate.get(date) ?? []).slice().sort((a, b) => minutesOfDay(a.startTime) - minutesOfDay(b.startTime)),
  }));
}

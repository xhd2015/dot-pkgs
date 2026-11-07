import { useState } from 'react';
import type { ReactNode } from 'react';

import { DayList } from './DayList.tsx';
import { GoalPanel } from './GoalPanel.tsx';
import { WeekHeader } from './WeekHeader.tsx';
import { addDays, getWeekMonday, minutesOfDay } from './date.ts';
import { useWallClock } from './useWallClock.ts';
import { DEFAULT_WEEKLY_SCHEDULE_LABELS } from './types.ts';
import type { WeeklyScheduleDay, WeeklyScheduleGoal, WeeklyScheduleLabels } from './types.ts';

export interface WeeklyScheduleViewProps {
  /** Monday of the visible week, `YYYY-MM-DD`. */
  weekStart: string;
  /** Today, `YYYY-MM-DD`. */
  today: string;
  /** Localized range label, for example `2026年11月2日 – 11月8日`. */
  weekLabel: string;
  goals: WeeklyScheduleGoal[];
  days: WeeklyScheduleDay[];
  labels?: Partial<WeeklyScheduleLabels>;
  /** `HH:mm` override; defaults to the live wall clock. */
  currentTime?: string;
  className?: string;
  onWeekChange?: (weekStart: string) => void;
  onTaskClick?: (taskId: string) => void;
  onTaskToggle?: (taskId: string) => void;
  onGoalClick?: (goalId: string) => void;
  onGoalToggle?: (goalId: string) => void;
  onGoalSchedule?: (goal: WeeklyScheduleGoal) => void;
  onAddTask?: (date: string) => void;
  renderWeekPicker?: (close: () => void) => ReactNode;
  renderGoalFooter?: () => ReactNode;
  renderDayFooter?: (day: WeeklyScheduleDay) => ReactNode;
  renderGoalStatus?: (goal: WeeklyScheduleGoal) => ReactNode;
}

/**
 * Shared, presentational weekly schedule.
 *
 * It owns only view state (visible week, focused day, collapsed sections).
 * Every mutation is a callback: both hosts keep their own API, forms and
 * modals, and both map their records onto `WeeklyScheduleGoal`/`Task`.
 */
export function WeeklyScheduleView({
  weekStart,
  today,
  weekLabel,
  goals,
  days,
  labels: labelOverrides,
  currentTime: currentTimeOverride,
  className,
  onWeekChange,
  onTaskClick,
  onTaskToggle,
  onGoalClick,
  onGoalToggle,
  onGoalSchedule,
  onAddTask,
  renderWeekPicker,
  renderGoalFooter,
  renderDayFooter,
  renderGoalStatus,
}: WeeklyScheduleViewProps) {
  const [focusedDate, setFocusedDate] = useState<string | null>(null);
  const clock = useWallClock();
  const currentTime = currentTimeOverride ?? clock.time;
  const labels: WeeklyScheduleLabels = { ...DEFAULT_WEEKLY_SCHEDULE_LABELS, ...labelOverrides };

  return (
    <div className={`ws-root${className ? ` ${className}` : ''}`}>
      <WeekHeader
        weekStart={weekStart}
        weekLabel={weekLabel}
        today={today}
        labels={labels}
        focusedDate={focusedDate}
        onWeekChange={(delta) => {
          setFocusedDate(null);
          onWeekChange?.(addDays(weekStart, delta * 7));
        }}
        onToday={() => {
          setFocusedDate(null);
          onWeekChange?.(getWeekMonday(today));
        }}
        onDayFocus={setFocusedDate}
        renderWeekPicker={renderWeekPicker}
      />

      <GoalPanel
        goals={goals}
        labels={labels}
        onGoalClick={onGoalClick}
        onGoalToggle={onGoalToggle}
        onGoalSchedule={onGoalSchedule}
        renderGoalStatus={renderGoalStatus}
        renderGoalFooter={renderGoalFooter}
      />

      <DayList
        days={days}
        today={today}
        focusedDate={focusedDate}
        labels={labels}
        currentTime={currentTime}
        nowMinutes={minutesOfDay(currentTime)}
        onDayFocus={setFocusedDate}
        onTaskClick={onTaskClick}
        onTaskToggle={onTaskToggle}
        onAddTask={onAddTask}
        renderDayFooter={renderDayFooter}
      />
    </div>
  );
}

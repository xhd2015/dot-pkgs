/**
 * Public entry for the shared weekly schedule.
 *
 * Import as `@xhd2015/dot-pkgs-react-diagrams/lifelog`.
 *
 * `schedule.css` is imported by `WeeklyScheduleView.tsx`, not here: a bare
 * `import './schedule.css'` inside this barrel is dropped by rolldown, which
 * treats a module listed in the package's `sideEffects` array as further
 * shakeable and collapses the barrel away. Keeping the import next to the
 * component that needs the styles survives tree-shaking and keeps the CSS in
 * the chunk that renders the schedule.
 */
export { WeeklyScheduleView } from './WeeklyScheduleView.tsx';
export type { WeeklyScheduleViewProps } from './WeeklyScheduleView.tsx';
export { WeekHeader } from './WeekHeader.tsx';
export { GoalPanel } from './GoalPanel.tsx';
export type { GoalPanelProps } from './GoalPanel.tsx';
export { DayList } from './DayList.tsx';
export type { DayListProps } from './DayList.tsx';
export { TaskGrid } from './TaskGrid.tsx';
export type { TaskGridProps } from './TaskGrid.tsx';
export {
  MONDAY_FIRST_WEEKDAY_NAMES,
  addDays,
  currentTimeString,
  daysBetween,
  formatISODate,
  getWeekDays,
  getWeekMonday,
  groupTasksByWeek,
  localToday,
  minutesOfDay,
  parseISODate,
  timeProgress,
  weekdayIndex,
} from './date.ts';
export { useDoingTimer, useWallClock } from './useWallClock.ts';
export type { DoingTimer } from './useWallClock.ts';
export { DEFAULT_WEEKLY_SCHEDULE_LABELS } from './types.ts';
export type {
  WeeklyScheduleDay,
  WeeklyScheduleGoal,
  WeeklyScheduleLabels,
  WeeklyScheduleTask,
} from './types.ts';

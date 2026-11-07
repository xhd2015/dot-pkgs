/**
 * Presentational types for the shared weekly schedule view.
 *
 * The kernel deliberately avoids host concepts such as `TaskStatus` or
 * `WeeklyGoal`: hosts map their own records onto these flat shapes.
 */

export interface WeeklyScheduleTask {
  id: string;
  title: string;
  /** `HH:mm`, wall clock. */
  startTime: string;
  /** `HH:mm`, wall clock. */
  endTime: string;
  /** `YYYY-MM-DD` of the day the task is scheduled to. */
  date?: string;
  /** Completed state, drives the checkbox and strike-through. */
  done?: boolean;
  /** Ended (done/expired/archived) but not necessarily completed. */
  ended?: boolean;
  /** Marks a task that is currently being worked on. */
  doing?: boolean;
  /** Wall clock `HH:mm` when the task was marked doing, enables the timer. */
  doingStartedAt?: string;
  /** Overdue state, drives the red alert icon. */
  overdue?: boolean;
  /** Renders the small GOAL tag next to the title. */
  goal?: boolean;
}

export interface WeeklyScheduleGoal {
  id: string;
  title: string;
  done?: boolean;
  /** Set when the goal has been scheduled onto a day. */
  dueDate?: string;
  doneDate?: string;
  doneCount?: number;
  totalCount?: number;
}

export interface WeeklyScheduleDay {
  /** `YYYY-MM-DD` */
  date: string;
  /** Host-localized day label, for example `Fri 11/06`. */
  label: string;
  tasks: WeeklyScheduleTask[];
}

export interface WeeklyScheduleLabels {
  goalsTitle: string;
  goalsProgress: (done: number, total: number) => string;
  goalsEmpty: string;
  scheduleTitle: string;
  scheduleCount: (count: number) => string;
  emptyDay: string;
  /** Shown in the time column for a task without `startTime`/`endTime`. */
  untimed: string;
  showAllWeek: string;
  scheduleGoal: string;
  unscheduled: string;
  done: string;
  scheduled: string;
  today: string;
  /** Monday-first weekday names for the week strip. */
  weekdays: string[];
  previousWeek: string;
  nextWeek: string;
  pickWeek: string;
  addTask: string;
  addGoal: string;
  linkGoal: string;
  goalTag: string;
  collapse: string;
  expand: string;
}

export const DEFAULT_WEEKLY_SCHEDULE_LABELS: WeeklyScheduleLabels = {
  goalsTitle: "This Week's Goals",
  goalsProgress: (done, total) => `${done}/${total} completed`,
  goalsEmpty: 'No goals set for this week',
  scheduleTitle: 'Schedule',
  scheduleCount: (count) => `${count} task${count === 1 ? '' : 's'}`,
  emptyDay: 'No tasks scheduled',
  untimed: 'Anytime',
  showAllWeek: 'Show all week',
  scheduleGoal: 'Schedule this goal',
  unscheduled: 'unscheduled',
  done: 'done',
  scheduled: 'scheduled',
  today: 'Today',
  weekdays: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
  previousWeek: 'Previous week',
  nextWeek: 'Next week',
  pickWeek: 'Pick a week',
  addTask: 'Add',
  addGoal: 'New goal',
  linkGoal: 'Add existing',
  goalTag: 'GOAL',
  collapse: 'Collapse',
  expand: 'Expand',
};

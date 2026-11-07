import { useState } from 'react';
import type { ReactNode } from 'react';

import { ChevronForwardIcon, ClockIcon, PlusIcon } from './icons.tsx';
import { TaskGrid } from './TaskGrid.tsx';
import type { WeeklyScheduleDay, WeeklyScheduleLabels } from './types.ts';

export interface DayListProps {
  days: WeeklyScheduleDay[];
  today: string;
  focusedDate: string | null;
  labels: WeeklyScheduleLabels;
  /** `HH:mm` wall clock; the red now-line only shows on `today`. */
  currentTime: string;
  nowMinutes: number;
  onDayFocus: (date: string | null) => void;
  onTaskClick?: (taskId: string) => void;
  onTaskToggle?: (taskId: string) => void;
  onAddTask?: (date: string) => void;
  /** Host-owned action row under a day, replaces the default add button. */
  renderDayFooter?: (day: WeeklyScheduleDay) => ReactNode;
}

function DaySection({
  day,
  dayNumber,
  today,
  isToday,
  collapsedByDefault,
  labels,
  currentTime,
  nowMinutes,
  onAddTask,
  onTaskClick,
  onTaskToggle,
  renderDayFooter,
}: {
  day: WeeklyScheduleDay;
  dayNumber: number;
  today: string;
  isToday: boolean;
  collapsedByDefault: boolean;
  labels: WeeklyScheduleLabels;
  currentTime: string;
  nowMinutes: number;
  onAddTask?: (date: string) => void;
  onTaskClick?: (taskId: string) => void;
  onTaskToggle?: (taskId: string) => void;
  renderDayFooter?: (day: WeeklyScheduleDay) => ReactNode;
}) {
  const [open, setOpen] = useState(!collapsedByDefault);
  const total = day.tasks.length;
  const done = day.tasks.filter((task) => task.done).length;
  const isPast = day.date < today;
  const classes = ['ws-day-section'];
  if (isToday) classes.push('ws-day-section--today');

  return (
    <div className={classes.join(' ')}>
      <div
        className="ws-day-separator ws-day-separator--clickable"
        role="button"
        tabIndex={0}
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            setOpen((value) => !value);
          }
        }}
      >
        <span className={`ws-day-number${isToday ? ' ws-day-number--today' : ''}`}>{dayNumber}</span>
        <span className="ws-day-separator-label">{day.label}</span>
        <span className="ws-day-separator-count">{done}/{total}</span>
        <span className={`ws-collapse-icon${open ? ' ws-collapse-icon--open' : ''}`}>
          <ChevronForwardIcon size={14} />
        </span>
      </div>

      {open && (
        <>
          {total > 0
            ? <TaskGrid tasks={day.tasks} labels={labels} currentTime={isToday ? currentTime : undefined} nowMinutes={isToday ? nowMinutes : undefined} onTaskClick={onTaskClick} onTaskToggle={onTaskToggle} />
            : <div className="ws-empty ws-empty--compact">{labels.emptyDay}</div>}

          {renderDayFooter
            ? renderDayFooter(day)
            : !isPast && onAddTask && (
              <div className="ws-day-add-row">
                <button type="button" className="ws-day-action-btn ws-day-action-btn--primary" onClick={() => onAddTask(day.date)}>
                  <PlusIcon size={14} />
                  <span>{labels.addTask}</span>
                </button>
              </div>
            )}
        </>
      )}
    </div>
  );
}

/**
 * The schedule section: a header with the task count and the seven collapsible
 * day sections. Days before today (current week only) start collapsed so the
 * user lands on what is still actionable.
 */
export function DayList({
  days,
  today,
  focusedDate,
  labels,
  currentTime,
  nowMinutes,
  onDayFocus,
  onTaskClick,
  onTaskToggle,
  onAddTask,
  renderDayFooter,
}: DayListProps) {
  const isCurrentWeek = days.some((day) => day.date === today);
  const visible = focusedDate ? days.filter((day) => day.date === focusedDate) : days;
  const taskCount = visible.reduce((sum, day) => sum + day.tasks.length, 0);
  const summary = focusedDate
    ? (visible[0]?.label ?? labels.scheduleTitle)
    : labels.scheduleTitle;

  return (
    <section className="ws-section">
      <div className="ws-section-header">
        <ClockIcon size={18} />
        <h3>{summary}</h3>
        <span className="ws-section-count">{labels.scheduleCount(taskCount)}</span>
        {focusedDate && (
          <button type="button" className="ws-show-all-week" onClick={() => onDayFocus(null)}>
            {labels.showAllWeek}
          </button>
        )}
      </div>

      {visible.map((day) => {
        const dayNumber = days.findIndex((candidate) => candidate.date === day.date) + 1;
        return (
          <DaySection
            key={day.date}
            day={day}
            dayNumber={dayNumber}
            today={today}
            isToday={day.date === today}
            collapsedByDefault={isCurrentWeek && day.date < today}
            labels={labels}
            currentTime={currentTime}
            nowMinutes={nowMinutes}
            onAddTask={onAddTask}
            onTaskClick={onTaskClick}
            onTaskToggle={onTaskToggle}
            renderDayFooter={renderDayFooter}
          />
        );
      })}
    </section>
  );
}

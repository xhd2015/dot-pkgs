import { useState } from 'react';
import type { ReactNode } from 'react';

import { CalendarIcon, ChevronForwardIcon, CheckCircleIcon, EllipseIcon, FlagIcon } from './icons.tsx';
import type { WeeklyScheduleGoal, WeeklyScheduleLabels } from './types.ts';

export interface GoalPanelProps {
  goals: WeeklyScheduleGoal[];
  labels: WeeklyScheduleLabels;
  collapsed?: boolean;
  onGoalClick?: (goalId: string) => void;
  onGoalToggle?: (goalId: string) => void;
  onGoalSchedule?: (goal: WeeklyScheduleGoal) => void;
  /** Replaces the default `done`/`scheduled …`/`unscheduled` text. */
  renderGoalStatus?: (goal: WeeklyScheduleGoal) => ReactNode;
  /** Host-owned "add goal" row, rendered below the list. */
  renderGoalFooter?: () => ReactNode;
}

/**
 * Green weekly-goals card: completion bar plus one row per goal.
 *
 * Scheduling a goal (the little calendar glyph) is only offered while the goal
 * has no `dueDate`, mirroring lifelog.
 */
export function GoalPanel({
  goals,
  labels,
  collapsed = false,
  onGoalClick,
  onGoalToggle,
  onGoalSchedule,
  renderGoalStatus,
  renderGoalFooter,
}: GoalPanelProps) {
  const [open, setOpen] = useState(!collapsed);
  const total = goals.length;
  const done = goals.filter((goal) => goal.done).length;
  const percentage = total > 0 ? Math.round((done / total) * 100) : 0;

  return (
    <section className="ws-section ws-section--goals">
      <div
        className="ws-section-header ws-section-header--clickable"
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
        <FlagIcon size={18} />
        <h3>{labels.goalsTitle}</h3>
        <span className="ws-section-count">{labels.goalsProgress(done, total)}</span>
        <span className={`ws-collapse-icon${open ? ' ws-collapse-icon--open' : ''}`}>
          <ChevronForwardIcon size={14} />
        </span>
      </div>

      {open && (
        <>
          {total > 0 && (
            <div className="ws-overall-progress">
              <div className="ws-overall-progress-bar">
                <div className="ws-overall-progress-fill" style={{ width: `${percentage}%` }} />
              </div>
              <span className="ws-overall-progress-text">{percentage}%</span>
            </div>
          )}

          {total > 0 ? (
            <div className="ws-goals-list">
              {goals.map((goal) => {
                const unscheduled = !goal.done && !goal.dueDate;
                return (
                  <div
                    key={goal.id}
                    className={`ws-goal-item${goal.done ? ' ws-goal-item--done' : ''}`}
                    onClick={() => onGoalClick?.(goal.id)}
                  >
                    <span
                      className="ws-goal-checkbox"
                      role="checkbox"
                      aria-checked={Boolean(goal.done)}
                      aria-label={goal.title}
                      tabIndex={0}
                      onClick={(event) => { event.stopPropagation(); onGoalToggle?.(goal.id); }}
                      onKeyDown={(event) => {
                        if (event.key === 'Enter' || event.key === ' ') {
                          event.preventDefault();
                          event.stopPropagation();
                          onGoalToggle?.(goal.id);
                        }
                      }}
                    >
                      {goal.done
                        ? <CheckCircleIcon size={18} className="ws-status-icon ws-status-icon--done" />
                        : <EllipseIcon size={18} className="ws-status-icon ws-status-icon--idle" />}
                    </span>
                    <span className={`ws-goal-title${goal.done ? ' ws-goal-title--done' : ''}`}>{goal.title}</span>
                    <span className={`ws-goal-status-label${goal.done ? ' ws-goal-status-label--done' : unscheduled ? ' ws-goal-status-label--pending' : ' ws-goal-status-label--scheduled'}`}>
                      {renderGoalStatus ? renderGoalStatus(goal) : defaultGoalStatus(goal, labels)}
                    </span>
                    {unscheduled && onGoalSchedule && (
                      <span
                        className="ws-goal-schedule-btn"
                        role="button"
                        tabIndex={0}
                        title={labels.scheduleGoal}
                        aria-label={labels.scheduleGoal}
                        onClick={(event) => { event.stopPropagation(); onGoalSchedule(goal); }}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter' || event.key === ' ') {
                            event.preventDefault();
                            event.stopPropagation();
                            onGoalSchedule(goal);
                          }
                        }}
                      >
                        <CalendarIcon size={14} />
                      </span>
                    )}
                    {goal.totalCount != null && goal.totalCount > 0 && (
                      <span className="ws-goal-count">{goal.doneCount ?? 0}/{goal.totalCount}</span>
                    )}
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="ws-empty">{labels.goalsEmpty}</div>
          )}

          {renderGoalFooter?.()}
        </>
      )}
    </section>
  );
}

function defaultGoalStatus(goal: WeeklyScheduleGoal, labels: WeeklyScheduleLabels): string {
  if (goal.done) return labels.done;
  if (goal.dueDate) return `${labels.scheduled} ${goal.dueDate.slice(5)}`;
  return labels.unscheduled;
}

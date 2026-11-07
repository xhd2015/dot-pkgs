import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';

import { AlertCircleIcon, CheckCircleIcon, EllipseIcon } from './icons.tsx';
import { minutesOfDay, timeProgress } from './date.ts';
import { useDoingTimer } from './useWallClock.ts';
import type { WeeklyScheduleLabels, WeeklyScheduleTask } from './types.ts';

interface TimeGroup {
  key: string;
  startTime: string;
  endTime: string;
  tasks: WeeklyScheduleTask[];
}

function groupByTimeRange(tasks: WeeklyScheduleTask[]): TimeGroup[] {
  const buckets = new Map<string, WeeklyScheduleTask[]>();
  const order: string[] = [];
  for (const task of tasks) {
    const key = `${task.startTime}|${task.endTime}`;
    const bucket = buckets.get(key);
    if (bucket) bucket.push(task);
    else {
      buckets.set(key, [task]);
      order.push(key);
    }
  }
  return order.map((key) => {
    const [startTime, endTime] = key.split('|');
    return { key, startTime, endTime, tasks: buckets.get(key) ?? [] };
  });
}

function TaskStatusIcon({ task }: { task: WeeklyScheduleTask }) {
  if (task.done) return <CheckCircleIcon size={16} className="ws-status-icon ws-status-icon--done" />;
  if (task.overdue) return <AlertCircleIcon size={16} className="ws-status-icon ws-status-icon--overdue" />;
  return <EllipseIcon size={16} className="ws-status-icon ws-status-icon--idle" />;
}

function TaskCell({ task, labels, onTaskClick, onTaskToggle, gridRow, borderBottom }: {
  task: WeeklyScheduleTask;
  labels: WeeklyScheduleLabels;
  onTaskClick?: (taskId: string) => void;
  onTaskToggle?: (taskId: string) => void;
  gridRow: number;
  borderBottom?: string;
}) {
  const timer = useDoingTimer(task.doingStartedAt, Boolean(task.doing && task.doingStartedAt));
  return (
    <div
      className={`ws-table-task-cell${task.done ? ' ws-table-task-cell--done' : ''}`}
      style={{ gridColumn: 2, gridRow, borderBottom }}
      onClick={() => onTaskClick?.(task.id)}
    >
      {timer.percentage > 0 && (
        <div className="ws-task-progress" style={{ width: `${timer.percentage}%` }} />
      )}
      <span
        className="ws-task-status-icon"
        role="checkbox"
        aria-checked={Boolean(task.done)}
        aria-label={task.title}
        tabIndex={0}
        onClick={(event) => { event.stopPropagation(); onTaskToggle?.(task.id); }}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            event.stopPropagation();
            onTaskToggle?.(task.id);
          }
        }}
      >
        <TaskStatusIcon task={task} />
      </span>
      {task.goal && <span className="ws-task-goal-tag">{labels.goalTag}</span>}
      <span className={`ws-task-title${task.ended ? ' ws-task-title--ended' : ''}`}>{task.title}</span>
      {timer.displayText && <span className="ws-task-timer">{timer.displayText}</span>}
    </div>
  );
}

export interface TaskGridProps {
  tasks: WeeklyScheduleTask[];
  labels: WeeklyScheduleLabels;
  /** `HH:mm`; when both are set a red "now" line is drawn. */
  currentTime?: string;
  nowMinutes?: number;
  onTaskClick?: (taskId: string) => void;
  onTaskToggle?: (taskId: string) => void;
}

/**
 * Table-style task list: the time range sits on the left and spans every task
 * sharing it, so a 09:00–10:00 block reads as one row with its tasks nested.
 */
export function TaskGrid({ tasks, labels, currentTime, nowMinutes, onTaskClick, onTaskToggle }: TaskGridProps) {
  const groups = groupByTimeRange(tasks);
  const gridRef = useRef<HTMLDivElement>(null);
  const [insideTop, setInsideTop] = useState<number | null>(null);
  const showIndicator = nowMinutes != null && currentTime != null;

  let beforeGroup = -1;
  let insideGroup = -1;
  let insideProgress = 0;
  let afterAll = false;
  if (showIndicator) {
    let placed = false;
    for (let index = 0; index < groups.length; index++) {
      const group = groups[index];
      const start = minutesOfDay(group.startTime);
      const end = minutesOfDay(group.endTime);
      if (nowMinutes! < start) {
        beforeGroup = index;
        placed = true;
        break;
      }
      if (nowMinutes! >= start && nowMinutes! < end) {
        insideGroup = index;
        insideProgress = timeProgress(group.startTime, group.endTime, currentTime!);
        placed = true;
        break;
      }
    }
    if (!placed) afterAll = true;
  }

  useEffect(() => {
    if (insideGroup < 0 || !gridRef.current) {
      setInsideTop(null);
      return;
    }
    const cells = gridRef.current.querySelectorAll<HTMLElement>('.ws-table-time-cell');
    const cell = cells[insideGroup];
    if (!cell) {
      setInsideTop(null);
      return;
    }
    const gridRect = gridRef.current.getBoundingClientRect();
    const cellRect = cell.getBoundingClientRect();
    setInsideTop(cellRect.top - gridRect.top + (cellRect.height * insideProgress) / 100);
  }, [insideGroup, insideProgress, tasks]);

  const renderIndicatorRow = (gridRow: number) => (
    <div key={`indicator-${gridRow}`} className="ws-table-indicator-row" style={{ gridColumn: '1 / -1', gridRow }}>
      <span className="ws-table-indicator-time">{currentTime}</span>
      <span className="ws-table-indicator-dot" />
      <span className="ws-table-indicator-line" />
    </div>
  );

  const elements: ReactNode[] = [];
  let rowIndex = 0;

  groups.forEach((group, groupIndex) => {
    if (beforeGroup === groupIndex) {
      rowIndex += 1;
      elements.push(renderIndicatorRow(rowIndex));
    }
    const startRow = rowIndex + 1;
    const span = group.tasks.length;
    const isLastGroup = groupIndex === groups.length - 1;
    const separator = isLastGroup ? undefined : '1px solid rgba(0,0,0,0.08)';

    group.tasks.forEach((task, taskIndex) => {
      rowIndex += 1;
      const isLastInGroup = taskIndex === span - 1;
      elements.push(
        <TaskCell
          key={task.id}
          task={task}
          labels={labels}
          onTaskClick={onTaskClick}
          onTaskToggle={onTaskToggle}
          gridRow={rowIndex}
          borderBottom={isLastInGroup ? separator : '1px solid rgba(0,0,0,0.04)'}
        />,
      );
    });

    elements.push(
      <div
        key={`time-${group.key}`}
        className="ws-table-time-cell"
        style={{ gridColumn: 1, gridRow: `${startRow} / span ${span}`, borderBottom: separator }}
      >
        <span className="ws-table-time-label">{group.startTime && group.endTime ? `${group.startTime} – ${group.endTime}` : labels.untimed}</span>
      </div>,
    );
  });

  if (afterAll) {
    rowIndex += 1;
    elements.push(renderIndicatorRow(rowIndex));
  }

  const totalRows = tasks.length + (beforeGroup >= 0 || afterAll ? 1 : 0);

  return (
    <div
      ref={gridRef}
      className="ws-table-grid"
      style={{ gridTemplateRows: `repeat(${Math.max(totalRows, 1)}, auto)` }}
    >
      {elements}
      {insideTop != null && (
        <div className="ws-table-inside-indicator" style={{ top: insideTop }}>
          <span className="ws-table-indicator-time ws-table-indicator-time--inside">{currentTime}</span>
          <span className="ws-table-indicator-dot" />
          <span className="ws-table-indicator-line" />
        </div>
      )}
    </div>
  );
}

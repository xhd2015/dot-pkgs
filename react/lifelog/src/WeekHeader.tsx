import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';

import { CalendarIcon, ChevronBackIcon, ChevronForwardIcon, LocateIcon } from './icons.tsx';
import { getWeekDays } from './date.ts';
import type { WeeklyScheduleLabels } from './types.ts';

export interface WeekHeaderProps {
  /** Monday of the visible week, `YYYY-MM-DD`. */
  weekStart: string;
  /** Host-localized range label, for example `Nov 3 – Nov 9`. */
  weekLabel: string;
  /** Today, `YYYY-MM-DD`; drives the "today" highlight and the jump button. */
  today: string;
  labels: WeeklyScheduleLabels;
  focusedDate: string | null;
  onWeekChange: (delta: number) => void;
  /** Jumps back to the week that contains `today`. */
  onToday: () => void;
  onDayFocus: (date: string | null) => void;
  /** Arrow-less week picker; the host renders its own calendar/popover. */
  renderWeekPicker?: (close: () => void) => ReactNode;
}
/**
 * Purple gradient week strip: arrows, range button, "today" shortcut and the
 * seven day chips. Clicking a chip focuses that day; clicking it again clears
 * the focus and shows the whole week.
 */
export function WeekHeader({
  weekStart,
  weekLabel,
  today,
  labels,
  focusedDate,
  onWeekChange,
  onToday,
  onDayFocus,
  renderWeekPicker,
}: WeekHeaderProps) {
  const [pickerOpen, setPickerOpen] = useState(false);
  const pickerRef = useRef<HTMLDivElement | null>(null);
  const days = getWeekDays(weekStart, today);
  const isCurrentWeek = days.some((day) => day.isToday);

  // Tapping anywhere outside the popup closes it, which is the only way out on
  // a phone where there is no Escape key.
  useEffect(() => {
    if (!pickerOpen) return;
    const onPointerDown = (event: PointerEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        setPickerOpen(false);
      }
    };
    document.addEventListener('pointerdown', onPointerDown, true);
    return () => document.removeEventListener('pointerdown', onPointerDown, true);
  }, [pickerOpen]);

  return (
    <div className="ws-week-header">
      <div className="ws-week-header-top">
        <button
          type="button"
          className="ws-week-arrow"
          aria-label={labels.previousWeek}
          onClick={() => onWeekChange(-1)}
        >
          <ChevronBackIcon size={16} />
        </button>
        <div className="ws-week-header-center">
          <button
            type="button"
            className="ws-week-range-btn"
            aria-label={labels.pickWeek}
            onClick={() => setPickerOpen((open) => !open)}
          >
            <CalendarIcon size={13} />
            <span>{weekLabel}</span>
          </button>
          {pickerOpen && renderWeekPicker && (
            <div className="ws-week-picker-popup" ref={pickerRef}>
              {renderWeekPicker(() => setPickerOpen(false))}
            </div>
          )}
        </div>
        {!isCurrentWeek && (
          <button
            type="button"
            className="ws-week-today-btn"
            aria-label={labels.today}
            onClick={() => {
              onToday();
              onDayFocus(null);
            }}
          >
            <LocateIcon size={12} />
            <span>{labels.today}</span>
          </button>
        )}
        <button
          type="button"
          className="ws-week-arrow"
          aria-label={labels.nextWeek}
          onClick={() => onWeekChange(1)}
        >
          <ChevronForwardIcon size={16} />
        </button>
      </div>

      <div className="ws-week-days">
        {days.map((day, index) => {
          const isFocused = focusedDate === day.date;
          const classes = ['ws-week-day'];
          if (day.isToday) classes.push('ws-week-day--today');
          if (isFocused) classes.push('ws-week-day--focused');
          return (
            <button
              type="button"
              key={day.date}
              className={classes.join(' ')}
              aria-pressed={isFocused}
              onClick={() => onDayFocus(isFocused ? null : day.date)}
            >
              <span className="ws-week-day-name">{labels.weekdays?.[index] ?? ''}</span>
              <span className="ws-week-day-num">{Number(day.date.slice(8, 10))}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

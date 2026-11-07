/**
 * Inline SVG icon set for the weekly schedule kernel.
 *
 * The kernel must not depend on `react-icons`, so the handful of glyphs it
 * needs are inlined here and inherit `currentColor`/`font-size`.
 */

import type { SVGProps } from 'react';

type IconProps = SVGProps<SVGSVGElement> & { size?: number };

const base = (size: number | undefined): SVGProps<SVGSVGElement> => ({
  width: size ?? '1em',
  height: size ?? '1em',
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round',
  strokeLinejoin: 'round',
  'aria-hidden': true,
  focusable: false,
});

export const ChevronBackIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}><path d="M15 18l-6-6 6-6" /></svg>
);

export const ChevronForwardIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}><path d="M9 18l6-6-6-6" /></svg>
);

export const CalendarIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <rect x="3" y="5" width="18" height="16" rx="2" />
    <path d="M8 3v4M16 3v4M3 11h18" />
  </svg>
);

export const LocateIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <circle cx="12" cy="12" r="7" />
    <path d="M12 2v3M12 19v3M2 12h3M19 12h3" />
  </svg>
);

export const FlagIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <path d="M5 21V4M5 4h11l-1.5 4L16 12H5" />
  </svg>
);

export const ClockIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <circle cx="12" cy="12" r="9" />
    <path d="M12 7v5l3 2" />
  </svg>
);

export const PlusIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}><path d="M12 5v14M5 12h14" /></svg>
);

export const LinkIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <path d="M10 13a5 5 0 0 0 7 0l2-2a5 5 0 0 0-7-7l-1 1" />
    <path d="M14 11a5 5 0 0 0-7 0l-2 2a5 5 0 0 0 7 7l1-1" />
  </svg>
);

export const CheckCircleIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <circle cx="12" cy="12" r="9" />
    <path d="M8.5 12.5l2.5 2.5 4.5-5" />
  </svg>
);

export const EllipseIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}><circle cx="12" cy="12" r="9" /></svg>
);

export const AlertCircleIcon = ({ size, ...rest }: IconProps) => (
  <svg {...base(size)} {...rest}>
    <circle cx="12" cy="12" r="9" />
    <path d="M12 7.5v5M12 16.2v.3" />
  </svg>
);

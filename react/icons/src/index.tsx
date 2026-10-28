import type { ReactNode } from 'react';
import { BraveIcon, ChromeIcon, FirefoxIcon, OperaIcon, SafariIcon } from './browsers';
import { CodexIcon, CommandCodeIcon, DshIcon, EdgeIcon, GrokIcon } from './marks';

export { BraveIcon, ChromeIcon, CodexIcon, CommandCodeIcon, DshIcon, EdgeIcon, FirefoxIcon, GrokIcon, OperaIcon, SafariIcon };

export function RunnerIcon({ id, size = 16 }: { id: string; size?: number }): ReactNode {
    switch (id) {
        case 'dsh':
            return <DshIcon size={size} />;
        case 'grok':
            return <GrokIcon size={size} />;
        case 'codex':
            return <CodexIcon size={size} />;
        case 'commandcode':
            return <CommandCodeIcon size={size} />;
        default:
            return null;
    }
}

export function BrowserIcon({ id, size = 16 }: { id: string; size?: number }): ReactNode {
    switch (id) {
        case 'brave':
            return <BraveIcon size={size} />;
        case 'chrome':
            return <ChromeIcon size={size} />;
        case 'firefox':
            return <FirefoxIcon size={size} />;
        case 'edge':
            return <EdgeIcon size={size} />;
        case 'safari':
            return <SafariIcon size={size} />;
        case 'opera':
            return <OperaIcon size={size} />;
        default:
            return null;
    }
}

export function UsageKindIcon({ kind, size = 16 }: { kind?: string; size?: number }): ReactNode {
    return <RunnerIcon id={(kind ?? '').toLowerCase()} size={size} />;
}

export type TerminalClientMessage =
    | { type: 'create'; tabId: string; cols: number; rows: number }
    | { type: 'input'; tabId: string; data: string }
    | { type: 'resize'; tabId: string; cols: number; rows: number }
    | { type: 'close'; tabId: string }

export function terminalSize(cols?: number | null, rows?: number | null): { cols: number; rows: number } {
    const c = Math.floor(cols ?? 0)
    const r = Math.floor(rows ?? 0)
    return { cols: c >= 2 ? c : 80, rows: r >= 1 ? r : 24 }
}

export function terminalCreate(tabId: string, cols?: number | null, rows?: number | null): TerminalClientMessage {
    const size = terminalSize(cols, rows)
    return { type: 'create', tabId, cols: size.cols, rows: size.rows }
}

export function terminalInput(tabId: string, data: string): TerminalClientMessage {
    return { type: 'input', tabId, data }
}

export function terminalResize(tabId: string, cols?: number | null, rows?: number | null): TerminalClientMessage {
    const size = terminalSize(cols, rows)
    return { type: 'resize', tabId, cols: size.cols, rows: size.rows }
}

export function terminalClose(tabId: string): TerminalClientMessage {
    return { type: 'close', tabId }
}

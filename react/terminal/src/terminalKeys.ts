export const terminalKeys = {
    tab: '\t',
    left: '\x1b[D',
    up: '\x1b[A',
    down: '\x1b[B',
    right: '\x1b[C',
    esc: '\x1b',
    c: '\x03',
    a: '\x01',
    r: '\x12',
    l: '\x0c',
} as const

/** Ctrl-A is 0x01 through Ctrl-Z 0x1a. Anything else is not a control letter. */
export function controlChar(letter: string): string | null {
    if (letter.length !== 1) return null
    const code = letter.toLowerCase().charCodeAt(0)
    if (code < 97 || code > 122) return null
    return String.fromCharCode(code - 96)
}

/** How much of the layout viewport is covered by the keyboard. */
export function keyboardInset(innerHeight: number, viewportHeight: number, offsetTop = 0): number {
    if (!Number.isFinite(innerHeight) || !Number.isFinite(viewportHeight)) return 0
    const offset = Number.isFinite(offsetTop) ? offsetTop : 0
    return Math.max(0, innerHeight - viewportHeight - offset)
}

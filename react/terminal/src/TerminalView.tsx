import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'
import { useEffect, useRef, useState } from 'react'
import { TerminalKeyPicker } from './TerminalKeyPicker.tsx'
import { terminalResize } from './protocol.ts'
import './TerminalView.css'

export type TerminalViewProps = {
    url: string
    sessionId: string
    active?: boolean
    /** delete reaps the shell. The default keeps it so the modal can reopen. */
    dispose?: 'keep' | 'delete'
    onData?: (data: string) => void
    onSessionId?: (id: string) => void
    onStatus?: (status: string) => void
    onReady?: (api: { term: Terminal; send: (data: string) => void }) => void
}

export function TerminalView({ url, sessionId, active = true, dispose = 'keep', onData, onSessionId, onStatus, onReady }: TerminalViewProps) {
    const ref = useRef<HTMLDivElement>(null)
    const sendRef = useRef<(data: string) => void>(() => {})
    const [host, setHost] = useState<HTMLDivElement | null>(null)
    const activeRef = useRef(active)
    activeRef.current = active
    const disposeRef = useRef(dispose)
    disposeRef.current = dispose

    useEffect(() => {
        const el = ref.current
        if (!el || !url || !sessionId) return
        const term = new Terminal({
            cursorBlink: true,
            fontSize: window.matchMedia('(pointer: coarse)').matches ? 16 : 14,
            fontFamily: '"SF Mono", Monaco, Menlo, Consolas, monospace',
            theme: { background: '#000000', foreground: '#ffffff', cursor: '#ffffff' },
            scrollback: 10000,
        })
        const fit = new FitAddon()
        term.loadAddon(fit)
        term.loadAddon(new WebLinksAddon())
        term.open(el)
        fit.fit()
        if (activeRef.current) term.focus()

        const ws = new WebSocket(url)
        ws.binaryType = 'arraybuffer'
        const sendJSON = (message: object) => {
            if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify(message))
        }
        const sendInput = (data: string) => {
            if (ws.readyState === WebSocket.OPEN) ws.send(data)
        }
        sendRef.current = sendInput
        ws.onopen = () => {
            onStatus?.('Connected')
            const size = fit.proposeDimensions()
            sendJSON(terminalResize(sessionId, size?.cols, size?.rows))
            fit.fit()
        }
        ws.onmessage = (event) => {
            if (event.data instanceof ArrayBuffer) {
                term.write(new Uint8Array(event.data))
                return
            }
            const text = String(event.data)
            let message: { type?: string; session_id?: string; status?: string; message?: string }
            try { message = JSON.parse(text) } catch { term.write(text); return }
            if (message.type === 'session_id' && message.session_id) {
                term.reset()
                onSessionId?.(message.session_id)
                return
            }
            if (message.type === 'session_state' && message.status === 'exited') {
                term.write('\r\n[Terminal exited]\r\n')
            }
            if (message.type === 'error' && message.message) term.writeln('\r\n' + message.message)
        }
        ws.onclose = () => onStatus?.('Disconnected')
        ws.onerror = () => onStatus?.('Error')
        const dataSub = term.onData((data) => {
            sendInput(data)
            onData?.(data)
        })
        const resizeSub = term.onResize(({ cols, rows }) => {
            sendJSON(terminalResize(sessionId, cols, rows))
        })
        const observer = new ResizeObserver(() => fit.fit())
        observer.observe(el)
        onReady?.({ term, send: sendInput })
        return () => {
            observer.disconnect()
            dataSub.dispose()
            resizeSub.dispose()
            sendJSON({ type: disposeRef.current === 'delete' ? 'close_delete' : 'detach_keep' })
            ws.close()
            term.dispose()
        }
    }, [url, sessionId])

    useEffect(() => {
        const el = ref.current
        if (!el) return
        el.style.display = active ? 'block' : 'none'
    }, [active])

    return (
        <div ref={(node) => setHost((current) => current === node ? current : node)} className="dotpkgs-terminal-host" style={{ display: active ? 'block' : 'none' }}>
            <div ref={ref} className="dotpkgs-terminal" />
            {active ? <TerminalKeyPicker host={host} send={(data) => sendRef.current(data)} /> : null}
        </div>
    )
}

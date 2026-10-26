import { useEffect, useRef, useState } from 'react'
import { controlChar, terminalKeys } from './terminalKeys.ts'
import './TerminalKeyPicker.css'

const buttons: { id: string; label: string; data: string }[] = [
    { id: 'tab', label: 'Tab', data: terminalKeys.tab },
    { id: 'left', label: '←', data: terminalKeys.left },
    { id: 'up', label: '↑', data: terminalKeys.up },
    { id: 'down', label: '↓', data: terminalKeys.down },
    { id: 'right', label: '→', data: terminalKeys.right },
    { id: 'esc', label: 'Esc', data: terminalKeys.esc },
    { id: 'c', label: '^C', data: terminalKeys.c },
    { id: 'a', label: '^A', data: terminalKeys.a },
    { id: 'r', label: '^R', data: terminalKeys.r },
    { id: 'l', label: '^L', data: terminalKeys.l },
]

export function TerminalKeyPicker({ host, send }: { host: HTMLElement | null; send: (data: string) => void }) {
    const bar = useRef<HTMLDivElement>(null)
    const [armed, setArmed] = useState(false)
    const [shown, setShown] = useState(false)
    const sendRef = useRef(send)
    sendRef.current = send

    useEffect(() => {
        const query = window.matchMedia('(max-width: 820px), (pointer: coarse)')
        const update = () => setShown(query.matches)
        update()
        query.addEventListener('change', update)
        return () => query.removeEventListener('change', update)
    }, [])

    useEffect(() => {
        if (!shown) return
        const place = () => {
            const el = bar.current
            if (!el || !host) return
            const vv = window.visualViewport
            const viewportHeight = vv?.height ?? window.innerHeight
            const offsetTop = vv?.offsetTop ?? 0
            const hostRect = host.getBoundingClientRect()
            const visibleBottom = offsetTop + viewportHeight
            const height = el.offsetHeight || 36
            const top = Math.min(hostRect.bottom, visibleBottom) - height - 8
            el.style.top = Math.max(hostRect.top + 8, top) + 'px'
            el.style.left = Math.max(8, hostRect.left + 8) + 'px'
            el.style.right = Math.max(8, window.innerWidth - hostRect.right + 8) + 'px'
        }
        place()
        const timer = window.setInterval(place, 250)
        window.visualViewport?.addEventListener('resize', place)
        window.visualViewport?.addEventListener('scroll', place)
        window.addEventListener('resize', place)
        return () => {
            window.clearInterval(timer)
            window.visualViewport?.removeEventListener('resize', place)
            window.visualViewport?.removeEventListener('scroll', place)
            window.removeEventListener('resize', place)
        }
    }, [shown, host])

    useEffect(() => {
        if (!armed) return
        const onKey = (event: KeyboardEvent) => {
            const data = controlChar(event.key)
            if (!data) return
            event.preventDefault()
            event.stopPropagation()
            sendRef.current(data)
            setArmed(false)
        }
        document.addEventListener('keydown', onKey, true)
        return () => document.removeEventListener('keydown', onKey, true)
    }, [armed])

    if (!shown) return null
    return (
        <div
            ref={bar}
            className="dotpkgs-terminal-keys"
            onPointerDown={(event) => event.preventDefault()}
        >
            {buttons.slice(0, 5).map((button) => (
                <button key={button.id} type="button" onClick={() => send(button.data)}>{button.label}</button>
            ))}
            <button type="button" className={armed ? 'is-on' : ''} onClick={() => setArmed((value) => !value)}>Ctrl</button>
            {buttons.slice(5).map((button) => (
                <button key={button.id} type="button" onClick={() => send(button.data)}>{button.label}</button>
            ))}
            <button
                type="button"
                onClick={() => {
                    void navigator.clipboard?.readText().then((text) => {
                        if (text) send(text)
                    }).catch(() => {})
                }}
            >Paste</button>
        </div>
    )
}

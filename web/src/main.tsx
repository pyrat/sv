import React, { useCallback, useEffect, useRef, useState } from 'react'
import { createRoot } from 'react-dom/client'
import {
  Activity,
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ArrowUpRight,
  Check,
  ChevronRight,
  CircleHelp,
  Crosshair,
  Gauge,
  Hexagon,
  Info,
  Move,
  Navigation,
  Radio,
  Settings2,
  ShieldCheck,
  Square,
  Waves,
  Wifi,
  X,
} from 'lucide-react'
import { command, type Axis, type Snapshot } from './api'
import './style.css'
import { CageContext } from './CageContext'
import { HomePanel, MotionSequence } from './HomePanel'
import { ActivityStream } from './ActivityStream'

const format = (n: number) => n.toFixed(3)
const icons = { x: ArrowRight, y: ArrowDown }

type JogControls = {
  axes: Axis[]
  active: string
  disabled: boolean
  hold: boolean
  step: number
  speed: number
  release: () => void
  startHold: (axis: string, direction: number) => Promise<void>
  send: (path: string, body: unknown) => Promise<void>
}
const JogContext = React.createContext<JogControls | null>(null)
function JogButton({
  axis,
  direction,
  label,
  children,
}: {
  axis: string
  direction: number
  label: string
  children: React.ReactNode
}) {
  const {
    axes,
    active,
    disabled,
    hold,
    step,
    speed,
    release,
    startHold,
    send,
  } = React.useContext(JogContext)!
  const a = axes.find((a) => a.id === axis)
  const isActive = active === `${axis}${direction}`
  return (
    <button
      className={`jog-key ${isActive ? 'held' : ''}`}
      aria-label={label}
      disabled={!a?.available || (disabled && !isActive)}
      onPointerDown={(e) => {
        if (hold && e.button === 0) {
          e.preventDefault()
          e.currentTarget.setPointerCapture(e.pointerId)
          void startHold(axis, direction)
        }
      }}
      onPointerUp={release}
      onPointerCancel={release}
      onLostPointerCapture={release}
      onKeyDown={(e) => {
        if (hold && (e.key === ' ' || e.key === 'Enter') && !e.repeat) {
          e.preventDefault()
          void startHold(axis, direction)
        }
      }}
      onKeyUp={(e) => {
        if (e.key === ' ' || e.key === 'Enter') release()
      }}
      onBlur={release}
      onClick={() => {
        if (!hold) void send('step', { axis, delta: direction * step, speed })
      }}
    >
      {children}
    </button>
  )
}

function App() {
  const [data, setData] = useState<Snapshot | null>(null)
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)
  const [mode, setMode] = useState<'target' | 'jog'>('target')
  const [hold, setHold] = useState(false)
  const [step, setStep] = useState(0.01)
  const [speed, setSpeed] = useState(0.1)
  const [targets, setTargets] = useState({ x: '2.450', y: '4.800' })
  const [selected, setSelected] = useState({ x: true, y: true })
  const [active, setActive] = useState('')
  const [help, setHelp] = useState(false)
  const [phoneView, setPhoneView] = useState('control')
  const [phonePlot, setPhonePlot] = useState(
    () => window.matchMedia('(max-width: 600px)').matches,
  )
  useEffect(() => {
    const media = window.matchMedia('(max-width: 600px)')
    const update = () => setPhonePlot(media.matches)
    media.addEventListener('change', update)
    update()
    return () => media.removeEventListener('change', update)
  }, [])
  const [diagnostics, setDiagnostics] = useState(false)
  const [trail, setTrail] = useState<[number, number][]>([])
  const initial = useRef(false)
  const gesture = useRef(0)
  const pressed = useRef(false)
  const lease = useRef<string | null>(null)
  const renewal = useRef<ReturnType<typeof setTimeout>>()
  const plot = useRef<SVGSVGElement>(null)
  const drag = useRef<{
    pointer: number
    x: number
    y: number
    startX: number
    startY: number
    moved: boolean
  } | null>(null)
  const [dragPreview, setDragPreview] = useState<{
    x: number
    y: number
  } | null>(null)
  const cancelDrag = useCallback(() => {
    drag.current = null
    setDragPreview(null)
  }, [])

  useEffect(() => {
    const cancelWhenHidden = () => {
      if (document.hidden) cancelDrag()
    }
    window.addEventListener('blur', cancelDrag)
    document.addEventListener('visibilitychange', cancelWhenHidden)
    return () => {
      window.removeEventListener('blur', cancelDrag)
      document.removeEventListener('visibilitychange', cancelWhenHidden)
    }
  }, [cancelDrag])

  useEffect(() => {
    let closed = false
    let timer: ReturnType<typeof setTimeout>
    const poll = async () => {
      try {
        const r = await fetch('/api/state', {
          signal: AbortSignal.timeout(1800),
          cache: 'no-store',
        })
        if (!r.ok) throw new Error('Controller unavailable')
        const snapshot: Snapshot = await r.json()
        if (closed) return
        setData(snapshot)
        setConnected(true)
        if (!initial.current) {
          setTargets({
            x: format(snapshot.axes[0].position),
            y: format(snapshot.axes[1].position),
          })
          setSelected({
            x: snapshot.axes[0].available,
            y: snapshot.axes[1].available,
          })
          setSpeed(
            Math.min(
              0.1,
              ...snapshot.axes
                .filter((a) => a.available)
                .map((a) => a.maxSpeed),
            ),
          )
          initial.current = true
        }
        const point: [number, number] = [
          snapshot.axes[0].position,
          snapshot.axes[1].position,
        ]
        setTrail((old) =>
          old.length &&
          old[old.length - 1][0] === point[0] &&
          old[old.length - 1][1] === point[1]
            ? old
            : [...old.slice(-150), point],
        )
      } catch {
        if (!closed) setConnected(false)
      }
      if (!closed) timer = setTimeout(poll, 150)
    }
    void poll()
    return () => {
      closed = true
      clearTimeout(timer)
    }
  }, [])

  const release = useCallback(() => {
    if (!pressed.current && !lease.current) return
    pressed.current = false
    gesture.current++
    lease.current = null
    clearTimeout(renewal.current)
    setActive('')
    void command('stop').catch((e) =>
      setError(`Stop not confirmed: ${e.message}`),
    )
  }, [])

  const stop = useCallback(() => {
    cancelDrag()
    pressed.current = false
    gesture.current++
    lease.current = null
    clearTimeout(renewal.current)
    setActive('')
    void command('stop').catch((e) =>
      setError(`Stop not confirmed: ${e.message}`),
    )
  }, [cancelDrag])

  useEffect(() => {
    const hidden = () => {
      if (document.hidden) release()
    }
    const escape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        stop()
        setHelp(false)
        setDiagnostics(false)
      }
    }
    window.addEventListener('blur', release)
    window.addEventListener('pointerup', release)
    window.addEventListener('pointercancel', release)
    document.addEventListener('visibilitychange', hidden)
    window.addEventListener('keydown', escape)
    return () => {
      release()
      window.removeEventListener('blur', release)
      window.removeEventListener('pointerup', release)
      window.removeEventListener('pointercancel', release)
      document.removeEventListener('visibilitychange', hidden)
      window.removeEventListener('keydown', escape)
    }
  }, [release, stop])
  useEffect(() => {
    if (!connected) release()
  }, [connected, release])

  const axes = data?.axes ?? []
  const busy = axes.some((a) =>
    ['moving', 'queued', 'jogging', 'stopping'].includes(a.state),
  )
  const fault = axes.some((a) => a.state === 'fault')
  const available = connected && !fault
  const disabled = !available || busy || pending || !!active
  useEffect(() => {
    if (disabled) cancelDrag()
  }, [disabled, cancelDrag])
  const maxSpeed = Math.min(
    ...axes.filter((a) => a.available).map((a) => a.maxSpeed),
    0.2,
  )
  const validTargets = axes
    .filter((a) => selected[a.id as 'x' | 'y'])
    .every(
      (a) =>
        targets[a.id as 'x' | 'y'].trim() !== '' &&
        Number.isFinite(Number(targets[a.id as 'x' | 'y'])) &&
        Number(targets[a.id as 'x' | 'y']) >= a.min &&
        Number(targets[a.id as 'x' | 'y']) <= a.max,
    )

  async function send(path: string, body: unknown) {
    setPending(true)
    setError('')
    try {
      await command(path, body)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setPending(false)
    }
  }
  async function startHold(axis: string, direction: number) {
    if (disabled || pressed.current) return
    pressed.current = true
    const generation = ++gesture.current
    setActive(`${axis}${direction}`)
    setError('')
    try {
      const result = await command('jog', { axis, direction, speed })
      if (!pressed.current || generation !== gesture.current) {
        await command('stop')
        return
      }
      lease.current = result.lease ?? null
      const renew = async () => {
        if (!lease.current || generation !== gesture.current) return
        try {
          await command('jog/renew', { lease: lease.current })
          if (pressed.current && generation === gesture.current)
            renewal.current = setTimeout(renew, 100)
        } catch (e) {
          setError((e as Error).message)
          release()
        }
      }
      renewal.current = setTimeout(renew, 100)
    } catch (e) {
      setError((e as Error).message)
      release()
    }
  }

  const x = axes[0],
    y = axes[1]
  const plotWidth = phonePlot ? 332 : 568
  const plotRight = 64 + plotWidth
  const px = (p: number) =>
    64 +
    ((p - (x?.min ?? 0)) / ((x?.max ?? 50) - (x?.min ?? 0) || 1)) * plotWidth
  const py = (p: number) =>
    64 + ((p - (y?.min ?? 0)) / ((y?.max ?? 30) - (y?.min ?? 0) || 1)) * 280
  const tx = Number(targets.x),
    ty = Number(targets.y)
  const canDrag = !disabled && !!x?.available && !!y?.available
  const markerX = dragPreview?.x ?? x?.position ?? 0
  const markerY = dragPreview?.y ?? y?.position ?? 0

  function updateDrag(e: React.PointerEvent<SVGSVGElement>) {
    const d = drag.current
    const matrix = plot.current?.getScreenCTM()
    if (!d || d.pointer !== e.pointerId || !matrix || !x || !y) return
    // Screen CTM handles responsive scaling and SVG letterboxing on mobile.
    const point = new DOMPoint(e.clientX, e.clientY).matrixTransform(
      matrix.inverse(),
    )
    const next = {
      x: Number(
        Math.max(
          x.min,
          Math.min(
            x.max,
            x.min + ((point.x - 64) / plotWidth) * (x.max - x.min),
          ),
        ).toFixed(3),
      ),
      y: Number(
        Math.max(
          y.min,
          Math.min(y.max, y.min + ((point.y - 64) / 280) * (y.max - y.min)),
        ).toFixed(3),
      ),
    }
    d.moved ||= Math.hypot(e.clientX - d.startX, e.clientY - d.startY) > 4
    d.x = next.x
    d.y = next.y
    if (d.moved) {
      setDragPreview(next)
      setTargets({ x: format(next.x), y: format(next.y) })
    }
  }

  function finishDrag(e: React.PointerEvent<SVGSVGElement>) {
    if (!drag.current || drag.current.pointer !== e.pointerId) return
    updateDrag(e)
    const destination = drag.current
    cancelDrag()
    if (destination?.moved && canDrag)
      void send('move', {
        targets: { x: destination.x, y: destination.y },
        speed,
      })
  }

  return (
    <JogContext.Provider
      value={{
        axes,
        active,
        disabled,
        hold,
        step,
        speed,
        release,
        startHold,
        send,
      }}
    >
      <div className="app-shell" data-phone-view={phoneView}>
        <aside className="rail">
          <a className="brand-icon" href="#" aria-label="DepthPilot home">
            <Waves size={25} />
          </a>
          <div className="rail-divider" />
          <button
            className="rail-button chosen"
            aria-label="Control station"
            title="Control station"
          >
            <Crosshair />
          </button>
          <button
            className="rail-button"
            aria-label="Axis diagnostics"
            title="Axis diagnostics"
            onClick={() => setDiagnostics(true)}
          >
            <Activity />
          </button>
          <button
            className="rail-button"
            aria-label="Configuration information"
            title="Configuration"
            onClick={() => setDiagnostics(true)}
          >
            <Settings2 />
          </button>
          <div className="rail-bottom">
            <button
              className="rail-button"
              aria-label="Help"
              onClick={() => setHelp(true)}
            >
              <CircleHelp />
            </button>
            <span className="avatar">DP</span>
          </div>
        </aside>
        <div className="workspace">
          <header>
            <div className="wordmark">
              depth<span>pilot</span>
              <span className="version"> / AQUACULTURE SYSTEMS</span>
            </div>
            <div className="header-right">
              <span className={`connection ${connected ? '' : 'lost'}`}>
                <i />
                {connected
                  ? 'Controller connected'
                  : 'Connecting to controller'}
              </span>
              <span className="header-divider" />
              <span className="device">
                <Hexagon size={14} />
                {data?.simulation ? 'VIRTUAL CONTROLLER' : 'JETSON ORIN NANO'}
              </span>
            </div>
          </header>
          <main>
            <nav className="phone-nav" aria-label="Phone views">
              {['control', 'home', 'history'].map((view) => (
                <button
                  key={view}
                  aria-current={phoneView === view ? 'page' : undefined}
                  onClick={() => setPhoneView(view)}
                >
                  {view}
                </button>
              ))}
              <button
                aria-label="Open diagnostics"
                onClick={() => setDiagnostics(true)}
              >
                <Settings2 size={18} />
              </button>
            </nav>
            <div className="phone-stop">
              <button className="stop-button" onClick={stop}>
                <Square size={16} />
                Stop all motion
              </button>
            </div>
            <div className="breadcrumb">
              SEA-CAGE OPERATIONS <ChevronRight size={12} /> CAMERA WINCH /
              PROTO 01
            </div>
            <div className="page-heading">
              <div>
                <h1>
                  Control station<span>.</span>
                </h1>
                <p>Camera winch control, from the cage edge.</p>
              </div>
              <div className={`mode-badge ${data?.simulation ? 'demo' : ''}`}>
                <span className="pulse-dot" />
                {data
                  ? data.simulation
                    ? 'SIMULATION MODE'
                    : 'HARDWARE MODE'
                  : 'ESTABLISHING LINK'}
              </div>
            </div>
            {data?.simulation && (
              <div className="sim-banner">
                <div>
                  <span className="sim-icon">
                    <Waves size={16} />
                  </span>
                  <strong>Your virtual winch is ready.</strong>
                  <span>
                    Demo travel: X 50 m × Y 30 m. No hardware connected.
                  </span>
                </div>
                <button onClick={() => setHelp(true)}>
                  About this environment <ArrowUpRight size={14} />
                </button>
              </div>
            )}
            {error && (
              <div className="error-banner" role="alert">
                <Info size={18} />
                <span>{error}</span>
                <button aria-label="Dismiss error" onClick={() => setError('')}>
                  <X size={16} />
                </button>
              </div>
            )}
            {!connected && data && (
              <div className="error-banner" role="alert">
                <Wifi size={18} />
                Connection lost. Positions are stale; movement controls are
                unavailable.
              </div>
            )}
            <CageContext
              simulation={!!data?.simulation}
              status={
                !connected
                  ? 'AWAITING TELEMETRY'
                  : fault
                    ? 'DRIVE FAULT'
                    : !axes.some((a) => a.available)
                      ? 'AXES NOT CONFIGURED'
                      : busy
                        ? 'WINCHES IN MOTION'
                        : 'WINCHES AT REST'
              }
            />
            <div className="dashboard-grid">
              <section className="visual-column">
                <div className="panel plot-panel">
                  <div className="panel-heading">
                    <div>
                      <span className="eyebrow">WINCH OPERATING SPACE</span>
                      <h2>Live position</h2>
                    </div>
                    <span className={`live-pill ${!connected ? 'stale' : ''}`}>
                      <Radio size={12} />
                      {connected ? 'LIVE' : 'OFFLINE'}
                    </span>
                  </div>
                  <div className="plot-meta">
                    <span>
                      <i className="legend-dot" />
                      Actual position
                    </span>
                    <span>
                      <i className="legend-target" />
                      Target position
                    </span>
                    {data?.home?.usable && (
                      <span>
                        <i className="legend-home" />
                        Home
                      </span>
                    )}
                    <span className="plot-units">METRES / m</span>
                  </div>
                  <div
                    className={`drag-hint ${dragPreview ? 'dragging' : ''}`}
                    aria-live="polite"
                  >
                    <Move size={14} />
                    <span>
                      {dragPreview
                        ? `Release to move both winches · X then Y · ${Math.round(speed * 1000)} mm/s`
                        : canDrag
                          ? 'Drag the glow. Release to move.'
                          : busy
                            ? 'Winches moving · drag available when at rest'
                            : 'Drag requires two available, connected axes'}
                    </span>
                    <span className="drag-hint-detail">
                      MOTOR AXES · NOT CAMERA COORDINATES
                    </span>
                  </div>
                  <div className="plot-wrap">
                    <svg
                      ref={plot}
                      className={`position-plot ${dragPreview ? 'is-dragging' : ''}`}
                      viewBox={`0 0 ${plotRight + 64} 397`}
                      role="group"
                      aria-label={`Live axis position: X ${x ? format(x.position) : 'unknown'} metres, Y ${y ? format(y.position) : 'unknown'} metres`}
                      onPointerMove={updateDrag}
                      onPointerUp={finishDrag}
                      onPointerCancel={cancelDrag}
                      onLostPointerCapture={cancelDrag}
                    >
                      <defs>
                        <pattern
                          id="minor-grid"
                          width="28.4"
                          height="28"
                          patternUnits="userSpaceOnUse"
                          x="64"
                          y="64"
                        >
                          <path
                            d="M 28.4 0 L 0 0 0 28"
                            fill="none"
                            stroke="#233438"
                            strokeWidth="0.5"
                          />
                        </pattern>
                        <radialGradient id="position-glow">
                          <stop stopColor="#53e9cb" stopOpacity=".18" />
                          <stop
                            offset="1"
                            stopColor="#53e9cb"
                            stopOpacity="0"
                          />
                        </radialGradient>
                        <linearGradient id="water-column" x2="0" y2="1">
                          <stop stopColor="#255254" stopOpacity=".22" />
                          <stop
                            offset="1"
                            stopColor="#0a202e"
                            stopOpacity=".4"
                          />
                        </linearGradient>
                      </defs>
                      <rect
                        x="64"
                        y="64"
                        width={plotWidth}
                        height="280"
                        fill="url(#water-column)"
                      />
                      <rect
                        x="64"
                        y="64"
                        width={plotWidth}
                        height="280"
                        fill="url(#minor-grid)"
                      />
                      {[0, 1, 2, 3, 4, 5].map((i) => (
                        <g key={i}>
                          <line
                            x1={64 + (i * plotWidth) / 5}
                            y1="64"
                            x2={64 + (i * plotWidth) / 5}
                            y2="344"
                            stroke="#2c3e41"
                            strokeWidth=".7"
                          />
                          <line
                            x1="64"
                            y1={64 + i * 56}
                            x2={plotRight}
                            y2={64 + i * 56}
                            stroke="#2c3e41"
                            strokeWidth=".7"
                          />
                          <text
                            x={64 + (i * plotWidth) / 5}
                            y="55"
                            textAnchor="middle"
                            className="axis-label"
                          >
                            {(
                              (x?.min ?? 0) +
                              (i * ((x?.max ?? 50) - (x?.min ?? 0))) / 5
                            ).toFixed(1)}
                          </text>
                          <text
                            x="48"
                            y={68 + i * 56}
                            textAnchor="end"
                            className="axis-label"
                          >
                            {(
                              (y?.min ?? 0) +
                              (i * ((y?.max ?? 30) - (y?.min ?? 0))) / 5
                            ).toFixed(1)}
                          </text>
                        </g>
                      ))}
                      <text x="64" y="43" className="axis-name">
                        Y / DOWN ↓
                      </text>
                      <text
                        x={plotRight}
                        y="390"
                        textAnchor="end"
                        className="axis-name"
                      >
                        X / ALONG →
                      </text>
                      <polyline
                        points={trail
                          .map((p) => `${px(p[0])},${py(p[1])}`)
                          .join(' ')}
                        stroke="#54d9bd"
                        strokeWidth="1.5"
                        opacity=".4"
                        fill="none"
                      />
                      {validTargets &&
                        Number.isFinite(tx) &&
                        Number.isFinite(ty) && (
                          <g transform={`translate(${px(tx)},${py(ty)})`}>
                            <path
                              d="M-7 0H7M0-7V7"
                              stroke="#8aaba9"
                              strokeWidth="1.2"
                            />
                            <circle
                              r="11"
                              fill="none"
                              stroke="#8aaba9"
                              strokeDasharray="2 4"
                            />
                          </g>
                        )}
                      {data?.home?.usable &&
                        data.home.saved?.axes.x &&
                        data.home.saved.axes.y && (
                          <g
                            className="home-plot-marker"
                            transform={`translate(${px(data.home.saved.axes.x.position)},${py(data.home.saved.axes.y.position)})`}
                            aria-label="Saved home position"
                          >
                            <path
                              d="M0-9L9 0L0 9L-9 0Z"
                              fill="#2e3225"
                              stroke="#d0c28b"
                            />
                            <path
                              d="M-4 1V-2L0-5L4-2V4H-4Z"
                              fill="none"
                              stroke="#dfd3a4"
                              strokeWidth=".8"
                            />
                            <text x="14" y="4" className="home-plot-label">
                              HOME
                            </text>
                          </g>
                        )}
                      {x?.available && y?.available && (
                        <>
                          {dragPreview && (
                            <g className="actual-position-ghost">
                              <line
                                x1={px(x.position)}
                                y1={py(y.position)}
                                x2={px(markerX)}
                                y2={py(markerY)}
                                stroke="#76b6a2"
                                strokeDasharray="4 5"
                                opacity=".6"
                              />
                              <circle
                                cx={px(x.position)}
                                cy={py(y.position)}
                                r="6"
                                fill="#a7e5d2"
                              />
                              <text
                                x={px(x.position) + 12}
                                y={py(y.position) + 20}
                                className="axis-name"
                              >
                                ACTUAL
                              </text>
                            </g>
                          )}
                          <g
                            className={`live-marker ${canDrag ? 'draggable' : ''} ${dragPreview ? 'dragging' : ''}`}
                            style={{
                              transform: `translate(${px(markerX)}px, ${py(markerY)}px)`,
                            }}
                            role="button"
                            aria-label="Drag winch position to set X and Y targets"
                            aria-disabled={!canDrag}
                            aria-describedby="plot-keyboard-help"
                            tabIndex={canDrag ? 0 : -1}
                            onPointerDown={(e) => {
                              if (!canDrag || e.button !== 0 || drag.current)
                                return
                              e.preventDefault()
                              plot.current?.setPointerCapture(e.pointerId)
                              drag.current = {
                                pointer: e.pointerId,
                                x: x.position,
                                y: y.position,
                                startX: e.clientX,
                                startY: e.clientY,
                                moved: false,
                              }
                              setDragPreview({ x: x.position, y: y.position })
                              setMode('target')
                              setSelected({ x: true, y: true })
                            }}
                            onKeyDown={(e) => {
                              if (!canDrag) return
                              if (e.key === 'Enter' && validTargets) {
                                e.preventDefault()
                                setSelected({ x: true, y: true })
                                void send('move', {
                                  targets: { x: tx, y: ty },
                                  speed,
                                })
                              }
                              if (
                                [
                                  'ArrowUp',
                                  'ArrowDown',
                                  'ArrowLeft',
                                  'ArrowRight',
                                ].includes(e.key)
                              ) {
                                e.preventDefault()
                                setMode('target')
                                setSelected({ x: true, y: true })
                                setTargets({
                                  x: format(
                                    Math.max(
                                      x.min,
                                      Math.min(
                                        x.max,
                                        (Number.isFinite(tx)
                                          ? tx
                                          : x.position) +
                                          (e.key === 'ArrowRight'
                                            ? 0.01
                                            : e.key === 'ArrowLeft'
                                              ? -0.01
                                              : 0),
                                      ),
                                    ),
                                  ),
                                  y: format(
                                    Math.max(
                                      y.min,
                                      Math.min(
                                        y.max,
                                        (Number.isFinite(ty)
                                          ? ty
                                          : y.position) +
                                          (e.key === 'ArrowUp'
                                            ? -0.01
                                            : e.key === 'ArrowDown'
                                              ? 0.01
                                              : 0),
                                      ),
                                    ),
                                  ),
                                })
                              }
                            }}
                          >
                            <circle r="64" fill="url(#position-glow)" />
                            <path
                              d={`M${64 - px(markerX)} 0H${plotRight - px(markerX)}M0 ${64 - py(markerY)}V${344 - py(markerY)}`}
                              stroke="#52d9bd"
                              strokeDasharray="3 5"
                              opacity=".3"
                            />
                            <circle
                              r="13"
                              fill="#123b34"
                              stroke="#50d5b8"
                              opacity=".7"
                            />
                            <circle r="5" fill="#67f0cf" />
                            <circle r="2" fill="#ecfff9" />
                            <circle
                              className="drag-ring"
                              r="21"
                              fill="none"
                              stroke="#72dcbc"
                              strokeDasharray="2 5"
                              opacity=".45"
                            />
                            <circle
                              className="drag-hit-area"
                              data-testid="position-drag-handle"
                              r="30"
                              fill="transparent"
                            />
                            <g
                              transform={`translate(${markerX > (x.min + x.max) / 2 ? -145 : 25},${markerY < y.min + (y.max - y.min) * 0.15 ? 26 : -38})`}
                            >
                              <rect
                                width="128"
                                height="25"
                                rx="4"
                                fill="#16352f"
                                stroke="#2d574b"
                              />
                              <text x="9" y="16" className="coordinate-label">
                                {format(markerX)} , {format(markerY)}
                              </text>
                            </g>
                          </g>
                        </>
                      )}
                    </svg>
                    <span id="plot-keyboard-help" className="sr-only">
                      Drag to preview, release to move X then Y. Arrow keys
                      adjust targets by 10 millimetres; Enter moves. Escape
                      cancels and stops.
                    </span>
                  </div>
                  <div className="plot-footer">
                    <span>
                      <Crosshair size={13} />X → along · Y ↓ down · metres
                    </span>
                    <span>
                      {busy ? 'Motion in progress' : 'Holding position'}
                      <i className={busy ? 'moving-dot' : ''} />
                    </span>
                  </div>
                </div>
                <div className="axis-cards">
                  {['x', 'y'].map((id, i) => {
                    const a = axes[i]
                    const Icon = icons[id as 'x' | 'y']
                    return (
                      <div className={`panel axis-card ${id}`} key={id}>
                        <div className="axis-card-top">
                          <span className="axis-symbol">
                            <Icon size={16} />
                          </span>
                          <h3>{id.toUpperCase()} axis</h3>
                          <span
                            className={`state-label ${a?.state === 'fault' ? 'fault' : ''}`}
                          >
                            {!connected ? 'OFFLINE' : a?.state.toUpperCase()}
                          </span>
                        </div>
                        <div className="position-reading">
                          <span data-testid={`${id}-position`}>
                            {a?.available ? format(a.position) : '—.———'}
                          </span>
                          <span>m</span>
                        </div>
                        <div className="axis-progress">
                          <i
                            style={{
                              width: `${a ? Math.max(0, Math.min(100, ((a.position - a.min) / (a.max - a.min)) * 100)) : 0}%`,
                            }}
                          />
                        </div>
                        <div className="axis-card-bottom">
                          <span>{a?.min.toFixed(1) ?? '0.0'} m</span>
                          <span>
                            {a?.available
                              ? a.error || 'Actual position'
                              : 'Not configured'}
                          </span>
                          <span>{a?.max.toFixed(1) ?? '—'} m</span>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </section>
              <section className="panel controls">
                <div className="panel-heading">
                  <div>
                    <span className="eyebrow">MOTION CONTROL</span>
                    <h2>Make your move</h2>
                  </div>
                  <Move size={20} className="muted" />
                </div>
                <div className="tabs" role="tablist" aria-label="Movement mode">
                  <button
                    role="tab"
                    aria-selected={mode === 'target'}
                    onClick={() => {
                      release()
                      setMode('target')
                    }}
                  >
                    <Crosshair size={15} />
                    Target position
                  </button>
                  <button
                    role="tab"
                    aria-selected={mode === 'jog'}
                    onClick={() => {
                      release()
                      setMode('jog')
                    }}
                  >
                    <Navigation size={15} />
                    Manual jog
                  </button>
                </div>
                <div className="control-body">
                  <MotionSequence motion={data?.motion} />
                  {mode === 'target' ? (
                    <div className="target-form">
                      <div className="section-label">
                        DESTINATION
                        <button
                          onClick={() => {
                            if (x && y)
                              setTargets({
                                x: format(x.position),
                                y: format(y.position),
                              })
                          }}
                          disabled={!connected}
                        >
                          Use current <ArrowUpRight size={12} />
                        </button>
                      </div>
                      {['x', 'y'].map((id, i) => (
                        <div className="target-row" key={id}>
                          <label className={`axis-tag ${id}`}>
                            <input
                              type="checkbox"
                              checked={selected[id as 'x' | 'y']}
                              disabled={!axes[i]?.available || busy}
                              onChange={(e) =>
                                setSelected({
                                  ...selected,
                                  [id]: e.target.checked,
                                })
                              }
                            />
                            {id.toUpperCase()}
                          </label>
                          <label className="number-field">
                            <span className="sr-only">
                              {id.toUpperCase()} target metres
                            </span>
                            <input
                              type="number"
                              step="0.001"
                              min={axes[i]?.min}
                              max={axes[i]?.max}
                              value={targets[id as 'x' | 'y']}
                              disabled={!axes[i]?.available || busy}
                              onChange={(e) =>
                                setTargets({ ...targets, [id]: e.target.value })
                              }
                            />
                            <span>m</span>
                          </label>
                        </div>
                      ))}
                      <p className="field-note">
                        Select the axes to move. Targets use the established
                        motor reference.
                      </p>
                      {!validTargets && (
                        <p className="validation" role="alert">
                          Enter targets within the configured travel range.
                        </p>
                      )}
                    </div>
                  ) : (
                    <div className="jog-form">
                      <div className="jog-options">
                        <button
                          className={!hold ? 'selected' : ''}
                          onClick={() => {
                            release()
                            setHold(false)
                          }}
                        >
                          Step
                        </button>
                        <button
                          className={hold ? 'selected' : ''}
                          onClick={() => {
                            release()
                            setHold(true)
                          }}
                        >
                          Hold to move
                        </button>
                        <select
                          aria-label="Jog step"
                          disabled={hold}
                          value={step}
                          onChange={(e) => setStep(Number(e.target.value))}
                        >
                          <option value="0.001">1 mm</option>
                          <option value="0.01">10 mm</option>
                          <option value="0.1">100 mm</option>
                        </select>
                      </div>
                      <div className="jog-pad">
                        <div className="up">
                          <JogButton
                            axis="y"
                            direction={-1}
                            label="Jog Y negative"
                          >
                            <ArrowUp size={20} />
                            <span>Y−</span>
                          </JogButton>
                        </div>
                        <div className="left">
                          <JogButton
                            axis="x"
                            direction={-1}
                            label="Jog X negative"
                          >
                            <ArrowLeft size={20} />
                            <span>X−</span>
                          </JogButton>
                        </div>
                        <div className="jog-center">
                          <Crosshair size={23} />
                        </div>
                        <div className="right">
                          <JogButton
                            axis="x"
                            direction={1}
                            label="Jog X positive"
                          >
                            <ArrowRight size={20} />
                            <span>X+</span>
                          </JogButton>
                        </div>
                        <div className="down">
                          <JogButton
                            axis="y"
                            direction={1}
                            label="Jog Y positive"
                          >
                            <ArrowDown size={20} />
                            <span>Y+</span>
                          </JogButton>
                        </div>
                      </div>
                      <p className="field-note centered">
                        {hold
                          ? 'Press and hold a direction. Release to stop.'
                          : 'One tap. One precise step.'}
                      </p>
                    </div>
                  )}
                  <div className="speed-control">
                    <div className="section-label">
                      <span>TRAVEL SPEED</span>
                      <span className="speed-value">
                        {Math.round(speed * 1000)}
                        <span>mm/s</span>
                      </span>
                    </div>
                    <input
                      aria-label="Travel speed"
                      type="range"
                      min="0.005"
                      max={maxSpeed || 0.2}
                      step="0.005"
                      value={speed}
                      disabled={busy || !!active}
                      onChange={(e) => setSpeed(Number(e.target.value))}
                      style={
                        {
                          '--range': `${(speed / (maxSpeed || 0.2)) * 100}%`,
                        } as React.CSSProperties
                      }
                    />
                    <div className="range-labels">
                      <span>PRECISION</span>
                      <span>CRUISE</span>
                    </div>
                  </div>
                  {mode === 'target' && (
                    <button
                      className="move-button"
                      disabled={
                        disabled ||
                        !validTargets ||
                        (!selected.x && !selected.y)
                      }
                      onClick={() =>
                        void send('move', {
                          targets: Object.fromEntries(
                            Object.entries(targets)
                              .filter(([id]) => selected[id as 'x' | 'y'])
                              .map(([id, v]) => [id, Number(v)]),
                          ),
                          speed,
                        })
                      }
                    >
                      <ArrowUpRight size={18} />
                      {pending
                        ? 'Sending command…'
                        : busy
                          ? 'Moving to position…'
                          : 'Move to position'}
                      <span>GO</span>
                    </button>
                  )}
                  <button className="stop-button" onClick={stop}>
                    <Square size={13} fill="currentColor" />
                    Stop all motion<kbd>ESC</kbd>
                  </button>
                  <p className="stop-note">
                    Controlled stop · motor holding torque retained
                  </p>
                </div>
                <div className="control-footer">
                  <ShieldCheck size={15} />
                  {data?.simulation
                    ? 'Simulated movement · hardware-free testing'
                    : 'Drive-controlled brake sequencing'}
                </div>
              </section>
              <HomePanel
                home={data?.home}
                axes={axes}
                disabled={disabled || !!dragPreview}
                speed={speed}
                send={send}
              />
              <ActivityStream
                events={data?.events ?? []}
                storage={data?.activity}
              />
            </div>
            <footer>
              <span>
                <Waves size={14} />
                DEPTHPILOT<span className="footer-dot">·</span>ENGINEERED FOR
                THE DEEP
              </span>
              <span>
                <i className="tiny-dot" />{' '}
                {connected ? 'Telemetry streaming' : 'Awaiting telemetry'}
                <span className="footer-dot">/</span>PROTO 01
              </span>
            </footer>
          </main>
        </div>
        {(help || diagnostics) && (
          <div
            className="modal-backdrop"
            onClick={() => {
              setHelp(false)
              setDiagnostics(false)
            }}
          >
            <section
              role="dialog"
              aria-modal="true"
              aria-labelledby="dialog-title"
              className="panel modal"
              onClick={(e) => e.stopPropagation()}
            >
              <button
                className="modal-close"
                autoFocus
                aria-label="Close dialog"
                onClick={() => {
                  setHelp(false)
                  setDiagnostics(false)
                }}
              >
                <X />
              </button>
              <span className="eyebrow">DEPTHPILOT / OPERATOR NOTES</span>
              <h2 id="dialog-title">
                {help
                  ? 'A real interface. A virtual winch.'
                  : 'Axis configuration'}
              </h2>
              {help ? (
                <>
                  <p>
                    Simulation uses the same Go API as the physical controller.
                    Move both axes, watch live positions, and test target moves
                    and jogging without a winch connected.
                  </p>
                  <p>
                    X and Y are independent motor axes, not camera coordinates.
                    The plot uses a top-left origin: X increases right along the
                    cage, and Y increases downward into the water. Demo travel
                    is X 0–50 m / Y 0–30 m at up to 200 mm/s. Hardware calibration is
                    supplied on the Jetson.
                  </p>
                  <p>
                    Outbound targets move X first, then Y. Return home moves Y
                    first, then X. Set current as home saves the stationary
                    drive positions without resetting counters.
                  </p>
                  <p>
                    Hold-to-jog uses a short-lived command lease. Releasing the
                    button, leaving the window, or losing fresh commands
                    requests a stop. Press Escape to stop all motion.
                  </p>
                </>
              ) : (
                <>
                  <p>
                    {data?.simulation
                      ? 'Demo calibration is active. These values do not describe a physical winch.'
                      : 'Calibration is loaded from the Jetson configuration file. Unconfigured axes cannot move.'}
                  </p>
                  {axes.map((a) => (
                    <div className="diagnostic-row" key={a.id}>
                      <strong>{a.id.toUpperCase()} axis</strong>
                      <span>
                        {a.available
                          ? `${a.min}–${a.max} m · ${Math.round(a.maxSpeed * 1000)} mm/s max`
                          : 'Not configured'}
                      </span>
                      {a.error && <small>{a.error}</small>}
                      {a.available && (
                        <div className="counter-readings">
                          <span>
                            Profile / command{' '}
                            <b>{a.profileCounter?.toLocaleString() ?? '—'}</b>
                          </span>
                          <span>
                            Encoder feedback{' '}
                            <b>{a.feedbackCounter?.toLocaleString() ?? '—'}</b>
                          </span>
                        </div>
                      )}
                    </div>
                  ))}
                  <p>
                    Metre conversion uses separate command-pulse and
                    encoder-feedback calibration. Changes require a service
                    restart. This prototype does not perform coordinated XY
                    interpolation.
                  </p>
                </>
              )}
              <button
                className="move-button"
                onClick={() => {
                  setHelp(false)
                  setDiagnostics(false)
                }}
              >
                Back to control station
                <ArrowRight size={16} />
              </button>
            </section>
          </div>
        )}
      </div>
    </JogContext.Provider>
  )
}

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)

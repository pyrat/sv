import { useState } from 'react'
import { ArrowRight, Home, MapPin, Pencil, Save, X } from 'lucide-react'
import type { Axis, HomeStatus, Motion } from './api'

type Props = {
  home?: HomeStatus
  axes: Axis[]
  disabled: boolean
  speed: number
  send: (path: string, body: unknown) => Promise<void>
}

export function HomePanel({ home, axes, disabled, speed, send }: Props) {
  const [editing, setEditing] = useState(false)
  const [units, setUnits] = useState<'positions' | 'counters'>('positions')
  const [draft, setDraft] = useState<Record<string, string>>({})
  const configured = axes.filter((a) => a.available)
  const saved = home?.saved
  const valid =
    configured.length > 0 &&
    configured.every((a) => {
      const value = Number(draft[a.id])
      return (
        draft[a.id]?.trim() &&
        Number.isFinite(value) &&
        (units === 'positions'
          ? value >= a.min && value <= a.max
          : Number.isInteger(value) &&
            value >= -2147483648 &&
            value <= 2147483647)
      )
    })
  function edit(next: 'positions' | 'counters') {
    setUnits(next)
    setDraft(
      Object.fromEntries(
        configured.map((a) => [
          a.id,
          String(
            next === 'positions'
              ? (saved?.axes[a.id]?.position ?? a.position)
              : (saved?.axes[a.id]?.commandCounter ?? a.profileCounter ?? 0),
          ),
        ]),
      ),
    )
    setEditing(true)
  }
  return (
    <section className="panel home-panel" aria-label="Home position">
      <div className="home-heading">
        <span className="home-icon">
          <Home size={20} />
        </span>
        <div>
          <span className="eyebrow">RETURN TO THE CAGE EDGE</span>
          <h2>Home position</h2>
        </div>
        <span
          className={`home-badge ${home?.atHome ? 'at-home' : ''}`}
          data-testid="home-status"
        >
          {!saved
            ? 'NOT SET'
            : !home?.usable
              ? 'REFERENCE REQUIRED'
              : home.atHome
                ? 'AT HOME'
                : 'HOME SAVED'}
        </span>
      </div>
      <div className="home-content">
        <div className="home-coordinates">
          {axes.map((a) => {
            const value = saved?.axes[a.id]
            return (
              <div className={`home-coordinate ${a.id}`} key={a.id}>
                <span>
                  {a.id.toUpperCase()} / WINCH {a.id === 'x' ? '01' : '02'}
                </span>
                <strong data-testid={`home-${a.id}`}>
                  {value ? value.position.toFixed(3) : '—.———'} <small>m</small>
                </strong>
                <span>
                  Command counter{' '}
                  <b>{value?.commandCounter.toLocaleString() ?? '—'}</b>
                </span>
                {value?.feedbackCounter !== undefined && (
                  <span>
                    Captured feedback{' '}
                    <b>{value.feedbackCounter.toLocaleString()}</b>
                  </span>
                )}
              </div>
            )
          })}
        </div>
        <div className="home-actions">
          <button
            className="move-button return-home"
            disabled={disabled || !home?.usable || home?.atHome}
            onClick={() => void send('home/return', { speed })}
          >
            <Home size={16} />
            Return home<span>Y → X</span>
          </button>
          <div className="home-secondary">
            <button
              disabled={disabled || !configured.length}
              onClick={() => void send('home/set', { source: 'current' })}
            >
              <MapPin size={14} />
              Set current as home
            </button>
            <button
              disabled={disabled || !configured.length}
              aria-label="Edit home position"
              onClick={() => edit(units)}
            >
              <Pencil size={14} />
            </button>
          </div>
        </div>
      </div>
      <div className="home-caption">
        <span>
          {saved
            ? `Saved ${new Date(saved.savedAt).toLocaleString()} · ${saved.source === 'current' ? 'captured from stationary drives' : `set using ${saved.source}`}`
            : 'Save a parking position for the configured winches. Counters are not reset.'}
        </span>
        {saved && (
          <button
            disabled={disabled}
            onClick={() => void send('home/clear', {})}
          >
            Clear home
          </button>
        )}
      </div>
      {saved && !home?.usable && (
        <p className="home-warning" role="status">
          {home?.reason}
        </p>
      )}
      {editing && (
        <div className="home-editor">
          <div className="section-label">
            <span>SET HOME DESTINATION</span>
            <button
              aria-label="Close home editor"
              onClick={() => setEditing(false)}
            >
              <X size={16} />
            </button>
          </div>
          <div className="home-editor-tabs">
            <button
              className={units === 'positions' ? 'selected' : ''}
              onClick={() => edit('positions')}
            >
              Metres
            </button>
            <button
              className={units === 'counters' ? 'selected' : ''}
              onClick={() => edit('counters')}
            >
              Command counters
            </button>
          </div>
          <div className="home-editor-fields">
            {configured.map((a) => (
              <label key={a.id}>
                <span>
                  {a.id.toUpperCase()} home{' '}
                  {units === 'positions' ? '(m)' : 'command counter'}
                </span>
                <input
                  aria-label={`${a.id.toUpperCase()} home ${units === 'positions' ? 'metres' : 'command counter'}`}
                  type="number"
                  step={units === 'positions' ? '0.001' : '1'}
                  value={draft[a.id] ?? ''}
                  disabled={disabled}
                  onChange={(e) =>
                    setDraft({ ...draft, [a.id]: e.target.value })
                  }
                />
              </label>
            ))}
            <button
              className="move-button"
              disabled={disabled || !valid}
              onClick={() =>
                void send('home/set', {
                  source: units,
                  [units === 'positions' ? 'targets' : 'counters']:
                    Object.fromEntries(
                      configured.map((a) => [a.id, Number(draft[a.id])]),
                    ),
                })
              }
            >
              <Save size={15} />
              Save home
            </button>
          </div>
          <p className="field-note">
            Changes the saved destination only. Raw command counts and
            encoder-feedback counts use separate calibration; this does not zero
            either counter.
          </p>
        </div>
      )}
    </section>
  )
}

export function MotionSequence({ motion }: { motion?: Motion | null }) {
  const running = motion?.status === 'running'
  return (
    <div
      className={`motion-sequence ${running ? 'running' : ''}`}
      data-testid="motion-status"
      aria-live="polite"
    >
      <div className="sequence-label">
        <span>
          {running
            ? motion.kind === 'home'
              ? 'RETURNING HOME'
              : 'OUTBOUND MOVE'
            : motion?.status === 'cancelled'
              ? 'SEQUENCE STOPPED'
              : 'TRAVEL SEQUENCE'}
        </span>
        <span>
          {running
            ? `${motion.stage + 1} / ${motion.order.length}`
            : 'ONE AXIS AT A TIME'}
        </span>
      </div>
      {running ? (
        <div className="sequence-stages">
          {motion.order.map((id, i) => (
            <span
              key={id}
              className={
                i < motion.stage
                  ? 'done'
                  : id === motion.activeAxis
                    ? 'current'
                    : ''
              }
            >
              {i > 0 && <ArrowRight size={12} />}
              <b>{id.toUpperCase()}</b>
              {i < motion.stage
                ? 'Complete'
                : id === motion.activeAxis
                  ? 'Moving'
                  : 'Waiting'}
            </span>
          ))}
        </div>
      ) : (
        <div className="sequence-idle">
          <span>
            Out <b>X → Y</b>
          </span>
          <span>
            Home <b>Y → X</b>
          </span>
        </div>
      )}
      {motion?.error && <p className="home-warning">{motion.error}</p>}
    </div>
  )
}

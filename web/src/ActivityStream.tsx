import { useEffect, useRef, useState } from 'react'
import {
  Activity,
  ArrowDown,
  Check,
  Clock3,
  Database,
  History,
  RefreshCw,
  X,
} from 'lucide-react'
import type { ActivityEvent, ActivityPage, ActivityStorage } from './api'

function eventDate(time: string, full = false) {
  const date = new Date(time)
  const sameDay = date.toDateString() === new Date().toDateString()
  return !full && sameDay
    ? date.toLocaleTimeString('en-GB')
    : date.toLocaleString('en-GB', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
}

function storageProblem(storage?: ActivityStorage) {
  if (!storage) return ''
  if (storage.error)
    return `Activity storage error: ${storage.error}. Entries awaiting storage: ${storage.queued}.`
  if (storage.dropped)
    return `${storage.dropped} activity entries could not be queued for storage during this session.`
  return ''
}

export function ActivityStream({
  events,
  storage,
}: {
  events: ActivityEvent[]
  storage?: ActivityStorage
}) {
  const [history, setHistory] = useState(false)
  const problem = storageProblem(storage)
  return (
    <section className="panel activity-panel">
      <div className="activity-header">
        <div>
          <Activity size={16} />
          <h2>Activity stream</h2>
          <span className="count">{storage?.retained ?? events.length}</span>
        </div>
        <div className="activity-tools">
          <span className={`activity-storage ${problem ? 'degraded' : ''}`}>
            <Database size={12} />
            {problem
              ? 'STORAGE ISSUE'
              : !storage?.persistent
                ? 'SESSION ONLY'
                : storage.queued
                  ? 'SAVING…'
                  : 'PERSISTED'}
          </span>
          <button className="history-button" onClick={() => setHistory(true)}>
            <History size={13} />
            View history
          </button>
        </div>
      </div>
      {problem && (
        <p className="activity-warning" role="status">
          {problem}
        </p>
      )}
      <div className="activity-list">
        {events.slice(0, 4).map((event, i) => (
          <div className="event" key={event.id ?? `${event.time}-${i}`}>
            <span className="event-icon">
              <Check size={12} />
            </span>
            <span title={event.text}>{event.text}</span>
            <time dateTime={event.time} title={eventDate(event.time, true)}>
              {eventDate(event.time)}
            </time>
          </div>
        ))}
        {events.length === 0 && (
          <div className="event muted">Waiting for controller activity…</div>
        )}
      </div>
      {history && <ActivityHistory onClose={() => setHistory(false)} />}
    </section>
  )
}

function ActivityHistory({ onClose }: { onClose: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [page, setPage] = useState<ActivityPage | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const request = useRef<AbortController | null>(null)
  const active = useRef(true)

  async function load(before = 0) {
    request.current?.abort()
    const controller = new AbortController()
    request.current = controller
    const timeout = setTimeout(() => controller.abort(), 4000)
    setLoading(true)
    setError('')
    try {
      const response = await fetch(
        `/api/activity?limit=50${before ? `&before=${before}` : ''}`,
        { signal: controller.signal, cache: 'no-store' },
      )
      const data = await response.json()
      if (!response.ok)
        throw new Error(data.error || 'Could not load activity history')
      if (!active.current || request.current !== controller) return
      const next: ActivityPage = data
      setPage((previous) =>
        before && previous
          ? { ...next, events: [...previous.events, ...next.events] }
          : next,
      )
    } catch (e) {
      if (active.current && request.current === controller)
        setError(
          (e as Error).name === 'AbortError'
            ? 'History request timed out. Try refreshing.'
            : (e as Error).message,
        )
    } finally {
      clearTimeout(timeout)
      if (active.current && request.current === controller) setLoading(false)
    }
  }

  useEffect(() => {
    active.current = true
    dialog.current?.showModal()
    void load()
    return () => {
      active.current = false
      request.current?.abort()
      dialog.current?.close()
    }
  }, [])

  return (
    <dialog
      ref={dialog}
      className="panel activity-history"
      aria-labelledby="history-title"
      onCancel={onClose}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className="history-inner">
        <div className="history-heading">
          <span className="history-symbol">
            <History size={22} />
          </span>
          <div>
            <span className="eyebrow">DEPTHPILOT / OPERATIONS LOG</span>
            <h2 id="history-title">Activity history</h2>
          </div>
          <button
            aria-label="Close activity history"
            className="history-close"
            onClick={onClose}
          >
            <X size={20} />
          </button>
        </div>
        <div className="history-summary">
          <span>
            <Database size={13} />
            {page
              ? `${page.storage.retained.toLocaleString()} retained · ${page.storage.persistent ? 'saved on the controller' : 'memory only'}`
              : 'Loading controller history…'}
          </span>
          <button
            className="history-button"
            onClick={() => void load()}
            disabled={loading}
          >
            <RefreshCw size={13} />
            Refresh history
          </button>
        </div>
        {page && (
          <p className="history-retention">
            Newest first · latest {page.storage.limit.toLocaleString()} entries
            retained
            {page.storage.queued
              ? ` · ${page.storage.queued} awaiting storage`
              : ''}
          </p>
        )}
        {error && (
          <p className="activity-warning" role="alert">
            {error}
          </p>
        )}
        {storageProblem(page?.storage) && (
          <p className="activity-warning" role="status">
            {storageProblem(page?.storage)}
          </p>
        )}
        <div className="history-events" aria-busy={loading}>
          {page?.events.map((event) => (
            <article className="history-event" key={event.id}>
              <span className="history-event-icon">
                <Clock3 size={14} />
              </span>
              <div>
                <p>{event.text}</p>
                <div className="history-event-meta">
                  <time dateTime={event.time}>
                    {eventDate(event.time, true)}
                  </time>
                  <span>#{event.id}</span>
                  {event.session && (
                    <span title={event.session}>
                      Session {event.session.slice(0, 8)}
                    </span>
                  )}
                </div>
              </div>
            </article>
          ))}
          {page && !page.events.length && !loading && (
            <p className="history-empty">No stored events yet.</p>
          )}
          {loading && !page && (
            <p className="history-empty">Loading activity…</p>
          )}
        </div>
        <div className="history-footer">
          <span>{page ? `${page.events.length} entries shown` : ''}</span>
          {!!page?.nextBefore && (
            <button
              className="history-button"
              disabled={loading}
              onClick={() => void load(page.nextBefore)}
            >
              <ArrowDown size={14} />
              {loading ? 'Loading…' : 'Load older activity'}
            </button>
          )}
        </div>
      </div>
    </dialog>
  )
}

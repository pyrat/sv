export type Axis = {
  id: string
  position: number
  target: number
  min: number
  max: number
  maxSpeed: number
  state: string
  error: string
  available: boolean
  updated: string
  profileCounter?: number
  feedbackCounter?: number
}
export type HomeAxis = {
  position: number
  commandCounter: number
  profileCounter?: number
  feedbackCounter?: number
}
export type HomeStatus = {
  saved: {
    savedAt: string
    source: string
    axes: Record<string, HomeAxis>
  } | null
  usable: boolean
  atHome: boolean
  reason: string
}
export type Motion = {
  kind: 'home' | 'outbound'
  order: string[]
  stage: number
  activeAxis: string
  status: string
  targets: Record<string, number>
  speed: number
  error?: string
}
export type Snapshot = {
  simulation: boolean
  axes: Axis[]
  events: ActivityEvent[]
  time: string
  home?: HomeStatus
  motion?: Motion | null
  activity?: ActivityStorage
}

export type ActivityEvent = {
  id?: number
  time: string
  text: string
  session?: string
}
export type ActivityStorage = {
  persistent: boolean
  retained: number
  limit: number
  queued: number
  dropped: number
  error?: string
}
export type ActivityPage = {
  events: ActivityEvent[]
  nextBefore: number
  storage: ActivityStorage
}

export async function command(
  path: string,
  body: unknown = {},
): Promise<{ lease?: string }> {
  const response = await fetch(`/api/${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(2500),
  })
  const data = await response.json()
  if (!response.ok)
    throw new Error(data.error || 'The controller rejected the command')
  return data
}

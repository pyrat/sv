# Unloaded two-drum commissioning

`cmd/depthpilot-rope` serves a dedicated commissioning UI and API from one Go
binary. It implements the [agreed rope mechanism](README.md#proposed-two-drum-rope-configuration)
without changing the legacy independent-axis dashboard or its configuration.
Use **both printed drums fitted, no rope attached**, and keep drum 2's brake
powered/released. This is a virtual-coordinate demonstration, not loaded operation.

## Hardware assumptions

| Setting | Drum 1 | Drum 2 |
| --- | --- | --- |
| Driver | Leadshine CS2RS-D507 | STEPPERONLINE CL57RS |
| Slave ID | 1 | 2 |
| Gearing | 20:1 | 20:1 |
| Virtual winding diameter | 80 mm | 80 mm |
| Command / reported position scale | 10,000 units per motor turn | 10,000 units per motor turn |
| Encoder resolution parameter | 4000 | 4000 |
| Existing peak current checked | 6 A | 4 A |
| Brake | Driver DO1 / relay, automatic | Constant 24 V, released |
| Positive command | Anticlockwise / wind in | Anticlockwise / wind in |

Rotation is viewed from each outer drum face towards its gearbox. Both drives
share one serial adapter at **38400 baud, 8N1**. One RTU handler serializes slave
selection and transactions. The process takes an advisory port lock; stop other
serial clients, including manual diagnostic scripts, before starting it.

Only the verified common register subset is used: `0x1003`, `0x2203`, `0x601D`,
`0x602A..0x602D` for state; seven path words at `0x6200..0x6206`; `0x6002` for
finite-path trigger/quick stop. `0x000F` forced enable is used only on Leadshine
slave 1, with DI1=`0x0008`, DO1=`0x0024`, and 250 ms brake delays verified first.
The CL57RS path/status subset was verified in its manual and bench tests; this
does not assert that every register is interchangeable between driver models.

## Run

```sh
cd controller
go test ./...
go build -o depthpilot-rope ./cmd/depthpilot-rope
# Local simulation, same API and motion orchestration:
./depthpilot-rope -simulate -listen 127.0.0.1:8080
# Jetson hardware (user must have serial-port permission):
./depthpilot-rope -simulate=false -listen :8080
```

The UI is embedded; no separate frontend server or npm build is required for
this mode. Bind to the trusted LAN only; this commissioning UI has no user login.
State-changing requests require JSON and reject cross-origin browser requests.

For browser acceptance testing, start simulation on `127.0.0.1:8088` and run
`npm run test:rope` from `web/` with a supported Node version and Playwright
Chromium installed. The script checks the server's simulation mode before issuing
any commands. `go test -race ./...` covers transforms, limits, paired programming,
partial failures, lost trigger acknowledgements, telemetry faults, timeouts,
disabled pending commands and brake/enable restoration.

For the Jetson, install the Linux ARM64 binary at
`/opt/depthpilot-rope/depthpilot-rope`, install
[`deploy/depthpilot-rope.service`](deploy/depthpilot-rope.service) into
`/etc/systemd/system/`, then run:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now depthpilot-rope
sudo systemctl status depthpilot-rope
journalctl -u depthpilot-rope -f
```

Startup reads telemetry only. It never enables or moves a drive and never
automatically captures a reference. A service restart requires a fresh origin.

## UI workflow

1. Check both drives are fault-free and the displayed telemetry is live.
2. Click **Capture current positions as virtual origin**. Independent command
   and feedback offsets are captured; no homing or counter-reset commands are sent.
3. Run the 100 mm buttons: **Right**, **Left**, **Lower**, **Raise**, then
   **Right + raise**. These are relative virtual steps. The target form is absolute
   within the temporary virtual coordinate frame.
4. Confirm direction visually. Horizontal moves rotate both drums in opposing
   directions; vertical moves rotate drum 2 alone; equal right-and-raise leaves
   drum 2 stationary. A 100 mm rope equivalent is approximately 143.24° of drum turn.
5. **Stop both drives** interrupts the plan and invalidates the virtual origin.
   Capture a new origin after rest before continuing.

If disabled drum 1 retains a pending path command, the controller first replaces
it with a quick-stop request so the old path cannot resume on enable.
When drum 1 has a nonzero planned move, the controller temporarily enables it, waits two seconds for brake
release, programs and reads back both required profiles, then issues adjacent
unicast triggers exactly once. It checks both drives for completion and position
error, waits one second after rest, and restores drum 1's forced enable to zero.
With the verified normally-open DI1 configuration, that disables drum 1 and lets
the driver control brake engagement. A separate `disabling` phase waits at least
300 ms and polls for up to two seconds for the enable bit, forced-enable register
and DO1 output to clear; acknowledgement of the write alone is not completion.
Drum 2 stays enabled with its brake released.
No EEPROM writes or automatic alarm-reset commands are issued.

For pure lower/raise moves with zero planned drum 1 displacement, drum 1 stays
disabled with its automatic brake engaged throughout. Its enable register and
path are not written. This avoids an unnecessary brake-release/enable cycle that
can shift the resting position slightly. Unexpected stationary-drum counter
changes still cause a paired stop; no position tolerance is added to that check.

When a requested rope target is unchanged from the preceding target, the planner
holds the settled counter if it is within 25 counts (approximately 0.031 mm) of
the absolute ideal. This avoids scheduling a one-count correction alongside a
full move on the other drum. The allowance is measured against the original
absolute target, not accumulated across moves. Larger discrepancies are rejected;
changed targets still use the origin-based absolute calculation. During motion,
the held drum is checked against its captured settled counter without tolerance.

## Geometry and coordination

`x` is rightward trolley travel and `z` is downward camera drop, in metres:

```text
ΔL1 = −Δx                 ΔL2 = Δx + Δz
motor turns 1 =  Δx        × 20 / (π × 0.080)
motor turns 2 = −(Δx + Δz) × 20 / (π × 0.080)
```

Targets use absolute command counters relative to the captured command origin,
so repeated moves do not accumulate rounding drift. Feedback is transformed
back into X/Z using its own captured origin.

Profiles use proportional motor speeds and a common nominal 0.25 s acceleration
and deceleration time. RPM and ramp parameters are integers. The planner searches
for compatible speeds at or below the requested speed and rejects
profiles with more than **25 ms predicted duration mismatch**, instead of treating
rounded speeds as perfectly synchronized. Stationary drums are not triggered.
The UI reports planned duration, per-drum RPM and measured trigger-request skew;
skew above 50 ms causes a paired stop. This is bounded approximate coordination
for unloaded demonstration, not deterministic hardware synchronization. It does
not validate the loaded camera path.

Initial bounds: ±1 m virtual X/Z, 0.1–250 mm Euclidean move length,
0.005–0.05 m/s requested Cartesian speed, 4–240 motor RPM and 30 s planned duration.
Some unequal drum-distance ratios cannot meet these integer-profile limits at a
given speed and are rejected. These bounds do not describe physical rope limits.

Faults, telemetry failures, motion timeout, unexpected counter changes or excessive
tracking error request stops on both drives and invalidate the reference. A lost
trigger acknowledgement is never retried. Driver stopping and brake restoration
are confirmed where communication permits. A lost RS485 link can prevent stopping;
finite profiles may finish even if the browser or Jetson loses contact. There is
no claimed drive-side communications watchdog. Shutdown requests stops and checks
standstill before restoring drum 1's enable state.

## API

### Common-start investigation (2026-11-02)

**Experimental mode implemented:** start with `-broadcast-start` to use one
broadcast frame for two-moving-drum plans. Sequential start remains the default
when the flag is omitted. The UI badge reports the configured mode; each trace
and CSV records whether that move used `broadcast` or `individual` triggering.
A broadcast's numeric trigger-request skew is zero because only one frame is
sent; this is **not a measurement of actual motor-start skew**.

The optional [`deploy/broadcast-start.conf`](deploy/broadcast-start.conf) systemd
override enables the experiment on the Jetson. Install it at
`/etc/systemd/system/depthpilot-rope.service.d/broadcast-start.conf`, reload systemd
and restart the service while at rest. To return to sequential start, remove only
that override, reload systemd and restart. Either restart requires a new origin.

The implementation sends one FC06 frame through the already-open descriptor for
the same configured serial device, under the same bus transaction mutex. The
normal RTU handler remains connected (idle close disabled); the broadcast path
does not read a reply. It leaves 5 ms before transmission and 20 ms afterwards for
transmission/USB buffering and bus turnaround. Short/failed writes are uncertain
outcomes: request addressed stops on both drives, never retry the trigger.

Both the [Leadshine CS2RS manual](https://www.leadshine.com/upfiles/downloads/19d35f99f92b99e0417f74134a985aff_1669716112793.pdf)
and [CL57RS manual](https://www.oyostepper.com/images/upload/File/CL57RS.pdf)
list slave address **0 as broadcast** in section 4.1, FC06 single-register writes,
and `0x6002 = 0x0010` as the path-0 start command in their PR register tables.
This makes one broadcast start a documented candidate for the existing single
adapter/shared bus. Broadcast start has now passed unloaded Right/Left and
mid-move stop tests on these two drives. Internal execution latency need not match.
Keep the physical slave addresses at 1 and 2; address 0 belongs to the broadcast
request, not a change to the DIP settings.

Proposed execution: individually program and read back both paths, then send one
FC06 broadcast start without expecting a response, followed by individually
addressed status checks. Do not retry an uncertain trigger. The current goburrow
RTU transport waits for a response even at address 0, so simply changing the slave
ID in the existing Write method is insufficient: the experimental mode therefore
uses the explicit write-only handling described above.

Use broadcast only when **both drums have nonzero planned movement**. Retain
individual triggering for one-drum moves so a stationary drive cannot execute a
stale path. Both manuals also describe the CTRG digital input and rising-edge
trigger mode; a common wired trigger is another option, requiring electrical
interface and path-selection setup. Two independent USB adapters do not provide
a deterministic shared start and are unnecessary for the broadcast approach.

Recorded 100 mm horizontal tests at 0.02 m/s showed 14.46–14.60 ms trigger-request
separation, 0.27–0.39 mm typical vertical deviation, and 0.45–0.49 mm peaks near
deceleration. Encoder-derived endpoint error was 0.005–0.010 mm. These are virtual
shaft-based measurements, not loaded camera accuracy. A first-order horizontal
start-skew estimate is `vertical mismatch ≈ horizontal speed × start skew`:
14.6 ms corresponds to 0.29 mm at 0.02 m/s and 0.73 mm at 0.05 m/s. Faster future
operation increases this term; winding calibration, mechanics and load response
remain separate error sources. Broadcast is a potential refinement, not a
prerequisite established by these slow unloaded results.

### Last-move visualization

Broadcast bench results (2026-11-02): 100 mm at 0.02 m/s completed in both
directions without faults. Right steady-travel vertical deviation was about
0.04–0.11 mm; Left middle-of-travel deviation was −0.043 to +0.032 mm. Recorded
startup peaks were 0.345/0.437 mm across ~262/263 ms interpolation intervals,
so those peaks should not be interpreted as precise instantaneous path error.
Endpoint errors were ~0.011/0.009 mm in the virtual encoder-derived frame.
A subsequent broadcast Right move interrupted by the UI stop ended with trace
outcome `stopped`, both drives at rest and fault-free, Drum 1 disabled after its
brake sequence, and the virtual reference invalidated as intended. Broadcast
mode remains enabled on the Jetson; stop commands remain individually addressed.

The **Last move** panel plots the requested straight X/Z segment and the path
reconstructed from shaft feedback with equal axis scales (positive Z downwards).
It marks start, target and latest aligned sample and reports maximum perpendicular
distance from the requested line, settled endpoint error, observed motion duration
and trigger-request skew. Endpoint error is reported only for completed moves;
interrupted or faulted traces are labelled accordingly. Duration runs from the
first trigger request to observed completion/standstill and includes polling latency.

Each drum's position-register transaction is timestamped at its start and end;
the midpoint is the nominal sample time. The common trace uses the earlier of the
latest sample times, linearly interpolating the other drum only between known
samples. It never extrapolates or bridges intervals over 500 ms. The panel reports
the maximum read half-window and interpolation interval; these are timing
limitations, not a bound on physical position accuracy. Accelerations and bus
latency still affect this encoder-derived estimate.

The latest move retains up to 2,000 raw samples per drum and 2,000 aligned points.
It remains visible after stop or recapturing the origin, with its original frame,
until another move replaces it or the service restarts. CSV export includes raw
command/feedback counters, transaction timestamps, aligned counters and coordinates,
the original coordinate offsets, target and trigger skew.

This visualization uses the fixed virtual diameter: it does not measure rope slip,
stretch, changing winding radius or actual camera motion.

| Request | Purpose |
| --- | --- |
| `GET /api/state` | Mode, phase, virtual coordinates, drum counters/status, plan and recent session events |
| `GET /api/trace.csv` | Download the last move's raw and aligned samples (404 before the first move) |
| `POST /api/reference` with `{}` | Capture current counters as a temporary origin |
| `POST /api/move` with `{"x":0.1,"z":0,"speed":0.02}` | Queue one finite coordinated move |
| `POST /api/stop` with `{}` | Request quick stop on both drives |

POST requests use `Content-Type: application/json`. Busy/faulted/unreferenced
movement requests are rejected. Session origin and events are intentionally not
persisted. This mode has no automatic home return, continuous jog or loaded-rope
calibration yet.

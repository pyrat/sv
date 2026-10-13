# DepthPilot control station

Dark marine-instrument UI for the Jetson: React + TypeScript + Vite with a Go
HTTP API. The built UI is served by Go, so **Node is only needed to build or
develop**, not to run the Jetson service. Fonts, icons, scripts and styles are
local: the deployed UI does not require an internet connection.

## Start the hardware-free environment

Requirements for the Makefile: **npm/npx** and **Go 1.25+**. It automatically
uses an isolated Node 22/npm 10 installation if your shell's Node is older than
22.12; your global Node installation is unchanged. Go automatically selects the
module's recommended Go 1.25.2 toolchain (or a newer installed toolchain).
First-time downloads require internet access; subsequent runs use the cache.
From the repository root:

```sh
make sim
```

This installs frontend dependencies when needed, builds the UI and Go service,
and starts the simulator in the foreground. Open **http://127.0.0.1:8080**, or run
`make open` in a second terminal. Press **Ctrl+C** to stop it.

If port 8080 is already in use, run `make sim PORT=8081` and open
**http://127.0.0.1:8081** (`make open PORT=8081`). For access from another device
on the LAN, run `make sim HOST=0.0.0.0`, then browse to the simulator machine's
IP address on port 8080. `make help` lists all commands.

`make sim` keeps separate saved home/activity files for each port. Port 8080
uses the original `.depthpilot-sim-home.json`; port 8081 uses
`.depthpilot-sim-8081-home.json`, with its own adjacent activity database.
Override `SIM_STATE=path` to choose a different state file (relative to
`controller/`), or `SIM_STATE=-` for an ephemeral simulator.
If startup reports `SQLITE_BUSY`, another service may still own the same activity
database. Stop that instance or select a separate state file; do not delete the
database or its companion files to clear a live lock.

For the direct `npm` commands below, use Node 22.12+ yourself, for example with
`nvm install 22 && nvm use 22`. The Makefile handles this automatically.

Open **http://127.0.0.1:8080**. The default is simulation, which opens no serial
devices and uses two independent simulated motors. Initial positions are
X = 2.450 m and Y = 4.800 m; demo travel is X 0–50 m / Y 0–30 m and maximum speed is 200 mm/s.
The banner always identifies simulation. Restarting resets the simulated motors
to those positions, but preserves saved home; use Clear home to remove it.

For frontend hot reload, leave the Go service running and run in another terminal:

```sh
cd web
npm run dev
```

Open the Vite URL printed in the terminal. Its `/api` proxy points to the local
Go service. The simulator implements the same HTTP contract as hardware.
It models time-based travel and stop/lease semantics, not motor inertia, brake
physics, rope stretch, or encoder errors.

## Persistent activity history

Activity is saved locally in SQLite and restored after service restarts. The
dashboard shows recent entries; **View history** opens dated, paginated history
with event IDs and service-session identifiers. The latest **10,000 events** are
retained by default; use `-activity-limit` to change this (1–1,000,000).

By default the database sits alongside the home state file:

- Simulator: `.depthpilot-sim-home.activity.sqlite`
- Hardware: `.depthpilot-home.activity.sqlite`
- Override with `-activity /path/to/activity.sqlite`.
- `-activity -` uses memory only. `-state -` also defaults activity to memory only.

The deployed systemd service stores it beside its home state under
`/var/lib/depthpilot/`. Simulator and hardware databases are separate; opening a
database belonging to the other mode is rejected. One service owns each database.
Stop the service before copying the database for backup, including any remaining
`-wal` and `-shm` companion files.

Writes are batched on a background worker, keeping disk I/O off the motor-control
loop. Normal shutdown flushes pending entries. Abrupt power loss can lose entries
still queued in memory. Storage failures are retried and displayed in the UI;
the bounded queue holds 1,024 entries, with overflow explicitly reported rather
than blocking motor control. Retention bounds entry count, and SQLite reuses
freed pages (the file need not shrink immediately).

`GET /api/activity?limit=50&before=123` returns newest-first `events`, a
`nextBefore` cursor (0 at the end), and `storage` status. Omit `before` for the
newest page; `limit` accepts 1–100. `/api/state` includes the latest 12 events and
`activity` storage status. History refreshes on opening or pressing Refresh history.

## Controls

On phones (600px wide or less), a compact view replaces the desktop sidebar and
sea-cage illustration. **Control**, **Home**, and **History** navigation keeps
each task separate. Control shows the live plot, axis readings and movement
controls; **Stop all motion** stays fixed at the bottom with safe-area spacing.
Diagnostics are available from the icon beside the phone navigation.

- **Target position:** select X and/or Y, enter metres, select speed, move.
  Outbound target moves run **X first, then Y**. The second axis remains queued
  until the first reports path completion, standstill and arrival at its target.
- **Drag the glowing position marker:** preview an X/Y destination and release
  to move **X then Y** at the selected speed. No motion command is sent during
  the drag. The actual feedback dot stays visible while previewing. Travel is
  clamped to the configured bounds; Escape, lost capture, window blur, and
  disconnected telemetry cancel the drag. Both axes must be available and idle.
  Keyboard users can focus the marker, adjust targets with arrow keys (10 mm),
  and press Enter to move.
- **Manual jog / Step:** 1 mm, 10 mm or 100 mm signed increments.
- **Manual jog / Hold to move:** pointer/touch press or hold Space/Enter on a
  focused direction button. Release, window blur or hidden tab requests a stop.
- **Stop all motion / Escape:** sends a drive quick-stop request on all axes.
  Any remaining queued stage is discarded; it does not resume automatically.
- **Set current as home:** capture both stationary positions and raw counters.
  This saves a destination; it does **not** reset either drive counter.
- **Edit home:** enter metres or signed command-counter targets for every
  available axis. Counter targets use command-pulse calibration, not encoder
  counts. The server validates all values against travel limits before saving.
- **Return home:** run **Y first, then X**, at the selected speed. A single-axis
  installation returns only that configured axis. A home marker and an At home
  indicator show the saved destination and arrival status.
- **Diagnostics:** axis availability, limits, configured speed, and fault details.

The plot and readouts display actual feedback, refreshed approximately every
150 ms. Targets are a separate marker. Failed polling marks values stale and
disables movement. X/Y are **motor axes**, not geometric camera coordinates;
the view uses a **top-left origin** for the demo X 0–50 m / Y 0–30 m ranges: positive X is
right/along the cage, positive Y is down/into the water. Dragging downward,
the down jog button, and the Down arrow key all increase Y. With custom travel
ranges, the top-left corner shows the configured minimum X/Y values. Set the
hardware axis `direction` to match this convention during commissioning.
Saved home is still an independent destination; changing the plot orientation
does not redefine zero or reset counters.

Two-axis commands execute one axis at a time, not synchronously interpolated.
The sequence panel shows which axis is moving and which is waiting. The ordering
is explicit by operation: target/drag moves are outbound X→Y; Return home is Y→X.
It is not inferred from the signs of the requested coordinate changes.
The sea-cage illustration establishes the installation context (above-water
winches and a submerged camera); it is explicitly schematic. Neither that
illustration nor the motor-axis plot claims a calibrated physical camera
position, cage depth, rope tension, or environmental measurements.

## Automated tests and API mocking

From the repository root, `make test` builds the app, installs the Playwright
Chromium browser, and runs Go checks and browser tests. Use `make test-go` or
`make test-ui` to run either group separately. Equivalent individual commands:

```sh
cd controller
go test -race ./...
go vet ./...
cd ../web
npm run build
npx playwright install chromium
npm test
```

Playwright automatically starts a dedicated Go simulator on port 18080, serves
the production UI, and tests real UI-to-API movement, step jogging, hold/release,
lease expiry, travel validation, mocked faults/disconnects, and mobile layout.
If you already have a test simulator on port 18080, use
`PW_REUSE_SERVER=1 npm test` **only against that simulator**. On machines with an
HTTP proxy, add `NO_PROXY=127.0.0.1,localhost` so loopback requests stay local.
Recent macOS versions need a recent Go toolchain if an older toolchain fails
with a `missing LC_UUID` loader error; verified here using Go 1.25.2.

Screenshots are written to `web/test-results/dashboard-desktop.png` and
`dashboard-mobile.png`; failure traces are retained. The test suite also uses
Playwright route fixtures to replace `/api/state` and command responses. This
lets you test unavailable axes, faults, and stale telemetry without a driver.

## Run on the Jetson LAN

Build the UI as above, then build the service on the Jetson:

```sh
cd controller
go build -o depthpilot-web ./cmd/depthpilot-web
./depthpilot-web -listen :8080 -simulate -web ../web/dist
```

Open **http://JETSON-IP:8080** from the local network. The default loopback bind
is deliberately changed to `:8080` here for LAN access. The service has no login;
it is intended for the requested trusted local control network.

For startup at boot, use [`depthpilot.service`](../controller/deploy/depthpilot.service).
The example assumes a `depthpilot` service account, files installed under
`/opt/depthpilot`, and `dialout` membership for serial access. Adjust paths/account
to match your Jetson, install to `/etc/systemd/system/`, then:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now depthpilot
```

## Hardware configuration

Copy [`config.example.json`](../controller/config.example.json) to
`controller/config.local.json`. Its numeric values are **demo placeholders**,
and `reference_valid` is false. Replace them with measured calibration before
enabling that axis. An omitted Y axis is displayed as unconfigured; no second
motor or adapter is assumed.

```sh
./depthpilot-web -listen :8080 -simulate=false -config config.local.json -web ../web/dist
```

Each axis needs:

| Field | Meaning |
| --- | --- |
| `port`, `slave`, `baud` | Dedicated USB/RS485 channel, slave 1–127, matching baud; 8N1 |
| `pulses_per_metre` | Command-pulse calibration used for absolute move targets |
| `feedback_counts_per_metre` | Separate actual-position feedback calibration |
| `pulses_per_revolution` | Configured command resolution; converts linear speed to motor RPM |
| `min_m`, `max_m` | Travel bounds in the established coordinate frame |
| `direction` | +1 or −1: maps positive UI travel to positive drive counts |
| `max_speed_mps` | Per-axis maximum linear speed |
| `ramp_ms_per_1000rpm` | Drive acceleration/deceleration parameter |
| `reference_valid` | Operator has established the drive reference for this session |
| `reference_id` | Stable coordinate-frame identity required for hardware home capture; change when the datum changes |

Zero is the drive's existing homed/established reference. No reference is inferred
from a saved position, and setting `reference_valid` does not perform homing.
Re-establish reference after power loss before enabling hardware movement.
The first implementation uses a **constant metres/count conversion**: multilayer
drum winding needs a measured nonlinear calibration model before these values
can represent deployed rope length over the full travel.

### Saved home, counters and the reference

Home is a saved parking destination, distinct from physical homing/zeroing.
Current profile/command counts and encoder-feedback counts are exposed separately
in diagnostics. Capturing home stores actual positions, converted command targets,
and the raw profile/feedback counter snapshots. Editing home does not move a motor.

The service atomically saves home to `.depthpilot-sim-home.json` in simulation or
`.depthpilot-home.json` in hardware mode, relative to its working directory. Both
files are ignored by Git. Override with `-state /path/to/home.json`; `-state -`
selects ephemeral state for tests. The systemd example uses an owned directory
under `/var/lib/depthpilot`. No motion queue or live-counter snapshot is restored
into the drives on startup. A corrupt state file is reported rather than replaced.

Hardware home capture requires a nonempty **`reference_id`** in each available
axis configuration, as well as `reference_valid`. Give the established coordinate
frame a stable ID (for example `cage-01-x-datum-v1`). Change it when the zero/datum
changes; merely re-establishing the same known datum can retain the ID. Saved home
is tied to the axis configuration and reference, and becomes unusable when these
change until it is set again. The `reference_id` is an operator-maintained identity,
not an automatic power-loss detector: re-establish the physical reference after
drive power loss before asserting `reference_valid` again.

Return-home arrival uses actual feedback within 1 mm or one calibrated count,
whichever is larger, and requires drive path completion and standstill. The next
axis will not run if feedback is missing, the first drive stops short, or any axis
faults. Stop and movement deadlines also discard the remaining stage. Physical
switch-based homing and counter-zero commands remain separate commissioning
operations; this home feature does not invoke them.

Motor tuning/current, encoder setup, hardware travel switches, brake assignment,
and brake timing remain in MotionStudio, following the wiring diagram and
[CS2RS manual](../controller/README.md). Nothing writes EEPROM or force-enables
the drive. Hardware faults latch in the web service; diagnose the drive, stop
motion, and restart the service after resolving them.

### Jog command lifecycle

1. `/api/jog` validates availability, direction, speed and travel; configures
   Pr6.00 (`0x01E1`) and Pr6.03 (`0x01E7`), then issues a random lease token.
2. The browser renews the lease every 100 ms while held. Each renewal grants
   400 ms; expired/revoked tokens can never create or restart motion.
3. The service refreshes `0x1801 = 0x4001/0x4002` on a 20 ms loop, interleaving
   refreshes with telemetry transactions. Only one jog can be active at a time.
4. Expiry, release, stop, faults, or service shutdown end refreshes and request
   `0x6002 = 0x40`. Software jog boundaries include a polling/ramp allowance.

V1.5 section 5.3 says continuous jog needs refresh intervals below 50 ms; it
does **not** establish a rated stop-on-expiry time or stopping distance. USB
latency and serial timing require bench measurement. Host software limits and
best-effort stop requests do not replace the driver's wired travel/stop inputs.
Finite target moves continue independently of the browser; the service monitors
them and applies a travel-time-plus-15-seconds deadline. A failed RS485 connection
can prevent any stop request reaching the drive. No automatic motion retries or
resume are performed. Brake sequencing stays in the drive.

## API

Same-origin JSON HTTP API. Errors use `{"error":"..."}` with non-2xx status.
Command requests require `Content-Type: application/json`.

| Endpoint | Request |
| --- | --- |
| `GET /api/state` | `{simulation, axes, events, time, home, motion}`; per-axis actual/target metres, raw counters, state, limits, timestamp |
| `POST /api/move` | `{"targets":{"x":2.5,"y":4.8},"speed":0.1}` |
| `POST /api/step` | `{"axis":"x","delta":-0.01,"speed":0.1}` |
| `POST /api/jog` | `{"axis":"x","direction":1,"speed":0.1}` → `{lease}` |
| `POST /api/jog/renew` | `{"lease":"returned-token"}` |
| `POST /api/stop` | `{}`; revokes all leases and requests all-axis quick stop |
| `POST /api/home/set` | `{"source":"current"}`; or `{"source":"positions","targets":{"x":2,"y":4}}`; or `{"source":"counters","counters":{"x":106000,"y":212000}}` |
| `POST /api/home/return` | `{"speed":0.1}`; Y→X ordered return to saved home |
| `POST /api/home/clear` | `{}`; persistently clears the saved destination |

All targets validate before either axis starts. Initial programming errors are
returned by the command API. A failure during a later stage cancels the sequence,
requests an all-axis stop, and is reported through `motion.error`, axis state and
the activity stream. Stop acknowledgment means the request was sent, not that
physical standstill has already been confirmed.

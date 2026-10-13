# Supervinsj

Source snapshots, hardware documentation, and development notes for smart-winch projects named **daWinchi** and **Ropey**. The code covers a two-axis winch (up/down and in/out), local button controls, remote commands, and hardware experiments.

## Repository status

The new [DepthPilot controller](controller/README.md) provides a buildable Go motor-control library and CLI for the CS2RS-D507 over USB/RS485. Historical reference material lives under `sv-docs/`. The standalone `dawinchi/wiggle` utility has a Go module; the archived firmware and motor experiment reference internal packages that are absent here.

The [historical daWinchi README](sv-docs/dawinchi/old-repo/dawinchi/README.md) contains an unfinished application template and references Vue/Express components that are not included in this checkout.

## Try the new control station

With **npm/npx** and **Go** installed, run from the
repository root:

```sh
make sim
```

Open **http://127.0.0.1:8080** (or run `make open` in another terminal).
This installs dependencies, builds the app, and starts two simulated axes with
no winch attached. Press **Ctrl+C** to stop. Use `make sim PORT=8081` if port 8080
is occupied, and `make help` for build/test commands.

If your shell has an older Node, the Makefile automatically uses an isolated,
cached Node 22/npm 10 toolchain. Go also selects the module's recommended
toolchain automatically. First-time toolchain downloads require internet access.

See the [control station guide](web/README.md) for LAN access and Jetson deployment.

## Repository map

| Path | Contents |
| --- | --- |
| [`web/`](web/README.md) | React/TypeScript control dashboard, live X/Y telemetry, target moves and jogging, simulator environment, and Playwright browser tests. |
| [`controller/`](controller/README.md) | Go CS2RS driver library/CLI plus the Jetson web/API service and two-axis simulator. |
| [`sv-docs/dawinchi/wiggle/`](sv-docs/dawinchi/wiggle/) | Go WebSocket client that repeatedly moves the winch vertically and reports errors to Slack. Includes `go.mod`, `go.sum`, and a prebuilt executable. |
| [`sv-docs/dawinchi/CCP/lib/`](sv-docs/dawinchi/CCP/lib/) | CircuitPython libraries for WIZnet Ethernet and an INA226 current sensor. No CCP application entry point is included. |
| [`sv-docs/dawinchi/panel-buttons-test.py`](sv-docs/dawinchi/panel-buttons-test.py) | Raspberry Pi GPIO button-state monitor. |
| [`sv-docs/dawinchi/`](sv-docs/dawinchi/) | Wiring diagrams, mechanical/electrical PDFs, panel UI artwork, field measurements, button-command notes, and bundled tools/libraries. |
| [`sv-docs/ropey/old-code/`](sv-docs/ropey/old-code/) | Partial Go firmware snapshot: winch coordination, panel handling, state persistence, settings, API-facing methods, GPIO abstraction, mocks, and panel tests. |
| [`sv-docs/ropey/mt/main.go`](sv-docs/ropey/mt/main.go) | Raspberry Pi motor/rotary-encoder experiment. |
| [`sv-docs/ropey/motor_driver_4020593.pdf`](sv-docs/ropey/motor_driver_4020593.pdf) | Motor-driver reference document. |

## Control model

The archived winch code coordinates two motors using encoder-counter targets:

- **Up/down** controls vertical movement.
- **In/out** controls horizontal movement.
- Local panel handling supports home positioning, precision mode, service mode, and factory-reset actions.
- API-facing methods support movement, stopping, settings, home-position changes, and CCP control.
- Firmware state includes motor and panel state, operating modes, remote lock, home position, and the last action source.
- State is stored as JSON records in a SQLite database at `/etc/winch.conf`.

Useful starting points are [`winch.go`](sv-docs/ropey/old-code/internal/winch/winch.go), [`move.go`](sv-docs/ropey/old-code/internal/winch/move.go), [`panel.go`](sv-docs/ropey/old-code/internal/winch/panel.go), and [`state.go`](sv-docs/ropey/old-code/internal/winch/state.go).

### Remote protocol

The `wiggle` client demonstrates JSON-RPC 2.0 messages over WebSockets, with an `Authorization: Bearer <token>` header:

| Default endpoint | Methods used by the client |
| --- | --- |
| `ws://127.0.0.1:9001/api/v1/query` | `counter` |
| `ws://127.0.0.1:9001/api/v1/cmd` | `move`, `stop_move` |

A counter query looks like:

```json
{
  "id": "example",
  "jsonrpc": "2.0",
  "method": "counter"
}
```

Responses expose `status`, `up_down_motor_counter`, `up_down_motor_speed`, `in_out_motor_counter`, and `in_out_motor_speed`. Movement requests use the four motor fields plus `sync`. These details are evidenced by the client and field notes; the server's RPC implementation is not included.

## Working with the code

### Wiggle utility

Requires Go **1.22.7 or newer**, as declared in its module. Build from the repository root:

```sh
cd sv-docs/dawinchi/wiggle
go mod download
go build -o wiggle-local .
```

Before running, configure the values in [`main.go`](sv-docs/dawinchi/wiggle/main.go) for the target installation:

| Setting | Current behavior |
| --- | --- |
| `wsBasePath` | Defaults to the local server at port `9001`, under `/api/v1/`. |
| `jwtToken` | Source-defined bearer token; not read from the environment. |
| `wiggleUpCounter` / `wiggleDownCounter` | Source-defined vertical targets, currently `650000` / `952242`. |
| `wiggleSpeed` | Vertical speed value of `30`; movement requests use `10` for in/out. |
| `wiggleMoveDelay` | Five-second pause after each movement response. |
| `SLACK_WEBHOOK` | Required environment variable for error notifications. |

The utility reads the initial in/out counter and uses it for subsequent moves. It repeats up/down movement, skips starting a new cycle during local hours `00:00` through `06:59`, and attempts to send `stop_move` on shutdown. Running it issues real movement commands to the configured server.

After configuring the source, run against the intended winch service:

```sh
export SLACK_WEBHOOK='https://hooks.slack.com/services/your/webhook/url'
go run .
```

### Button monitor

On a Raspberry Pi with Python 3 and `RPi.GPIO` installed:

```sh
python3 sv-docs/dawinchi/panel-buttons-test.py
```

The script uses BCM numbering, configures pins **6, 14, 15, 16, and 17** as inputs with pull-ups, and prints their states every half-second. Press `Ctrl+C` to exit and clean up GPIO state.

### Archived firmware and motor experiment

Neither `sv-docs/ropey/old-code/` nor `sv-docs/ropey/mt/` includes a Go module. Their imports reference missing packages from `github.com/CreateView/dawinchi-hw`, `github.com/CreateView/xzmove`, and `github.com/CreateView/ropey`. The firmware entry point also contains unresolved `firmware` and `winch` references. Restoring the matching source trees and module configuration is necessary before building or running their tests.

The partial [`cmd/firmware/main.go`](sv-docs/ropey/old-code/cmd/firmware/main.go) records these configuration values:

| Environment variable | Default / requirement |
| --- | --- |
| `HW_ENV` | `debug` |
| `HW_API_PORT` | `9001` |
| `HW_API_BEARER_TOKEN` | Required |
| `HW_LOG_FILE_PATH` | `/var/log/winch.log` |

The hardware code targets Raspberry Pi GPIO/SPI through `go-rpio`; the older initialization notes also reference ADS1115 sensing and LS7366R counters. The separate `mt` experiment exercises motor drivers and rotary encoders.

## Hardware and field references

- **CCP wiring:** [diagram 1](sv-docs/dawinchi/CCP-wire-diag-1.png), [diagram 2](sv-docs/dawinchi/CCP-wire-diag-2.png), and an [editable Draw.io diagram](sv-docs/dawinchi/Untitled%20Diagram.drawio).
- **Panel UI:** [OLED UI design](sv-docs/dawinchi/dawinchi_ccp_oled_ui_design-hd.png).
- **Hardware PDFs:** [drum drawing](sv-docs/dawinchi/drum-115669.pdf), [8610-0533 drawing set](sv-docs/dawinchi/8610-0533_rev_1.0.0_shts_1_-_5_of_5.pdf), and additional motor documentation in [`dawinchi/`](sv-docs/dawinchi/).
- **Movement measurements:** [`winch-move-algo.txt`](sv-docs/dawinchi/winch-move-algo.txt) records counter readings, load/height observations, and current measurements.
- **Site tests:** [`NTNU-site-test.txt`](sv-docs/dawinchi/NTNU-site-test.txt) contains recorded positions and example movement requests.
- **Button experiments:** [`test-buttons.txt`](sv-docs/dawinchi/test-buttons.txt) and [`button-cmds.txt`](sv-docs/dawinchi/button-cmds.txt) contain `raspi-gpio` sequences for simulating panel inputs.

The measurements and target positions are installation-specific development notes, not a general calibration specification.

## Bundled dependencies and licensing

The archive includes an Adafruit CircuitPython WIZnet5k source tree, WIZnet5k and WSGI ZIP archives, and an `mpy-cross` executable. The bundled libraries have their own license files or source-header notices.

No repository-level license file is present. The MIT statement in the historical template README does not establish licensing for the complete archive; consult the applicable files and project owners for reuse terms.

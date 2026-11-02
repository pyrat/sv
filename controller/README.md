# DepthPilot motor controller

Go 1.22+ Linux/Jetson controller for the **Leadshine CS2RS-D507**, using
Modbus RTU through the USB/RS485 adapter in the
[prototype wiring diagram](../sv-docs/depthpilot-prototype-wiring.svg).
Supports single-axis absolute/relative position moves, telemetry and quick stop.

The [web control station](../web/README.md) adds a polished React UI, a Go HTTP
service, metre calibration, two-axis simulation, and browser-tested target/step/
hold-to-jog controls. Start it with `go run ./cmd/depthpilot-web -simulate` after
building `web/`; the CLI below remains available for direct drive bring-up.
The web service executes outbound targets X→Y and Return home Y→X, with persistent
home destinations, raw-counter diagnostics and a counter-based home editor. Each
stage waits for completion/arrival before the next axis starts. The single-drive
CLI remains a low-level operation and does not run these two-axis sequences.
These sequential moves describe the current implementation; they do not implement
the coordinated two-drum rope mechanism below.

## Proposed two-drum rope configuration

The agreed mechanism uses **two winding/storage drums**, a trolley travelling on
a separate support cable or rail, and a camera suspended below the trolley:

- **Drum 1:** rope runs from the drum to a fixed pulley at the far end of the
  span, returns to the trolley, and terminates at the trolley.
- **Drum 2:** rope runs from the drum to a pulley attached to the trolley, turns
  downwards, and terminates at the camera.
- The trolley pulley redirects the camera rope; it does not provide a 2:1
  lifting advantage because it does not move vertically with the camera.

### Prototype drum assignment and unloaded commissioning

| Actuator | Drive | Modbus slave | Gear reduction | Prototype motor |
| --- | --- | --- | --- | --- |
| Drum 1 — far-end pulley/trolley rope | Leadshine CS2RS-D507 | 1 | 20:1 | More powerful of the two prototype motors |
| Drum 2 — trolley pulley/camera rope | STEPPERONLINE CL57RS | 2 | 20:1 | STEPPERONLINE motor used in the bench tests |

Production will likely use the same motor model for both drums. Keep each drum's
drive settings and calibration explicit rather than assuming identical motors.

Initially fit the drums to the gearboxes **without rope** and validate coordinated
rotation visually. Both use copies of
[`drum-one-piece.stl`](../mechanical/prototype-drums/drum-one-piece.stl); see the
[one-piece drum assembly notes](../mechanical/prototype-drums/ONE-PIECE.md).
The STL and generator define an **80 mm barrel diameter**, **110 mm flange
diameter**, **50 mm overall height**, and **24 mm straight winding width**.
Use the barrel diameter, not the flange diameter, for the initial virtual payout.
These printed drums are specified for unloaded demonstration, not loaded lifting.

For this no-rope stage, use a fixed **80 mm virtual winding diameter** for both:

```text
virtual payout per drum revolution  = π × 0.080 = 0.251327412 m
motor revolutions per drum revolution = 20
virtual payout per motor revolution = π × 0.080 / 20 = 0.012566371 m
motor revolutions for rope change ΔL = ΔL × 20 / (π × 0.080)
command pulses for rope change ΔL   = motor revolutions × configured pulses/revolution
```

Read and verify each drive's command resolution separately; the CL57RS was tested
at 10,000 command pulses per motor revolution. Wind-in/pay-out motor signs remain
to be established by slow tests with a clearly specified viewing direction. Mark
each drum so turns and direction are visible. With equal virtual diameters and
gearing, a horizontal-only command should produce equal drum rotation magnitudes
in their respective wind-in/pay-out directions; a vertical-only command should
rotate drum 2 alone. A rightward move with an equal upward camera move should leave
drum 2 stationary (`Δx + Δz = 0`).

This stage validates the kinematic transform and coordinated motor behaviour, not
rope tension or loaded positioning. Once rope is introduced, account for its
centreline diameter (approximately barrel diameter plus rope diameter on the
first layer), subsequent layers and measured payout. The fixed 80 mm model is
only the initial unloaded reference.

This replaces the earlier continuous-rope-loop/fixed-diameter X pulley proposal.
The two drums are coupled actuators, not independent horizontal and vertical axes.

### Coordinates and rope-length relationship

Define `x` as trolley position increasing towards the far-end pulley, and `z` as
camera drop below the trolley increasing downwards. Define `L1` and `L2` as rope
paid out from drums 1 and 2; positive changes mean payout, negative changes mean
winding in.

For taut ropes with approximately horizontal runs between the drums, trolley and
fixed pulley, and a vertical camera rope:

```text
ΔL1 = −Δx
ΔL2 =  Δx + Δz
```

| Requested motion | Drum 1 | Drum 2 |
| --- | --- | --- |
| Right 1 m, maintain camera drop | Wind in 1 m | Pay out 1 m |
| Left 1 m, maintain camera drop | Pay out 1 m | Wind in 1 m |
| Lower camera 1 m, hold trolley position | Hold | Pay out 1 m |
| Raise camera 1 m, hold trolley position | Hold | Wind in 1 m |

For example, moving right 1 m while lowering the camera 0.5 m requires drum 1 to
wind in 1 m and drum 2 to pay out 1.5 m. Significant height offsets require the
actual sloping-segment geometry instead of the horizontal approximation. Support
cable sag also means constant drop below the trolley is not necessarily constant
camera elevation.

### Tension and winding calibration

Drum 1 actively pulls the trolley towards the far-end pulley; paying out cannot
push it back. Tension in drum 2's rope, supplied by the suspended camera/load,
provides the opposing pull. Reliable leftward travel depends on load weight,
trolley friction, rope angles and acceleration. If the trolley sticks, winding
drum 2 can lift the camera instead of translating the trolley. Commission the
mechanism under its intended load with gentle acceleration and check tension.

Both drums can change effective winding radius as rope builds up. A constant
metres-per-revolution calibration is insufficient for multilayer winding over
the full travel. Each drum needs a winding model or a measured mapping from motor
position to rope payout, including its gear ratio and starting fill. Irregular
winding and rope stretch can reduce repeatability. Motor encoders measure shaft
rotation, not actual rope payout or camera position.

Before implementing the geometry, record:

- Support span, usable trolley travel, and drum/pulley heights.
- Each drum's core diameter, usable winding width, rope diameter and stored length.
- Verify the installed 20:1 gearing and establish each drum's wind-in/pay-out sign.
- Camera/load weight, allowable camera drop, and travel limits.
- Reference trolley position, camera drop and initial rope fill on each drum.

### Controller implementation requirements

The API should accept trolley position and camera drop, plan a common smooth
trajectory, transform it to rope lengths, and convert those lengths into motor
positions using each drum's calibrated winding map. Velocity and acceleration
profiles must correspond to that same trajectory; simply starting independent
moves together, or moving X then Y, does not maintain the requested camera path.
Coordinated execution, reference/homing handling, brake sequencing, limits and
fault/stop behaviour still need implementation and commissioning.

The connected drives are **Leadshine CS2RS-D507 (slave 1)** and **STEPPERONLINE
CL57RS (slave 2)** on a shared RS485 adapter at **38400 baud, 8N1**. The service
must serialize transactions on that shared bus. The CL57RS has been bench-tested
with encoder feedback, and its 4 A peak-current setting has been saved to EEPROM.
See the [bus and DIP-switch diagram](../cl57rs-modbus-wiring.svg).

`config.example.json` still contains the earlier `cl57pr` X-axis placeholder and
independent-axis calibration. It is not a configuration for this agreed mechanism.
The current service rejects enabled `cl57pr` hardware, and the direct CLI remains
CS2RS-specific. CL57RS integration and the coupled rope model are pending; bench
tests do not establish application support. Example calibration values remain
placeholders and both references start invalid.

## Build and use

```sh
cd controller
go test ./...
go build -o depthpilot ./cmd/depthpilot
./depthpilot -port /dev/serial/by-id/YOUR-ADAPTER status
./depthpilot -port /dev/serial/by-id/YOUR-ADAPTER -relative -position 200 -rpm 60 -ramp 1000 move
./depthpilot -port /dev/serial/by-id/YOUR-ADAPTER stop
```

Flags precede the command. Defaults: slave 1, 38400 baud, 8N1, one-second
transaction timeout, 30-second motion deadline. Match the drive's DIP switches
and Pr5.24 serial format. Only one process should own the serial adapter.
Each move replaces PR path 0 in RAM; nothing is saved to EEPROM.

Position arguments are signed **command pulses**, velocity is motor RPM, and
ramps are ms/1000 RPM. Absolute targets use the drive's existing reference.
Set up the reference and pulse resolution in MotionStudio before absolute
positioning. Telemetry exposes profile and actual position separately; no
rope-length or gearbox conversion is assumed.

The drive must already be enabled and idle with no alarms. Configure motor
current, encoder resolution, limits and homing in MotionStudio. Brake output
assignment and timing can be configured directly over Modbus RS485; see the
[verified controller configuration and Phoenix relay wiring](../sv-docs/depthpilot-brake-configuration.md).
DO1 is assigned Brake Control and saved to EEPROM, with 250 ms release and
engagement delays. The driver controls the Phoenix relay and brake timing (Pr4.19/Pr4.20); software
does not switch the relay or force-enable the motor. Pr0.07=0 only returns
enable control to I/O, so it is deliberately not exposed as a disable command.

Move waits for path completion. SIGINT/SIGTERM, deadline expiry, drive faults,
or communication errors request one best-effort quick stop. Stop uses the
Pr8.23 ramp and retains holding torque. A lost RS485 link can prevent that
request reaching the drive; finite PR moves may continue. No communication
watchdog is claimed. A failed trigger has an unknown outcome and is never
automatically retried. A separate `stop` invocation acknowledges the request,
not physical standstill. Hardware stop/limits and brake commissioning are as
specified in the wiring drawing.

## Source and register map

Source: **DepthPilot Proto1 BOM #7**, driver part #69, attachment
`6040-4_Modbus_RTU_Closed_Loop_Stepper_Drive_CS2RS_User_Manual_V1.5.pdf`.
[Supplier's same manual](https://www.damencnc.com/userdata/file/6040-4_Modbus_RTU_Closed_Loop_Stepper_Drive_CS2RS_User_Manual_V1.5.pdf),
sections 4.7, 5.5–5.6 and appendix pp. 54–56.

| Address | Use |
| --- | --- |
| `0x1003` | Fault/enable/running/completion bits |
| `0x2203`, `0x601D` | Current alarm and PR alarm |
| `0x6002` | Path status; `0x10` starts path 0, `0x40` quick-stops |
| `0x602A..0x602D` | Profile and actual signed 32-bit positions, high word first |
| `0x6200..0x6206` | Path 0 mode, position high/low, RPM, acceleration, deceleration, pause |

FC03 reads, FC16 programs the seven path registers, FC06 triggers. `0x6207`
aliases the trigger and is excluded from path programming. Position mode is
`0x0001` absolute or `0x0041` relative; jump, overlap and interrupt are off.
No homing command is exposed because the drawing has no selected home input.

The older Pico/CL57T prototype plan predates this wiring. The BOM also contains
CL57RS/CL57PR parts; the implemented register map is specifically for CS2RS.
The drawing names a Waveshare adapter while the BOM lists FIT0737; either
requires the correct Linux serial device and RS485 wiring. Tests use a fake
register transport; physical motion/brake behaviour still requires bench testing.

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
current, encoder resolution, limits, homing and brake output in MotionStudio.
The driver controls the Phoenix relay and brake timing (Pr4.19/Pr4.20); software
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
CL57RS/CL57PR alternatives; this register map is specifically for CS2RS.
The drawing names a Waveshare adapter while the BOM lists FIT0737; either
requires the correct Linux serial device and RS485 wiring. Tests use a fake
register transport; physical motion/brake behaviour still requires bench testing.

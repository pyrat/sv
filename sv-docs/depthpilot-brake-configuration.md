# DepthPilot controller and Phoenix brake relay

## Power arrangement

Retain **48 V for the Leadshine CS2RS-D507** and the **DDR-15L-24 DC/DC
converter for the 24 V brake and relay**. The relay is **K1, Phoenix Contact
2966171** in the [prototype wiring diagram](depthpilot-prototype-wiring.svg).

48 V was retained for motor torque at approximately **830–840 RPM** (20:1
gearbox, 92 mm bare drum, 0.20 m/s rope speed). This is a design target, not
a measured loaded operating point. The driver's maximum supply rating is 50 V.

## Controller configuration — 22 October 2026

Communication was verified on the Jetson:

| Setting | Verified value |
| --- | --- |
| SSH | `createview@172.26.240.121` |
| Hostname | `legschocolate-c2d` |
| USB–RS485 device | `/dev/ttyUSB0` |
| Device symlink | `/dev/serial/by-id/usb-1a86_USB_Serial-if00-port0` |
| Protocol / slave | Modbus RTU / **1** |
| Serial format | **38400 baud, 8N1** |

Only one process should use the adapter at a time. Commissioning commands
used `sudo` because the login user lacked direct access to the serial device.

**SO1 in the parameter table is DO1 on the physical connector.** The output
common is **COMO** (letter O), sometimes read as COM0.

| Parameter | Register | Readback | Meaning |
| --- | --- | --- | --- |
| Pr0.07 | `0x000F` | `0x0000` | Software forced enable off; input controls enable |
| Pr4.02, SI1 / DI1 | `0x0145` | `0x0008` | Normally-open Enable; saved; DI1 unwired |
| Pr4.11, SO1 / DO1 | `0x0157` | `0x0024` | Normally-open Brake Control; saved |
| Pr4.12, SO2 / DO2 | `0x0159` | `0x0000` | Unassigned; initial readback |
| Pr4.13, SO3 / DO3 | `0x015B` | `0x0000` | Unassigned; initial readback |
| Pr4.19 | `0x0167` | `250` (`0x00FA`) | Brake release delay, ms |
| Pr4.20 | `0x0169` | `250` (`0x00FA`) | Brake engagement delay, ms |

The delay range is 0–1500 ms. Existing 250 ms delays were retained; physical
timing has not been validated. `0x00A4` is the inverted brake assignment and
is **not** the selected configuration.

### Modbus setup and persistence

These are raw hexadecimal register addresses, not 4xxxx labels.

1. **FC06 (Write Single Register):** write `0x0024` to `0x0157`.
2. **FC03 (Read Holding Registers):** read one register at `0x0157`; confirm
   `0x0024`.
3. **FC06:** write `0x2211` to `0x1801` to save **all current parameters** to
   EEPROM, not just SO1.
4. **FC03:** read one register at `0x1901`. `0x5555` means save successful;
   `0xAAAA` means failure. Success resets to `0x1111` after reading, so capture
   the first result.

**MotionStudio is not required for this configuration.** The driver handles
brake sequencing; the Jetson does not toggle the relay for each move.

### Verified results

- SO1 write/readback/restore test succeeded.
- SO1 was then set to `0x0024` again and read back successfully.
- EEPROM save returned **`0x5555`**, confirming successful saving.
- Both delays read back as **250 ms**.
- Initial current-alarm register `0x2203` read **`0x0000`**.
- No motion commands were sent. The relay and brake were not connected during
  the initial configuration testing. Subsequent wiring and power-cycle results
  are recorded below.

### Subsequent wiring and startup-enable check

The user subsequently reported 24 V across the relay coil, and 24 V from
both Phoenix COM and NO to supply negative while energised. Physical brake
engagement/release has not yet been confirmed.

DI1 initially read `0x0088` (normally-closed Enable), with Pr0.07 at zero
and drive status `0x0032` (enabled, not running). This default enables the
drive at startup and explains the energised brake relay while stationary.

With DI1 confirmed unwired, Pr4.02 was changed to **`0x0008` (normally-open
Enable)** and verified by readback. EEPROM saving returned **`0x5555`**.
After the change, status `0x1003` read **`0x0000` (disabled, not running)**,
current alarm remained zero, DO1 remained `0x0024`, and Pr0.07 remained zero.
No motion command was sent.

The manual requires saving and restarting for input-function configuration.
The user confirmed that the relay indicator went out when the drive disabled.
After the user power-cycled the driver, read-only queries confirmed DI1
`0x0008`, software forced enable `0x0000`, DO1 `0x0024`, both delays at
250 ms, drive status `0x0000` (disabled, not running), and current alarm
`0x0000`. The saved configuration and disabled startup state are therefore
verified. Physical brake holding and release remain to be tested.
With DI1 inactive and software forced enable off, the intended startup state
is disabled with the power-off brake engaged. A deliberate enable action is
required before movement; writing zero to Pr0.07 only returns control to the
input and is not an unconditional disable command.

### Movement checks on 22 October 2026

Tests temporarily forced enable via Pr0.07, waited 2 seconds before a finite
relative move, waited after completion, then returned Pr0.07 to zero. No
EEPROM save was issued during movement tests. DI1 remained unwired.

- 500 pulses at 10 RPM completed with actual position 500.
- 20,000 pulses at 60 RPM completed without alarms, but actual position changed
  from 500 before enabling to 20,400 afterwards (19,900 counts). No position
  sample was taken between enabling and this move, so the 100-count
  discrepancy remains unexplained. The user reported smooth movement and
  relay on before movement/off afterwards.
- 1,400,000 pulses (140 nominal motor revolutions) at 840 RPM, with
  1000 ms/1000 RPM acceleration/deceleration, completed in approximately
  10.96 seconds including command/communication overhead. This represents
  10 seconds of travel at the nominal speed plus ramp time, not a timed jog.
  Actual position was 20,400 both before and after enabling, 1,420,397
  immediately after completion, and 1,420,400 after a one-second settling
  interval. Both settled actual-position and profile-position deltas were
  **exactly 1,400,000**. The earlier 100-count discrepancy did not recur.
  Final disabled status showed profile and actual position both 1,420,400,
  no alarms, and software forced enable zero.
  The user confirmed smooth operation at 840 RPM and that the relay operated
  as expected during the sequence.

These are drive-counter checks, not independent measurements of shaft/rope
travel or mechanical brake timing. The deliberate pauses also do not verify
brake sequencing at the minimum configured delays.

## Phoenix Contact relay wiring

Power off the driver and 24 V supply before wiring. Support the mechanism
before powering up: brake control may release the brake when the drive enables.

| From | To |
| --- | --- |
| Converter **+24 V** | Phoenix **A1+** |
| Phoenix **A2−** | Driver **DO1** |
| Driver **COMO** | Converter **0 V / −** |
| Converter **+24 V** | Phoenix **COM (11)** |
| Phoenix **NO (14)** | Motor **brake +** |
| Motor **brake −** | Converter **0 V / −** |

A link between Phoenix **A1+ and COM** can feed both from the same +24 V
supply. Leave DO2 and DO3 disconnected for this brake circuit.

```text
Relay coil:
24 V + ── A1+ [Phoenix coil] A2− ── DO1 [driver output] COMO ── 24 V −

Brake power:
24 V + ── COM (11) [normally-open contact] NO (14) ── Brake +
24 V − ──────────────────────────────────────────── Brake −
```

**Phoenix COM gets +24 V; driver COMO gets 0 V. They are different nodes.**

- **A1+ and A2− power the coil together.** DO1 switches its return path;
  A2− is not a separate logic-trigger input.
- **COM and NO are voltage-free switch contacts.** Energising the coil joins
  COM to NO. NO gets +24 V because COM is supplied with +24 V.
- Relay energised → brake powered → power-off brake **released**.
- Relay de-energised → COM–NO open → brake unpowered → brake **engaged**.

The driver switches the relay coil, **not the motor brake coil directly**.
Digital outputs are rated 24 V / 100 mA maximum (50 mA recommended).
The relay includes coil suppression; observe A1+/A2− polarity. Provide
suitable suppression across the motor brake as shown in the wiring drawing.
If a flyback diode is approved for the brake, its stripe (cathode) goes to
brake + and its anode to brake −. Verify physical brake timing with the
chosen suppression.

## Reference

[Leadshine CS2RS Series user manual](https://www.leadshine.com/upfiles/downloads/19d35f99f92b99e0417f74134a985aff_1669716112793.pdf):
I/O terminal table, brake output section, Pr4.11–Pr4.20 parameter table,
Modbus function codes and auxiliary control/save-status registers.

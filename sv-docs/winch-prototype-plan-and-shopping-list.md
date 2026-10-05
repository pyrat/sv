# Two-axis camera winch: prototype plan and shopping list

Date: 5 October 2026

## Decision summary

Build and test **one axis first**, using an online-orderable STEPPERONLINE **NEMA 23 closed-loop stepper with encoder and power-off brake**, a directly attached **10:1 planetary gearbox**, and a **CL57T-V41 driver**. Expand to two axes after measuring rope pull and validating winding repeatability.

Use a **Raspberry Pi** for the application and X/Y calculations, a **Raspberry Pi Pico** for coordinated step pulses, and a **ULN2803A** transistor interface to the drivers' 24V control inputs.

This is a prototype selection, **not a validated replacement for the original 16 Nm gearmotor**. Torque capability, shaft loading, brake circuitry and winding mechanics remain to be verified. Prices below were observed during the conversation in USD, excluding shipping and taxes; stock and prices may change.

## Requirements established in this conversation

- Two motors, each driving its own rope drum, for **X/Y positioning of an underwater camera**.
- Motors and drums are **above water on the side of fish cages**, exposed to a marine environment.
- Camera weighs approximately **20 kg in air**, but has flotation and reportedly little effective submerged weight. Actual rope tension and current-induced drag are unknown.
- Rope: approximately **4 mm Dyneema**.
- Target maximum rope speed: approximately **one metre in five seconds**, interpreted as **0.20 m/s**. If five seconds includes acceleration and stopping, the required cruising speed will be higher.
- Prefer reasonably priced, standard components purchasable online without supplier enquiries.
- Prefer a compact motor-plus-gearbox assembly with the drum on the output, rather than external belt reduction.
- Position estimation will use **predictable winding and a calibrated rope-length-versus-drum-turns model**.

The exact two-rope geometry, deployment length, submerged load and required positioning accuracy are still unspecified. Two ropes jointly determine position only with suitable geometry and tension; they do not necessarily correspond to independent X and Y axes.

**Documentation note:** the existing `motor-options-review.md` records a 15 kg camera and a different procurement shortlist. This file records the 20 kg figure and prototype decisions from this conversation; the weight discrepancy should be resolved before final sizing.

## Original motor and drum

### Original Smart Automation motor

Drawing: **180202A-1**, assembly **56JXE300K.63ZYT03B-24V.brake+HS30B-50P-8A-V**.

The visible specification table gives:

| Feature | Original assembly |
| --- | --- |
| Motor type | Appears to be brushed DC; red/black motor leads |
| Supply | 24V DC |
| Reduction | 55:1 |
| Rated motor speed | 4,000 RPM |
| Rated gearbox output speed | 72.7 RPM |
| Rated gearbox output torque | 16 Nm |
| Rated current | 9.6 A |
| Brake | WR220, 24V, 1 Nm, mounted on motor side |
| Encoder | HS30B-50P |
| Output | 12 mm keyed shaft, approximately 28 mm projection |
| Protection | IP50 |

The drawing does not establish whether the original brake is power-applied or power-released. Its 1 Nm rating is not a quoted drum-side holding torque.

Matching shaft and drum dimensions strongly suggest the drum mounted directly on the **gearbox output**, not on the bare motor shaft. An assembly photograph/drawing would confirm this.

### Drum

Drawing: **115669**, “Hoisting Pulley Ø92/200-60”, POM-C black.

- Barrel diameter: **92 mm**.
- Flange diameter: **200 mm**.
- Approximate winding width: **46.6 mm**.
- Overall width: **60 mm**.
- Bore: **12 mm H7**, with **4 mm keyway**.
- Drawing weight: **0.6 kg**.

With 4 mm rope, the first-layer rope centreline diameter is approximately **96 mm**, and each additional layer adds **8 mm**. Assuming 11 evenly packed turns per layer:

| Layers | Approximate cumulative rope capacity | Outermost rope centreline diameter | Drum speed at 0.20 m/s |
| ---: | ---: | ---: | ---: |
| 1 | 3.3 m | 96 mm | 39.8 RPM |
| 5 | 19.4 m | 128 mm | 29.8 RPM |
| 10 | 45.6 m | 168 mm | 22.7 RPM |
| 11 | 51.7 m | 176 mm | 21.7 RPM |

These are geometric estimates, not guaranteed capacity. Rope compression, crossover turns, guide geometry and uneven winding affect the result. At 11 layers the outer rope surface is about 180 mm in diameter, leaving approximately 10 mm radial clearance below the flanges.

## Motor options discussed

- **Initial STEPPERONLINE BLDC:** F5B60-24GN-30S/5GN50K, 24V, 60W, 60 RPM, 6.3 Nm, 50:1 spur gearbox. It is a three-phase brushless DC motor, not a stepper. Its driver receives DC and generates the motor phase currents. No physical brake or equivalent positioning encoder was verified.
- **Direct-drive NEMA 23:** rejected as an established sizing solution. Its 2 Nm holding torque corresponds to only about 42 N on the first rope layer or 23 N near 11 layers, before running-torque limitations and margin.
- **NEMA 34 alternative:** 34E1KBK50-45, 4.5 Nm holding torque, 5 Nm brake, observed price $126.18 versus $92.16 for the NEMA 23. Larger running torque might help, but direct-drive suitability was not established.
- **Twirl/Parvalux:** relevant geared motor families with brake/encoder options, but no confirmed suitably priced, ready-to-order complete configuration was selected.
- **AliExpress:** listings advertising geared motors with encoders and optional brakes were found, but variants, specifications and complete prices could not be verified. None was selected for purchase.
- **Selected prototype:** NEMA 23 motor with encoder/brake plus a bolt-on gearbox. This gives a compact assembly while keeping motor and gearbox separately replaceable.

A **planetary gearbox is not a holding brake**. It increases torque and reduces speed but can generally be back-driven. A physical power-off brake holds the mechanism without motor power; electronic braking is a different function.

## Selected motor axis

### Motor: 23E1KBK20-20

- NEMA 23, 57 × 57 mm frame.
- Two-phase stepper; **2 Nm holding torque**, **5 A per phase**.
- **8 mm D-shaft**, 21 mm projection, 15 mm flat; no keyway.
- **1,000 PPR differential incremental encoder**.
- **24V, 2 Nm power-off brake**, listed approximately 0.16 A / 4 W.
- IP40: requires environmental protection on the cage.

### Gearbox: EG23-G10-D8

- Inline **10:1 planetary reduction**, listed efficiency **94%**.
- NEMA 23 input, built-in input coupling.
- Accepts **8 mm input shafts, 12–31 mm long**, matching the motor's published shaft dimensions.
- **14 mm keyed output**, 35 mm projection.
- **10 Nm maximum permissible torque**, **20 Nm momentary permissible torque**.
- Listed backlash up to 15 arc-min; IP54 gearbox rating does not protect the whole motor assembly.

The motor turns approximately **220–400 RPM** for the expected drum-speed range. Verify the motor torque curve at the selected supply voltage. Output capability is limited by both running motor torque and the gearbox rating; do not infer 20 Nm continuous output from 2 Nm holding torque × 10.

### Driver: CL57T-V41

- One driver per motor.
- Recommended supply range **24–48V DC**, with 36V listed as typical.
- Up to **5 A RMS / 8 A peak**; configure using the manufacturer's current convention.
- STEP/DIR control, encoder input, alarm output and brake-control output.
- Supports 1,000 PPR feedback; selectable 5V/24V command-input logic.
- RS232 configuration interface; configuration cable listed as included.

The motor encoder connects to the driver. STEP/DIR does **not** return actual encoder position to the Pi. Track commanded pulses, monitor faults and home to a reference; independently measured position telemetry would require an additional documented interface. Do not assume the RS232 configuration port supplies real-time position feedback.

## Mechanical integration

Preferred compact layout:

```text
Motor with encoder and brake → attached planetary gearbox → drum
```

- Existing drum bore is **12 mm**, selected gearbox output is **14 mm**: design a suitable modified hub/bore or separate shaft connection.
- Verify key dimensions, engagement length and axial retention from the drawings.
- Confirm gearbox allowable radial load at the actual drum overhang before mounting directly.
- If necessary, use a coupling and bearing-supported drum shaft; this remains 1:1 after the gearbox.
- Use a replaceable mounting plate and accessible connectors to facilitate future component changes.
- NEMA mounting improves replaceability but does not standardise every shaft, encoder pinout or electrical requirement.

## Position model and winding

1. Establish a repeatable home/reference position for each axis.
2. Guide rope into predictable layers, ideally using a level-wind mechanism advancing approximately one rope diameter per drum revolution.
3. Maintain suitable winding tension to avoid loose wraps and rope burying.
4. Calibrate each drum separately by recording measured deployed length against measured/known drum rotation under representative tension.
5. Interpolate between calibration points; test payout and retrieval repeatability.
6. Convert requested X/Y into rope lengths using the actual mounting geometry, then use the calibration model to obtain motor-turn targets.
7. Recover a known reference after power loss. Saved commanded position alone does not establish actual position after unpowered movement.

The level-wind mechanism and geometric X/Y model have not yet been designed. Gearbox backlash, rope stretch and current-induced deflection also affect camera position.

## Raspberry Pi integration

```text
Raspberry Pi: UI, networking, X/Y geometry and winding calibration
    │ USB serial
    ▼
Raspberry Pi Pico: coordinated pulses, acceleration, homing and faults
    │ 3.3V GPIO
    ▼
ULN2803A: open-collector STEP/DIR/ENABLE interface
    │ 24V command signalling
    ├── CL57T-V41 A → motor A; encoder A → driver A
    └── CL57T-V41 B → motor B; encoder B → driver B

24V supply → appropriate brake-switching circuits → holding brakes
```

The Pico provides deterministic pulse generation; ordinary Linux GPIO bit-banging on the Pi is not the planned motion source. One Pico can control both axes.

Example command resolution: 1,600 pulses per motor revolution with 10:1 reduction gives **16,000 pulses per drum revolution**. At the 96 mm first-layer diameter this is approximately **53,000 pulses/metre**, or **10.6 kHz at 0.20 m/s**. This is command resolution, not guaranteed physical accuracy.

### Selected signal interface: ULN2803A

[Adafruit product 970](https://www.adafruit.com/product/970) is a through-hole eight-channel transistor-array chip, **not a complete terminal board**. It accepts 3.3V inputs and provides open-collector outputs rated up to 50V.

Use the CL57T's **24V logic setting**. The Darlington output voltage drop makes this preferable to the previously discussed 5V arrangement for this chosen interface.

| Prototype connection | Destination |
| --- | --- |
| Pico GP2 | ULN2803A IN1 |
| ULN2803A OUT1 | Driver PUL− |
| Pico GP3 | ULN2803A IN2 |
| ULN2803A OUT2 | Driver DIR− |
| Pico GP4 | ULN2803A IN3 |
| ULN2803A OUT3 | Driver ENA− |
| Regulated +24V control supply | Driver PUL+, DIR+, ENA+ |
| Pico GND and 24V supply negative | ULN2803A GND |

- Add **10 kΩ pull-downs** on used ULN inputs for a defined startup state.
- For these non-inductive signal loads, **COM can remain unconnected**. It is the clamp-diode common, not a chip supply pin.
- Never connect 24V to Pico GPIO.
- Verify pulse width, signal polarity and enable behaviour on the bench. An inactive transistor output must not be assumed to mean the driver is disabled.
- One chip has enough channels for STEP/DIR/ENABLE on both axes.
- Brake switching and 3.3V-compatible alarm feedback are separate circuits still to finalise.

### Brake sequence

Establish motor holding torque, release the brake, wait its specified release time, then move. Decelerate to rest and engage the brake before removing holding torque. Use the documented CL57T brake-output interface and timings; do not assume it can power the brake coil directly. Emergency stopping while moving requires a suitable dynamic brake rating, not just a static holding rating.

## Shopping list: one-axis prototype

### A. Motor, gearbox and driver — selected products

| Qty | Item | Model / product link | Observed price |
| ---: | --- | --- | ---: |
| 1 | Motor, encoder and power-off brake | [23E1KBK20-20](https://www.omc-stepperonline.com/p-series-nema-23-closed-loop-stepper-motor-2nm-283-28oz-in-with-electromagnetic-brake-23e1kbk20-20) | $92.16 |
| 1 | Bolt-on 10:1 gearbox | [EG23-G10-D8](https://www.omc-stepperonline.com/eg-series-planetary-gearbox-gear-ratio-10-1-backlash-15-arc-min-for-8mm-shaft-nema-23-stepper-motor-eg23-g10-d8) | $34.09 |
| 1 | Closed-loop driver | [CL57T-V41](https://www.omc-stepperonline.com/stepper-motor-driver/closed-loop-stepper-driver-v4-1-0-8-0a-24-48vdc-for-nema-17-23-24-stepper-motor-cl57t-v41) | $36.39 |
| | **One-axis subtotal** | Excludes the remaining items below | **$162.64** |

### B. Control electronics

| Qty | Item | Specification |
| ---: | --- | --- |
| 1 | Raspberry Pi | Pi 4 Model B, 2GB or more; reuse an existing suitable Pi |
| 1 | Pi power supply | Correct official supply for chosen model |
| 1 | microSD card | 32GB, preferably high-endurance |
| 1 | Raspberry Pi Pico H | Original Pico with headers fitted |
| 1 | USB data cable | USB-A to Micro-USB for Pi/Pico |
| 1 | ULN2803A | [Adafruit 970](https://www.adafruit.com/product/970) |
| 1 | DIP socket | 18-pin, for the ULN2803A |
| 1 | Prototyping board | Solderable perfboard with room for connectors |
| 6 | Resistors | 10 kΩ, ¼W; sufficient for two axes |
| 1 set | Headers, terminals and hookup wire | For controller/interface assembly |

### C. Power, brake and wiring — some selections provisional

| Qty | Item | Specification / selection status |
| ---: | --- | --- |
| 1 | Motor-driver supply | **36V DC, approximately 150–200W**, provisional single-axis bench sizing; verify against operating load |
| 1 | Brake/control supply | Regulated **24V DC, at least 1A** for prototype |
| 1 | Brake-switching interface | **Exact part/circuit pending** CL57T output review; switch approximately 0.16A at 24V, with appropriate coil suppression |
| 1 | Alarm-input interface | **Exact circuit pending**; adapt driver alarm output to 3.3V Pico inputs |
| 1 | Home switch | Sealed mechanical switch with repeatable actuator arrangement |
| As needed | Motor extension | Matching connectors/pinout and conductor size for 5A phase current |
| As needed | Encoder extension | Shielded, matching pinout; supplied motor leads are short |
| As needed | Brake cable | Two-core cable suitable for installation |
| 1 set | Distribution hardware | Terminal blocks, appropriately sized fuses, ferrules, glands and disconnect |
| As needed | Regenerative-energy handling | Assess bus-voltage rise during load-driven motion; an ordinary DC supply generally cannot absorb returned energy |
| 1 | Protective enclosure | For cage installation, suitable for salt spray/rain with thermal management; motor protection also required |

The brake/alarm circuits, final power-system design and protective hardware are **not yet a completed electrical BOM**. A motor brake's 24V rating does not require the motor driver itself to run on 24V.

### D. Mechanical and test equipment

| Qty | Item | Specification / selection status |
| ---: | --- | --- |
| 1 | Drum | Engineered modification/new hub for **14 mm keyed output**, or separate shaft connection |
| 1 | Mounting plate | Removable motor/gearbox mounting plate |
| 1 set | Hardware | Drawing-matched bolts, key and shaft retention |
| If required | Drum shaft, bearings and coupling | Based on permissible gearbox output-shaft load |
| 1 | Rope guide / level wind | **Mechanism still to design** for calibrated layering |
| As needed | Rope | 4 mm Dyneema; test/deployment length to establish |
| 1 | Inline force gauge | Approximately **0–500N**, with useful low-tension resolution |

## First order and validation sequence

1. Order the selected motor, gearbox and driver; Pico H; USB data cable; ULN2803A and prototyping parts. Add Pi/accessories if needed and a force gauge.
2. Finalise the brake/alarm interfaces, power arrangement and shaft/drum mounting.
3. Test motor operation, encoder fault handling, homing and brake sequencing without the camera load.
4. Measure representative camera rope tension, including hauling at the intended speed and relevant rope angles/current conditions.
5. Validate motor running torque, transmission temperature, output-shaft loading and braking against measured requirements.
6. Build repeatable winding and calibrate rope length in both directions.
7. Add the second axis and validate coordinated motion using the actual X/Y geometry.

## Expansion to two axes

Add a second motor, gearbox and driver: **another $162.64**, giving **$325.28** for both motor/gearbox/driver sets at observed prices.

Also add the second drum assembly, home switch, cables, brake switching and fault input, and size the supplies for simultaneous motion. The same **Pi, Pico and ULN2803A** serve both axes.

## Open decisions

- Actual submerged camera weight, rope tensions and current drag.
- Exact two-rope geometry and means of maintaining tension across the workspace.
- Deployment length, positioning accuracy and duty cycle.
- Verified torque-speed performance at the selected voltage; 10 Nm gearbox ceiling.
- Gearbox radial/overhung-load allowance and drum hub adaptation.
- Level-wind mechanism and calibration procedure.
- Final brake/alarm circuits, pulse timing and power-loss behaviour.
- Regenerative-energy management and marine enclosure design.

**Bottom line:** the selected NEMA 23 assembly is a reasonably priced, modular test platform. The next step is one instrumented axis, not an assumption that it already reproduces the old 16 Nm winch's performance.

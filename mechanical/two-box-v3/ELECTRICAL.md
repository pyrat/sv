# One-cable interconnect — iteration 3

## Architecture

```text
POWER / COMPUTE BOX                     WINCH BOX
Mains -> fused NDR-240-48 -> 48 V --------+-> fused driver A -> motor A
                          hybrid cable  +-> fused driver B -> motor B
                                        +-> 48-to-24 V converter
                                             -> relay K1 -> brake A
                                             -> relay K2 -> brake B
       -> DDR-60L-12 -> Jetson
                        -> USB-RS485 =======> driver A -> driver B
```

The inter-box link is **one 2 metre outer-jacket hybrid cable**, with power conductors
and a shielded twisted data pair. The 48 V and data connections terminate in
one suitably rated sealed hybrid connector at each box. **Confirmed: the 2 m
cable unplugs at both containers.** Provide two panel-mounted bulkhead connectors
and two mating cable-end connectors, with strain relief at both cable ends.
Select the connector family before cutting the provided hybrid interface plates.
Provide sealing caps for both bulkheads when disconnected; the selected
connector's mated rating does not automatically apply when unplugged.
Use recessed/touch-protected contacts on the supply-side connection and keying
that prevents incorrect mating. Final connector parts and pin assignments remain
to be selected; this confirmation does not change the blank-plate CAD.

RJ45 (rather than RG45) describes the plug used by some drivers. Their Modbus
RTU interface is **RS485, not Ethernet**. Keep driver-specific RJ45 leads inside
the winch box. Do not assume an Ethernet cable or standard RJ45 connector is
suitable for supplying both motors. No PoE arrangement is assumed.

## Provisional conductor functions — not connector pin numbers

| Function | Requirement |
| --- | --- |
| +48 V | Power conductor sized for total current, cable length and temperature |
| DC return | Same power-current capability as +48 V |
| RS485 D+ / D− | Shielded twisted pair with suitable characteristic impedance |
| Signal reference | Separate conductor if required by selected transceivers/drivers |
| Protective/bond conductor | Include in the same cable if the bonding design requires it |
| Shield/drain | EMC termination per connector/drive installation guidance |

Do not use the data shield as a power return or protective-earth conductor.
Use the actual manufacturer's A/B or D+/D− mapping: naming varies. Do not assign
connector pin numbers until the connector and both driver variants are settled.
Choose a keyed connector with sufficient DC voltage/current rating, environmental
rating in the mated state and power/data separation. Unless explicitly rated for
it, disconnect power before mating/unmating.

## Modbus and brake control

- One RS485 master at the Jetson, two unique slave addresses and a linear
  daisy-chain through the drivers; avoid a star.
- Match baud/parity settings; terminate the two physical bus ends according to
  device manuals. Avoid adding duplicate built-in termination or bias networks.
- Confirm whether the chosen USB adapter/driver interface provides isolation and
  what signal-reference wiring it requires.
- The relays do not themselves provide Modbus control. Wire K1/K2 coils through
  suitable local driver outputs **only after checking output function, voltage,
  current and brake sequencing support**. Otherwise add a local Modbus I/O
  module and reserve mounting space; it is not currently modelled.
- Verify the brake's energised/released behaviour, suppression requirements and
  sequencing with motor holding torque. Configure communication-loss behaviour
  deliberately rather than relying on a software default.

## Load and cable selection

The existing NDR-240-48 is nominally **48 V, 5 A, 240 W**. That budget includes
both drives and the Jetson/converter loads, not 5 A available solely to the
winch. Measure or calculate simultaneous running/startup demand. Driver phase
current is not the same as DC input current. Check drive bus voltage limits and
regenerative energy during deceleration; the PSU is not assumed to absorb it.

The provisional DDR-15L-24 supplies about **24 V, 0.63 A (15 W)**. Confirm that
both brake coils **plus relay coils** fit its continuous and transient budget.
If not, use a larger converter and update its mounting envelope. The two relay
contacts must be rated for the actual inductive brake load.

For a copper power pair, an initial voltage-drop estimate is:

`drop ≈ 2 × one-way length(m) × current(A) × 0.0175 / area(mm²)`

For the confirmed **2 m one-way cable length**, a provisional 1.5 mm² copper
power pair at 5 A gives approximately **0.23 V drop (0.49% of 48 V)** and
**1.17 W total pair loss** at 20 °C, before connector losses. The 5 A calculation
is a sizing reference, not a measured winch load. Final cable selection still
depends on ampacity, flexing, ambient temperature, insulation, UV/water resistance,
bundle heating and connector contact resistance must also be checked. Put
appropriate circuit protection at the source and local branches.

## Inputs needed to finish the cable/interface design

1. Whether the confirmed 2 m cable flexes during operation.
2. Actual motor/driver variants and both brake coil ratings.
3. Saltwater exposure, if applicable; detachable connections at both ends are confirmed.
4. Total measured load and the grounding/isolation arrangement.

These determine cable construction, connector part, fuse sizes and final
interface cutouts. The enclosure currently supplies replaceable blank plates
so those choices do not require reprinting its main bodies.

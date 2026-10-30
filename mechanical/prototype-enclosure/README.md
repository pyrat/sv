# DepthPilot prototype enclosure — v1

Printable panel enclosure for two external **110 mm demonstration drums**, two
internal geared motors/drivers, Jetson Orin Nano developer kit, NDR-240-48 PSU,
12 V / 24 V converters and brake relays. Units: **millimetres**.

**Assembled shell: 440 W × 460 D × 228 H.** This is a spacious first-fit layout,
not a compact production enclosure. All printable pieces fit within 230 × 230 mm
in their supplied orientations. The Centauri Carbon is 256 × 256 × 256 mm;
there is room for a 5 mm brim around the largest panels.

## Start here

- [`layout.svg`](layout.svg): dimensioned top/front arrangement.
- [`layout.glb`](layout.glb): coloured 3D shell, external drums and component envelopes, exported
  in standard glTF metres (0.440 × 0.460 × 0.228 m).
- [`stl/`](stl/): individual **printable** parts, already oriented on the bed.
- [`assembly-reference.stl`](assembly-reference.stl): assembled shell for CAD
  reference only; **do not print this as one object**.
- [`component-envelopes.json`](component-envelopes.json): component positions.
- [`validation.json`](validation.json): part quantities, bounds, volumes and mesh checks.
- [`generate.py`](generate.py): editable parametric constructive-solid-geometry source.

The model is a **first-fit enclosure kit**: shell and joining hardware are modelled;
equipment mounting holes, cable-entry cutouts and terminal guarding must be
finished against the actual hardware. Component blocks are space reservations,
not vendor CAD. The source checks these blocks against the shell and each other;
it does not validate cable routing, fastener assemblies or thermal performance.

## Layout

Coordinates: X left-to-right, Y front-to-back, Z upwards. Floor occupies Z=0–4,
roof Z=224–228. Front panels occupy Y=0–4, rear panels Y=456–460. Side panels
fit between them, at X=0–4 and X=436–440. Floor/roof have four tiles each.

| Feature | Position / size |
| --- | --- |
| Output shaft axes | X=150 and 290, Z=100; point out through the front |
| Front clearance openings | 64 × 64, for flush 57 mm gearbox faces; **not** mounting patterns |
| External demonstration drums | Ø110 × 50; 140 mm centre spacing, 30 mm flange gap |
| Motor/gearbox reservations | 57 × 195 × 57 each, Y=0–195 |
| Driver reservations | 45 × 155 × 120 each, alongside the motors |
| Jetson + connectors reservation | 145 × 130 × 85, X=75–220, Y=210–340 |
| PSU | 63 W × 113.5 D × 125.2 H; X=345, Y=326.5, Z=40 |
| DIN rail mounting centres | X=20,110,180 and 240,330,400; Z=100 on rear panels |

The motor envelope uses the BOM's **116 mm Stepperonline motor + 78.8 mm
MGL23-G20-D8 gearbox**. The older MG gearbox and Leadshine alternative are also
in the BOM; check the actual pair before fixing mounting holes. Driver sizes are
deliberately allocated envelopes, not confirmed dimensions of either driver.

The PSU is upright with 36 mm below and 58.8 mm above to the inner shell,
exceeding the manufacturer's 20/40 mm bottom/top guidance. It is separated from
the Jetson by 125 mm in X, and from the motor/driver region by at least 125.5 mm
in Y. Keep that empty space available for airflow. Wiring and guards must not
fill those ventilation gaps. Converter envelopes still need comparison with
their vendor drawings, including DIN clip and terminal access.

## Print list

| STL name | Quantity | Notes |
| --- | ---: | --- |
| `floor-tile` | 4 | 220 × 230 × 4 |
| `roof-tile` | 4 | Same footprint, ventilation slots |
| `front-left`, `front-right` | 1 each | 220 × 220 × 4; different shaft positions |
| `rear-tile` | 2 | 220 × 220 × 4; solid rail-mounting band |
| `side-tile` | 4 | 226 × 220 × 4 |
| `side-fan-tile` | optional | Replaces one side tile; 80 mm fan, 71.5 mm hole pitch |
| `seam-splice` | 24 | Two per tile-seam segment |
| `corner-angle` | 40 | Two per perimeter segment and vertical corner |
| `rail-spacer` | 6 | Ø16 × 8, M4 clearance bore |
| `equipment-pad` | 2 optional | 160 × 120 × 6; drill to suit equipment |
| `fit-coupon` | 1 first | Check M4 holes and printing behaviour |

Suggested PETG settings: 0.4 mm nozzle, 0.2 mm layers, 5 walls, 5 top/bottom
layers, 25–35% gyroid. Use 100% infill for angles/splices/spacers. Print all
parts exactly as supplied, 100% scale. Panels lie flat; angles stand on their
L-shaped ends. Inspect the small horizontal slot bridges on angles in the slicer.
Use a brim for large panels if needed. A 4 mm PETG sheet can warp; print one
floor tile and a joint sample before committing to the entire kit.

## Purchased hardware

- About **256 M4 × 16 bolts, nuts and pairs of washers** for the full shell:
  96 at 24 splices and 160 at 40 angles. Buy spares. Nuts are accessible from
  inside; no heat-set inserts or printed threads are used. Check actual washer
  stack and choose shorter bolts where protrusion would interfere.
- Six M4 × approximately 25 bolts for rear DIN rails, spacers, washers and nuts;
  confirm against the rail profile. Two lengths of TS35/7.5 rail, about **200 mm**
  each, spanning X=10–210 and X=230–430. Drill rail web at the panel hole centres.
- Metal motor/gearbox brackets and a rigid metal mounting base. A nominal
  **400 × 210 × 3 mm aluminium plate** fits the front bay, X=20–420, Y=10–220.
  Drill to suit the actual brackets and support it above the floor joints.
  Set bracket elevation so shafts land at Z=100. The BOM's ST-M9 fit remains
  unconfirmed. Printed front panels provide clearance, not motor support.
- Jetson mounting standoffs, driver fasteners and wiring clips to suit hardware.
- Optional 80 mm enclosure fan and matching fasteners. Its voltage must match
  its supply. Mount on a rear side panel; verify fan depth clears the components.
- Suitable fused mains inlet, strain relief, protective-earth connections and
  a terminal cover selected to the actual PSU. Inlet opening is intentionally
  drill-to-fit in an unvented region after selecting the inlet.

The printed panels are enclosure skins; motor torque is carried by the metal
brackets/base. This matches the existing loose-rope demonstration scope.

## Assembly

1. **Fit trial:** print one coupon, one splice and one angle. Check M4 clearance,
   nut/washer access, slot bridging and flatness. Coupon's centre Ø14.3 hole is
   a print-size reference, not the production shaft opening.
2. Lay four floor tiles as a 440 × 460 rectangle. Seams are X=220 and Y=230.
   Install splices on the upper/inside surface, at stations 48 mm from each end
   of each seam segment. A splice has holes 28 mm across the seam and 16 mm
   along it, matching the panels' holes 14 mm in from each edge.
3. Join each pair of front, rear and side panels with two internal splices.
   Front shaft holes must be at global X=150 and 290, Z=100. Side panels are
   226 mm long, so two span Y=4–456. Floor and roof edges remain flush.
4. Fit internal corner angles: two at each of the eight floor perimeter
   segments, two at each of eight roof perimeter segments, and two at each of
   four upright corners. Use stations 48 mm from segment ends. Extra centre
   stations are available but not required by this print count.
5. **Angle orientation matters:** the leg with holes 10 mm from the corner goes
   against the overlapping panel (floor/roof, or front/rear at upright corners).
   The leg with holes 14 mm from the corner goes against the inset wall. Slots
   accommodate the 4 mm edge/station offset where side walls meet the floor.
   Select the matching pair of panel holes; do not force panels into alignment.
6. Install motor support/base and test gearbox clearance through the
   64 mm square front openings. Set the stationary output face flush with the
   outside front surface at Y=0, so the panel does not consume shaft engagement.
   Drill hardware mounting holes with parts in position. Keep
   floor-joint hardware clear underneath the metal plate using spacers. Check
   shaft projection before fitting drums: the nominal 34 mm shaft with the drum
   rear face at Y=-3 leaves 31 mm insertion, matching the 30 mm printed clamp.
   If the actual shaft projection is measured from a different boss datum,
   adjust bracket position accordingly. Drum envelopes extend to Y=-53, giving
   a nominal 513 mm overall depth including drums.
7. Install rear DIN rails on the 8 mm spacers: rear inner surface Y=456,
   spacer front Y=448, 7.5 mm rail front approximately Y=440.5. The component
   rear planes at Y=440 are nominal; adjust for actual clips and rail geometry.
   Keep all PSU terminal and ventilation clearances accessible.
8. Position the Jetson/drivers using the layout. Drill equipment pads only
   after confirming hole patterns. Add fastening holes to the floor as needed;
   do not rely on loose pads. Route signal wiring separately from mains/motor
   wiring, and add a fitted cover over the PSU mains terminals.
9. Assemble the roof's four tiles with eight internal splices. It can lift off
   as one assembly after removing the roof-to-angle bolts. Fit roof nuts on
   the inside while the assembly is accessible; secure nuts to angles if regular
   servicing is expected. Check airflow with the lid fitted during first operation.

Roof, shell and vent slots do not provide a water/dust ingress rating. This
version is a ventilated indoor bench-prototype enclosure.

## Checks before committing the full print

Verify actual motor/gearbox lengths, gearbox boss and bracket pattern, both
driver dimensions, Jetson carrier/heatsink and connector access, and converter
dimensions. Cable bends and excess 1.7 m harness length need physical routing;
the envelope check does not model those cables. Fit the drums with their
documented gearbox-face gap and confirm that the enclosure front does not
consume the shaft engagement needed by their clamps.

## Regenerate

```sh
python3 -m venv /path/to/enclosure-env
/path/to/enclosure-env/bin/pip install -r mechanical/prototype-enclosure/requirements.txt
/path/to/enclosure-env/bin/python mechanical/prototype-enclosure/generate.py
```

Generator checks: each print is one connected manifold, exported STL is
watertight with positive volume, all print dimensions ≤240 mm, all print bases
at/above Z=0, and component blocks clear shell panels and each other. Reference
assembly contains separate touching panels deliberately. GLB omits joints,
fasteners, metal base, rails and cables; drums are cylindrical envelopes.

Dimension references:
- [Mean Well NDR-240 datasheet](https://www.meanwell.com/upload/pdf/NDR-240/NDR-240-SPEC.PDF)
- [NDR installation clearances](https://www.meanwell.com/Upload/PDF/NDR%20DIN%20rail.pdf)
- DepthPilot Proto1 BOM: NDR-240-48, 23HS22-4004D-E1000-B280 and MGL23-G20-D8.
- [`../prototype-drums/ONE-PIECE.md`](../prototype-drums/ONE-PIECE.md): demonstration drums.

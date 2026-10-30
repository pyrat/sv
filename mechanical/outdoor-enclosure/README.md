# Outdoor enclosure development kit — TPU static seals

**Design intent: sustained outdoor rain and splashing, with drums outside.**
This is a new enclosure architecture, not a waterproofing patch to the vented
v1 shell. It is **not yet an operational weatherproof winch enclosure**: the
shaft interfaces and heat rejection still require engineered hardware. The
supplied closed boxes and blank covers can be printed and water-tested now.
No IP rating is claimed; a watertight STL is a mesh property, not a water test.

## Files

- [Layout and gasket section](layout.svg)
- [3D assembly](assembly.glb), standard glTF metres; STL/source dimensions are mm
- [Printable rigid parts and TPU gaskets](stl/)
- [Validation results](validation.json)
- [Reproducible CAD generator](generate.py)
- [Complete ZIP](../outdoor-enclosure-development.zip)

## Four independent cells

Each body prints as one piece, **230 × 230 × 220 mm**, open top upward. Each
solid lid is **238 × 238 mm**, with a 10 mm overlapping rain skirt. Closed cell
height is 226 mm. These parts fit the Centauri Carbon's 256 mm build area:
a 5 mm brim around a lid occupies 248 mm. Check the actual slicer bed exclusions.

The four cells stand in a 2 × 2 arrangement on a shared rigid mounting frame,
with 30 mm service gaps. Overall bodies occupy **490 × 490 mm**; lids occupy
**498 × 498 mm**. This excludes external drums, shaft hardware, glands and sinks.

| Cell | Intended contents | Interface facing |
| --- | --- | --- |
| Front left | Motor/gearbox A and driver | Front: shaft cartridge |
| Front right | Motor/gearbox B and driver | Front: shaft cartridge |
| Rear left | Jetson and 12 V converter | Rear: heat-transfer hardware |
| Rear right | 48 V PSU, 24 V converter and brake relays | Rear: heat-transfer hardware |

This is a functional allocation, **not a verified component packing model**.
The 195 mm motor/gearbox envelope leaves about 17 mm between its back and the
rear wall when starting behind the 12 mm front interface. That is tight for
cables and may need a right-angle connector or a different motor location.
There is a 186 × 186 mm top opening through the reinforced lid rim. Check that
equipment can be inserted before choosing its mounting position.

The former single-plane DIN layout does not transfer unchanged: rear cells
have separate mounting systems. PSU installation clearances and orientation
still apply. Use metal internal brackets fixed to the blind floor bosses for
DIN rails; there are no through-wall DIN screws.

Independent boxes eliminate the cross-shaped seam and gasket junctions of
the panel enclosure. Electrical links run through **purchased weather-rated
cable glands**, preferably downward with drip loops, under the frame. Select
glands for the actual cable diameter and wall thickness before drilling holes.
Do not route a cable through a lid gasket. The GLB omits cable links and glands.

## TPU gasket design

Each lid and interface uses a **single continuous printed loop**, with no glued
end joint. The nominal cross-section is:

| Property | Dimension |
| --- | ---: |
| Rectangular groove width | 4.0 mm |
| Groove depth | 2.0 mm |
| Free gasket width | 2.8 mm |
| Free gasket height | 2.4 mm |
| Nominal height reduction at seated hard stops | 16.7% |
| Gasket cross-section / available groove area | 84% |

The remaining groove width lets the TPU spread laterally. The lid/cover lands
seat directly against the body and act as compression stops. Torque is **not**
the method for setting squeeze. Eight M4 screws surround each seal, with blind
insert pockets outside its perimeter. Blind holes avoid a screw passage from
outside into the dry volume. Lid and cover screw pitch is a prototype choice;
the full-size seal needs testing for deflection between screws.

95A TPU may need considerable clamping force compared with a soft silicone or
EPDM seal. Its printed porosity, surface finish and long-term compression set
are unknown until tested. Actual hardness and filament chemistry should be
recorded. For sustained wet exposure, check the filament maker's hydrolysis and
UV data; polyether TPU generally offers better hydrolysis resistance than
polyester TPU. The geometry can also accept a suitable continuous commercial
seal after adjusting the groove to its supplier's squeeze guidance.

### Print the seal-test kit first

Print `seal-test-base`, `seal-test-lid`, and the three TPU test rings:
**2.2, 2.4 and 2.6 mm** high (approximately 9%, 17%, 23% squeeze into a 2 mm
groove). Their width and groove cross-section match the enclosure.

1. Fit the chosen M4 inserts in the base. Four pockets are nominal Ø5.8 × 6 mm.
   Match insert OD/length to this model, or change `INSERT_D` before printing.
2. Place a dry absorbent witness in the central recess. Centre a gasket in the
   groove: its outside edge is 11.6 mm in from the test base's edges.
3. Fit lid and four M4 bolts/washers. With a 6 mm lid and about 1 mm washer,
   M4 × 12 is a starting length, giving about 5 mm insertion. Check that bolts
   cannot bottom out before the lands seat. Tighten evenly and gently.
4. Test with electronics absent. Spray from above and from the sides, then
   dry the exterior before opening and inspecting the witness. A coupon pass
   establishes only a starting gasket profile, not an enclosure IP rating.
5. Leave compressed, repeat the trial after several days and after warm/cool
   cycles. Inspect for permanent flattening, creep and moisture. Test the
   production-size lid and all interface seals separately before installation.

TPU starting print settings: dry filament, 0.2 mm layers, solid infill, slow
manufacturer-recommended speed, enough perimeters to make the narrow ring
solid. Inspect for pinholes and layer gaps. Do not use vase mode. Place the seam
away from corners and inspect it closely. Do not stretch the ring into place.

## Rigid parts and hardware

| STL | Quantity | Purpose |
| --- | ---: | --- |
| `cell-body` | 4 | Monolithic body, blind mounting pockets, gasket grooves |
| `rain-skirt-lid` | 4 | Solid roof and overlapping skirt |
| `lid-gasket-tpu` | 4 | Continuous TPU lid seal |
| `interface-blank-cap` | 4 | Temporary closure for leak testing |
| `interface-gasket-tpu` | 4 | Continuous static seal at interface |
| `seal-test-base`, `seal-test-lid` | 1 each | Small compression/water-test chamber |
| `seal-test-gasket-*.stl` | 1 each | Compare three squeeze levels |

ASA is preferred for the outdoor rigid shells because of UV and temperature
exposure; use its appropriate enclosed-printer process and ventilation. PETG
can be used for fit/water trials, with outdoor suitability depending on the
actual grade. Do not rely on raw FDM walls being impermeable: use a compatible
exterior sealing coating after surface preparation and test it. Keep coating
off the precision gasket lands/grooves unless its thickness is accounted for.

Suggested rigid-part starting point: 0.2 mm layers, 8 walls, 8 top/bottom layers;
check the sliced 6 mm walls are solid. Use local solid modifiers around grooves,
insert pockets and bosses. Body is upright; lid prints with its broad top face
on the bed and skirt upward. Interface cap prints flat. The 100 mm square side
opening **needs support under its horizontal roof**, and the reinforced rim
also needs appropriate support or tuned overhang handling. Keep support scars
away from the sealing grooves; clean and inspect the front groove carefully.

M4 heat-set insert pockets in the top rim and interface are **6 mm deep**;
underside pockets are **5 mm deep** and retain 3 mm of floor above. Use shorter
inserts than the pocket and avoid overheating through the blind floor. Four
underside pockets at (30,30), (200,30), (30,200), (200,200) attach each cell to a
common frame. These are enclosure fixings, not the winch load path.

Each cell has nine internal Ø14 mm bosses at X=35/115/195, Y=40/115/190,
Z=8–14. Their Ø3.4 mm pilot holes end at Z=9. Use selected short fasteners or
adapt the CAD for inserts; **do not drill through** the floor. Support motor
loads on a separately engineered metal frame/base, with any structural
penetration sealed by a designed bulkhead connection.

For four lids and four caps, allow 64 M4 bolts/washers and 64 suitable inserts,
plus 16 shorter underside inserts/fixings and four test-kit inserts/fixings.
Use corrosion-resistant hardware appropriate to the environment. Exact bolt
length must suit the chosen insert and washer stack. Caps and lids are 6 mm
thick; approximately M4 × 12 is the initial stack, not a universal specification.

## Shaft interface: static TPU, purchased rotary seals

Each body has a **100 × 100 mm clear opening**, centred at local X=115, Z=105.
The front reinforcing pad is 140 × 140 mm and 12 mm thick. The removable cap is
136 × 136 × 6 mm; eight screw centres are defined by `port_bolts` in the source.
Its TPU gasket is outside the opening. This is a common interface, **not a
finished shaft seal or a motor support**.

Replace the front caps with machined metal plates/cartridges holding suitable
rotary lip seals, and, if needed, external bearings and an extended shaft.
Choose seal material and shaft finish for the environment and rotational speed.
**A printed TPU ring is not a rotary shaft seal.** A keyed gearbox shaft cannot
be assumed to provide a continuous smooth sealing land; its short projection
must also accommodate the existing 30 mm drum clamp. The final design may need
a smooth seal sleeve, different drum hub or extended supported output shaft.
Confirm the actual gearbox boss, keyway extent and shaft dimensions before
machining or committing to the front-cell arrangement.

The gearbox's stated IP54 does not itself establish the enclosure's outdoor
rating, and the existing 64 mm square v1 clearance openings are not used here.

## Cooling in a closed box

There are **no open fan vents**. An internal circulation fan can move heat to
the wall but cannot remove heat from a sealed plastic box. The rear interface
caps are temporary closures: replace or enlarge them for gasketed aluminium
heat-transfer plates with external sinks as the measured load requires. Motor
cells also need an engineered conductive path to the external frame or their
own heat exchanger; the rear sinks cannot cool isolated front cells by airflow.

The BOM records up to 28 W for the complete Jetson. At 240 W output and 90%
efficiency, the PSU alone would dissipate about **27 W**, before converters,
drivers and motors. Do not treat the 240 W electrical output as enclosure heat,
and do not assume the device's full-rated load is the operating point.

For each cell measure its worst-case losses and use
`required total thermal resistance <= allowed temperature rise / heat watts`,
including the component-to-plate path, plate-to-sink interface and sink-to-air
path. For example, 30 W with a 20 °C budget needs **≤0.67 °C/W total**; that is
an illustration, not a selected sink. Include sun loading and the selected
maximum ambient temperature. A circulating fan plus plastic walls is not a
substitute for this calculation and a closed-lid temperature trial.

Add a purchased **hydrophobic pressure-equalisation membrane vent**, with its
specified seal and orientation, to each isolated cell if needed. Such a vent
addresses pressure cycling and humidity management, not bulk heat removal.
Its thread/cutout is left to the selected part. Condensation still needs
assessment; a weather-sealed enclosure is not automatically dry internally.

## Completion gates for the outdoor build

1. Confirm TPU hardness/grade, saltwater exposure and maximum ambient/sun exposure.
2. Validate the coupon and then the full-size TPU seals and printed walls.
3. Select shaft-seal/bearing/hub arrangement and verify motor/driver packing.
4. Select thermal plates/sinks for **each** cell and verify temperature under load.
5. Select glands/pressure vents and model their holes, fastening and seals.
6. Repeat water tests on the fully assembled, unpowered enclosure with its real
   cables, shafts and thermal hardware fitted. Then perform operating thermal
   tests. No IP designation is earned by CAD or informal spray tests alone.

## Regeneration

```sh
python3 -m venv /path/to/cad-env
/path/to/cad-env/bin/pip install -r mechanical/outdoor-enclosure/requirements.txt
/path/to/cad-env/bin/python mechanical/outdoor-enclosure/generate.py
```

Checks cover manifold connectivity, exported-STL watertightness/volume, print
bounds, lid/body and cap/body non-interference, gasket groove fit and nominal
gland displacement area. GLB green loops show compressed-height **location
surrogates**, not a material deformation simulation. Hardware, component
packing, frame, shaft cartridges, sinks and cable routing are not modelled.

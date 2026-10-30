# DepthPilot v4 — hinged electronics roof

Both motors/gearboxes remain on the base. Both drivers, the 24 V converter and
the two brake relays move with the roof on a suspended metal carrier. **The
entire upper hood opens**, not just the thin top panel. This avoids drawing a
loaded carrier through the narrow top rim. The horizontal opening joint is at
Z=180 mm; there is one shared interior when closed.

## Files

- [Closed model](winch-00-degrees.glb)
- [Half-open model](winch-45-degrees.glb)
- [90° open model with prop](winch-90-degrees.glb)
- [Side-elevation drawing](hinged-layout.svg)
- [STLs](stl/) — rigid shell parts, TPU seals and clearly named metal fit templates
- [Validation and print quantities](validation.json)
- [Editable Python generator](generate.py)
- [Complete ZIP](../hinged-roof-v4.zip)

The separate PSU/Jetson box remains the [v3 design](../two-box-v3/README.md),
238 × 238 × 386 mm including its lid. The connection remains one **2 m hybrid
48 V/RS485 cable, unplugging at both ends**. This package replaces the winch
box geometry; it does not contain the side box. All STL/source dimensions are
mm; GLB files use metres. Component blocks are reservations, not vendor CAD.

## Envelope

| Feature | Dimensions / coordinates (mm) |
| --- | --- |
| Closed shell, including roof | **238 W × 238 D × 396 H** |
| Base print | 230 × 230 × 180 |
| Upper hood print | 230 × 230 × 210 |
| Roof print | 238 × 238 × 16 |
| Rear hinge axis | Parallel to X, at Y=246, Z=180 |
| Modelled opening | 0–90° |
| Open hood maximum rear reach | Y=462 |
| Open hood maximum height | Z=430 |

The modelled brackets, clamps, clevises and blank caps enlarge the closed
envelope to **288 W × 260 D × 396 H**, X=-38…250, Y=-6…254. This excludes
drums, a real connector, fastener heads and an external heatsink. Leave at least
232 mm behind the body's rear wall for the opening motion, plus working room.
Drums and the side box are omitted from these views to expose the mechanism.
Final connector and sink projections will increase the envelope.

The height increased 50 mm from v3 to keep the suspended driver layout clear
of the closing joint and leave space above the drivers. Do not mix the v3
winch base/riser with these parts: hinge mounts, closing joint and cable entry
have changed.

## Mechanism and closure

Two identical rear hinge sets sit at X=26 and 160. Each fixed bracket has two
12 mm ears; the moving bracket has a 19 mm centre knuckle, with 0.5 mm axial
clearance on each side. Bores are Ø6.4 for nominal M6/6 mm pins. Use retained
metal pins with the smooth shank spanning the bearing surfaces. Each leaf has
four M4 fixing holes at 32 × 16 mm pitch. Blind insert pockets in reinforced
body pads open at the outside rear wall, with 8 mm nominal depth.

**Hinge STLs are fit templates for metal fabrication**, not load-qualified
printed hinges. The carrier, equipment and any heatsink create a substantial
moment. Size the final metal hardware, inserts and reinforced-shell attachment
for measured mass, opening loads and expected handling. Adjust CAD pockets to
the actual inserts before printing.

Four external screw-clamp pairs close the hood: on both sides at Y=70 and
185. They remain completely outside the TPU seal. The upper bracket has a
Ø4.5 clearance hole; the lower has a Ø3.3 pilot for an **M4 tapped hole in
metal**. Each bracket mounts to a reinforced side pad with two M4 bolts.
Use four removable M4 closing screws, approximately 14 mm long including the
8 mm upper flange and a washer, with the final length checked against the blind
tapped hole. Remove all four screws before opening. There are no printed
threads. These clamp STLs are also metal fit templates.

The lid is bolted permanently to the hood for normal operation using eight
M4 screws and a separate TPU loop. Opening the hood does not release those
eight screws. At the main closing joint, hard lands stop gasket compression.
Hinge positional tolerance and clamp preload must allow the lands to seat
uniformly; verify with contact impressions and a full-size seal trial.

## Open-position support

A **removable metal prop** on the left holds the hood at 90°. It is fitted
only after supporting and opening the hood, and removed while supporting the
hood before closing. Use two retained Ø5 mm pins in Ø5.5 holes. The prop has
**337.31 mm hole-centre spacing**, 12 mm width, 6 mm thickness and rounded
6 mm-radius ends (349.31 mm overall length). It fits the 8 mm-wide clevis slots.
The clevis mounts are included as metal fit templates with blind side-pad
fixings. The prop reference STL is in assembly position and **is not a printable
part**; cut it from suitable metal and check buckling/attachment capacity.

The model validates the fitted prop against the fully open assembly. It does
not claim an automatic stay, a self-locking hinge, or a mechanical overtravel
stop. Do not swing beyond the modelled 90° range. Store the removed prop
separately. Select practical locking pins and add their retention tethers.

## Equipment mounting

### Motors

The 210 × 206 × 3 mm metal base plate sits at Z=16–19 on four 2 mm spacers over
the floor bosses. Its M4 holes are at (20,20), (210,20), (20,210), (210,210).
Floor boss pockets are blind Ø5.8, 5 mm deep; match the chosen short insert.
The separate underside enclosure fixings remain at X/Y=30/200.

The motor/gearbox reservations are 57 × 195 × 57 mm with shaft centres at
X=55/175, Z=65. Use metal brackets to locate the actual assemblies above the
plate and maintain those axes. Vendor bracket patterns are **drill-to-fit**;
they are not fabricated from guessed dimensions. The plate and printed-floor
attachment are a prototype mounting concept, not a rated winch load path.
Provide a structural metal frame/attachment sized for the intended rope loads.

### Roof carrier

The carrier is a metal assembly, with these fabrication blanks:

| Piece | Size | Quantity |
| --- | --- | ---: |
| Roof carrier plate | 170 × 172.5 × 3 | 1 |
| Driver mounting panel | 140 × 130 × 3 | 1 |
| Converter/relay mounting panel | 175 × 112 × 3 | 1 |
| Driver suspension strips | 8 × 47 × 3 | 2 |
| DIN-panel suspension strips | 8 × 22 × 3 | 2 |

The panels and strips are shown touching in their assembled positions; fabricate
as a welded assembly or detail equivalent bolted angles after confirming access.
No weld strength or bolted-angle design is implied by the touching meshes.
The top plate fixes at X=40/190 and Y=45/190 through four holes to reinforced
roof bosses. Pocket depth is 5 mm from the boss underside; it does not penetrate
the roof. Account for insert length and bolt engagement rather than bottoming
the screws. Include a bonding strap where required.

Drivers are upright when closed at Z=210–328, with their rear mounting faces
against the vertical panel at Y=144.5. The nominal driver blocks remain
25.5 × 79.5 × 118 mm; verify actual drive models, mounting ears and connectors.
The 24 V converter and relays are above the rear motor region, against a
second panel at Y=204.5. Their stated envelopes omit the DIN rail itself:
allow for rail/clip projection by shifting hardware forward during detailed
mounting, and recheck all clearances. All component holes are drill-to-fit.

### Wiring and thermal path

The hybrid connector plate now sits on the **fixed right side of the base**,
centred at Y=120, Z=135. It has a 40 mm square opening and a 76 mm square blank
cap. Select the weather-rated bulkhead connector before cutting its hole.

Route the motor, encoder, brake and incoming power/data connections through
restrained flexible loops near the rear pivot. Keep the loops inside the
sealing perimeter, clear of the gasket and pinching surfaces. Reserve bend
radius for the actual harness, separating motor and signal wiring. Cable motion
is not represented by the rigid collision test; prototype a full opening cycle
with the real harness before fixing cable clips.

The carrier can support a conductive path to a gasketed external metal roof
plate/heatsink, but **the printed roof is not a heatsink**. Sink size, thermal
interface, final roof cutout and seal remain to be engineered from measured
losses. The winch base also needs a heat path for motor losses. Include the
selected sink mass in hinge/prop loads and its shape in the sweep check.

## Gaskets and printing

The main and roof seals use identical continuous TPU rings: 200.8 mm outside,
195.2 inside, 2.4 mm free height. The 4 mm-wide × 2 mm-deep grooves give 16.7%
nominal squeeze with lateral displacement room. Use the earlier
[seal-test kit](../outdoor-enclosure/README.md) to tune your filament first;
hardness, porosity and compression set are not yet established.

Print shell parts in supplied orientations. All printable extents are ≤238 mm,
so a 5 mm lid brim occupies 248 mm on the Centauri Carbon. ASA is preferred
for outdoor UV/temperature exposure, subject to the filament and print process;
PETG is suitable for fit trials. Start at 0.2 mm layers, 8 walls and 8 top/bottom
layers and inspect the slice for solid sealing lands. The base shaft openings,
internal rims and roof bosses need appropriate support/bridging treatment.
Keep support scars off gasket surfaces. TPU rings should be solid, dry and
printed using their manufacturer's slow-speed settings.

File names ending `fit-template` are dimension/assembly trials for metal parts.
They are intentionally distinguished from the shell, spacer and gasket prints.
Mesh watertightness is not physical waterproofness. Leak-test printed walls,
shaft cartridges, connector, roof interface and gaskets together before outdoor
use. No IP rating is claimed. Shaft caps are temporary blanks; rotary seals,
shaft extensions and drum clamp engagement still require detailed hardware.

## Verification and regeneration

```sh
python3 -m venv /path/to/cad-env
/path/to/cad-env/bin/pip install -r mechanical/hinged-roof-v4/requirements.txt
/path/to/cad-env/bin/python mechanical/hinged-roof-v4/generate.py
```

The generator checks positive-volume watertight connected print meshes, STL
reload, printer bounds, closed pairwise part interference, main gasket groove
fit, open prop fit and the rigid moving assembly against fixed geometry at
**every integer degree from 0 to 90**. This is sampled motion verification,
not a continuous collision proof or strength simulation. Closing screws are
assumed removed for opening. Pin heads, washers, wiring and mounting hardware
are not included. Small fitting tolerances need physical confirmation.

Before a full shell print, make hinge/clamp/clevis templates and gasket coupons,
verify selected component dimensions and make a physical carrier/hinge mock-up.
The supplied model is mechanically more detailed than v3 but remains a prototype
design requiring final metal hardware, mounting, cooling and water validation.

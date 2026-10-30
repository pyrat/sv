# DepthPilot iteration 3 — two containers, one cable

**Latest mounting decision:** motors/gearboxes on the base; both drivers, the
24 V converter and brake relays on a carrier beneath a hinged winch-box roof.
See [hinged-roof requirements](HINGED-ROOF.md). The current CAD files still use
the earlier removable lid; the hinge, carrier and opening sweep are not yet
implemented in this v3 folder. Use the new
[v4 hinged-hood model](../hinged-roof-v4/README.md) for the revised winch box.

This revision puts both geared motors, both drivers, the 48-to-24 V converter
and two brake relays in **one shared winch enclosure**. A separate side enclosure
holds the mains/48 V PSU, Jetson, 12 V converter and RS485 adapter. A single
**2 metre hybrid cable, detachable at both containers**, carries power and Modbus
between them. Each enclosure will have a sealed bulkhead connector and the cable
will have two mating plugs. Connector selection and plate cutouts remain pending.
Drums remain outside.

## Deliverables

- [Layout drawing](layout.svg)
- [3D assembly and component envelopes](assembly.glb)
- [Printable STLs](stl/)
- [Electrical architecture](ELECTRICAL.md)
- [Component coordinates](component-layout.json)
- [Geometry validation](validation.json)
- [Editable generator](generate.py)
- [Complete ZIP](../two-box-v3.zip)

This is a dimensioned development design. Shaft cartridges, equipment mounts,
thermal hardware and the selected hybrid connector still need detailing. Blank
interface plates permit enclosure trials; they do not represent working shaft
or cooling assemblies. Outdoor water and thermal performance are untested.

## Dimensions and construction

| | Winch container | Power/compute container |
| --- | --- | --- |
| Body footprint | 230 × 230 mm | 230 × 230 mm |
| Overall including lid | 238 × 238 × 346 mm | 238 × 238 × 386 mm |
| Lower body height | 180 mm | 180 mm |
| Upper extension height | 160 mm | 200 mm |

Dimensions exclude drums, connectors, cooling sinks and mounting frame. Each
base and extension prints separately on the Centauri Carbon. The largest XY
part is the 238 mm lid; a 5 mm brim occupies 248 mm. No partition closes the
opening between a base and its extension: each assembled enclosure has one
shared interior. The drawing shows boxes beside one another; their separation
is illustrative and can change with cable length.

The body joint is at Z=180. Eight internal bolts clamp the extension onto a
continuous TPU gasket in the base. Internal bolt positions are 28 mm from the
outer edges; the upper rim leaves straight screwdriver access before equipment
is installed. Both joint rims have a **158 × 158 mm central opening**, so install
large lower components before fitting the extension. Lid opening is 186 mm square.

The base and extension are fixed together for normal service; remove the lid
for routine access. No rain skirt covers the body joint: that joint relies on
its compressed gasket. The lid has a 10 mm rain skirt with 0.5 mm side clearance.

## Component layout

Coordinates use X across, Y front-to-back and Z up. Each box has its own origin.

### Winch box

- Motor/gearbox axes: **X=55 and 175, Z=65 mm**, pointing through the front.
  Reservations are 57 × 195 × 57 mm, using the earlier Stepperonline motor and
  MGL gearbox assumption. Both actual assemblies must be checked, especially
  if one is the Leadshine alternative.
- Ø110 demonstration drums have **120 mm centre spacing and 10 mm flange gap**.
  The drawn rear drum face is Y=-30 solely to reserve shaft-cartridge space.
  This requires an engineered shaft/hub arrangement; the existing short shaft
  cannot be assumed to reach through that stack.
- Drivers sit above the motors at Z=145–263, X=56–81.5 and X=148–173.5.
  The **25.5 × 79.5 × 118 mm** blocks are a provisional slim-drive layout.
  Check the selected CL57RS/CS2RS-D507 drawing and mounting orientation before
  using these as hardware templates. Modelled side clearance is at least 50 mm,
  spacing between drivers is 66.5 mm, and space below them to the motor tops is
  51.5 mm. Wiring, mounts and connector bodies are not represented.
- DDR-15L-24 reservation: 17.5 × 54.5 × 90 mm.
- Two relay reservation: 12.4 × 80 × 94 mm combined; rail clips, end stops and
  wire bends need extra physical verification.
- Right-hand hybrid interface centre: local Y=115, Z=250.

Use metal brackets/base to carry motor torque and drum loads. Printed skins,
floor bosses and shaft blank caps are not the structural winch load path.
The upper drivers need a removable metal mounting structure with a heat path
to an external sink. The included 150 mm square mounting blanks are drill-to-fit
templates, not designed structural brackets or driver heatsinks.

### Power/compute box

- Upright NDR-240-48 at (20,95,35), size 63 × 113.5 × 125.2 mm.
  It has 27 mm to the inner floor below and over 40 mm free above. Keep its
  ventilation path clear when adding DIN mounts. The reinforced joint rim and
  wiring must not obstruct its top vents.
- Jetson/connector reservation at (12,30,265), size 145 × 130 × 85 mm.
- DDR-60L-12 at (169,150,261), size 52.5 × 54.5 × 90 mm; verify actual drawing.
- USB–RS485 adapter reservation at (25,180,265), size 45 × 35 × 25 mm.
- Left-hand hybrid interface centre: local Y=115, Z=250.

The nearest upper heat-source reservation begins 100.8 mm above the PSU body.
That is a spatial allowance, not a thermal qualification. Mount PSU and
converters upright on supported DIN rail; metal rail/brackets are not modelled.
The mains inlet, fuse, earth terminals and covered wiring require space and
selected hardware before cutting their openings. They stay in this box.

## Static seals and interfaces

Body and lid seals share the same continuous TPU loop: **200.8 mm outside,
195.2 mm inside, 2.4 mm high**. Groove width is 4 mm, depth 2 mm; nominal
compression is 16.7%. Rigid mating lands stop compression. Screw pockets are
blind; body-joint screws are inside the seal and lid screws outside it.

| Interface | Opening | Cap | Screw-square pitch | Qty |
| --- | --- | --- | --- | ---: |
| Shaft | 70 × 70 | 104 × 104 × 6 | 92 | 2 |
| Thermal | 100 × 100 | 136 × 136 × 6 | 124 | 3 |
| Hybrid cable | 40 × 40 | 76 × 76 × 6 | 64 | 2 |

Each interface has eight screws at the corners and midpoints of the given
square, and a continuous TPU gasket. Interface gasket height is also 2.4 mm.
Caps are supplied without functional cutouts. Select the actual sealed connector,
shaft cartridge and thermal plate before replacing or machining the caps.

Use purchased rotary seals on a smooth compatible shaft/sleeve. Static TPU
gaskets must not rub on a rotating keyed shaft. Final shaft projection, bearing
support, drum clamp engagement and seal wear surface need to be designed together.

There are no open fan vents. Each box needs an engineered conductive heat path
to external sinks; the winch box must reject motor and driver losses. The power
box must reject PSU, Jetson and converter losses. Internal fans only circulate
air. The thermal caps identify potential interfaces, not sized heat exchangers.
Select sealed membrane pressure vents if required and model their holes after
selection. No IP rating is claimed.

## Printing and assembly

Print counts are in `validation.json`: one each of the four body sections, two
lids, four main TPU rings, two shaft caps/rings, three thermal caps/rings, two
hybrid caps/rings and two optional mounting blanks.

For rigid parts, ASA is preferred for outdoor UV/temperature exposure; PETG
is suitable for fit trials subject to its actual grade. Start at 0.2 mm layers,
8 walls and 8 top/bottom layers; inspect the sliced walls and seal lands for
solid fill. Print bodies/extensions upright, lids upside down as supplied,
and caps flat. Side opening roofs and inward top rims require support or
validated bridging/overhang settings. Remove support without damaging seal lands.
All STL files are supplied in print orientation, in mm.

Print TPU rings solid, dry and slowly according to the filament supplier.
First use the [previous revision's seal-test kit](../outdoor-enclosure/stl/):
it has the same groove cross-section and 2.2/2.4/2.6 mm height samples. TPU
hardness/grade is not yet known; test compression, creep and leakage before
printing every gasket. Printed walls also need leak testing and possibly a
compatible sealing coating. A watertight mesh is not proof of impermeable walls.

1. Print joint samples and fit inserts. All main insert bores are nominal
   Ø5.8 × 6 mm; change `INSERT_D` for the chosen insert. Underside pockets are
   only 5 mm deep and retain 3 mm of floor above them.
2. Fit the lower motor or PSU assemblies and their metal mounts before adding
   the upper extension. Internal floor bosses have Ø3.4 mm blind pilot holes;
   do not drill through the shell.
3. Seat the body gasket, add extension and tighten its eight internal screws
   evenly to seat the rigid lands. The 12 mm extension flange plus washer and
   about 5 mm engagement suggests M4 × 18; measure the actual stack to prevent
   blind-hole bottoming. Install these screws before the upper electronics.
4. Install upper mounting structures, equipment and wiring. Mount the two brake
   relays and 24 V converter in the winch box and the 12 V converter in the side
   box. Provide restraint for both ends of the hybrid cable.
5. Fit interface gaskets and selected hardware. M4 × 12 is a starting length
   for 6 mm caps/lids plus washers, subject to actual inserts. Lid removal should
   expose all upper equipment without removing the body joint.
6. Fit continuous lid seals and lids. Check full-size compression uniformity,
   especially between screws, then water-test unpowered with actual interfaces
   and cable fitted. Check temperature under worst-case load with lids closed.

For the two boxes allow 88 main M4 inserts/bolts: 16 body-joint, 16 lid and 56
interface fixings. Add eight shorter underside inserts/fixings, mounting
hardware and spares. Choose corrosion resistance after confirming saltwater
exposure. Nominal main bore dimensions are not a universal insert specification.

## Regenerate and validation

```sh
python3 -m venv /path/to/cad-env
/path/to/cad-env/bin/pip install -r mechanical/two-box-v3/requirements.txt
/path/to/cad-env/bin/python mechanical/two-box-v3/generate.py
```

The generator exports STLs, GLB, drawing, component/validation JSON and ZIP.
Checks cover print extents, single connected positive-volume meshes, STL reload,
shell/cap non-interference, component block non-interference, main gasket fit,
joint-screw tool paths and selected layout clearances. GLB green seals show
compressed-height location surrogates, not material simulation. The cable is an
illustrative route. Hardware fastening, interface gasket deformation, actual
component fit, wire bends, loads, thermal behaviour and water sealing require
physical/detail-design verification.

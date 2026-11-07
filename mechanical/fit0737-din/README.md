# FIT0737 DIN-rail tunnel cradle

Printable cradle for the DFRobot FIT0737 USB-to-RS485 adapter, with a captive,
positively locking sliding DIN latch. Prototype: geometry checked, physical fit
and retention forces untested.

## Geometry and retention

- Standard 35 × 7.5 mm top-hat rail, nominal 1 mm steel. Rail runs along Z;
  adapter length runs along Y, perpendicular to the rail.
- Uses the approximate **83 × 26 × 17 mm overall envelope** from the
  [datasheet](https://mm.digikey.com/Volume0/opasdata/d220001/medias/docus/9029/FIT0737.pdf).
  This is not a measured plastic-body drawing. The tunnel has 0.7 mm clearance
  per side: 27.4 mm wide × 18.4 mm deep. Both connector ends remain open.
- A short central tunnel prevents the adapter lifting out sideways. Two cable
  ties through the backplate slot pairs secure it axially. These provide friction
  retention, not dimensioned connector-shoulder stops; check pull resistance on
  the actual adapter before use. Keep ties off terminal screws and LEDs.
- The latch reuses `../jetson-orin-din/generate.py`: fixed hook at one flange,
  moving hook at the other. A guide with both a roof and a bottom keeper captures
  the slider in both directions through the print plane; a pin through its
  closed slot limits travel to 6 mm, and a flexible tooth locks it closed.
  It cannot be removed by simply pulling the tab along its travel axis.
- **Fit DIN end stops on both sides** to block movement along the rail. The
  flange hooks prevent pull-off; they do not positively locate the mount along
  the uniform rail profile.

## Files and printing

- `fit0737-din-print-in-place.stl`: millimetres, one object with two shells.
- `assembly.glb`: coloured reference with rail and approximate adapter envelope.
- `generate.py`: parametric source and geometry checks.
- `validation.json`: generated results.

Print in the supplied orientation, Z upward. PETG, 0.4 mm nozzle, 0.2 mm layers,
5 walls and approximately 40% infill. Preserve both shell positions. The latch
has 0.4 mm moving clearances. The added bottom keeper means this is **not a
support-free print**: the slider and main chassis start 2.4 mm above the lowest
face. Use removable supports under their exposed undersides, with a tuned
support interface beneath the slider. Remove this material through the open
guide ends and confirm free travel. Do not fill the side gaps or roof gap with
supports; do not let a brim fuse the shells. The tunnel's
upper sidewall bridges 18.4 mm: use tuned bridging, or removable supports inside
the adapter tunnel only. The rail guide roof has a 6.8 mm bridge. Clear strings
and free the latch before installing the adapter. Check the latch locks fully.

Slide the adapter into the tunnel and tighten two cable ties through the slot
pairs, around the housing and backplate. Foam shims can accommodate a smaller
actual housing. Verify both connectors remain fully accessible.

To attach: press the locking tongue toward -X and pull the slider toward -Y
6 mm. Engage the fixed hook, pivot onto the rail, then push the slider back
toward +Y until its tooth locks. To release, depress the tongue and pull again.
Leave room at the -Y end for release access and the USB cable.

## Regenerate

```sh
python3 -m venv .venv
.venv/bin/pip install -r mechanical/fit0737-din/requirements.txt
.venv/bin/python mechanical/fit0737-din/generate.py
```

Checks cover connected watertight shells, nominal envelope/rail clearance,
61 rigid-slider travel positions, lock engagement, end stops and lateral/roof
capture. They do not simulate tongue deformation, print tolerances or loads.

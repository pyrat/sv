# Jetson Orin Nano edge-on DIN bracket

**Prototype — geometry checked, not yet physically printed or load-tested.**

- NVIDIA P3768 developer-kit carrier, plastic base removed.
- Board vertical and **perpendicular to the rail length**. Looking at the fan,
  **GPIO is at the top and USB is on the right** (confirmed orientation).
- DIN mount on the **bottom 79 mm short edge, opposite GPIO**, centred on the
  board width. The release slider pulls toward the right.
- Standard 35 × 7.5 mm top-hat rail, nominal 1 mm steel thickness.
- One print, containing a chassis and captive sliding latch. No rail screws,
  nuts, inserts, separate spring, or latch assembly.
- Four M2.5 × 8 mm screws attach the board directly to printed pilot holes.

## Files

- `jetson-orin-din-print-in-place.stl`: print this, in millimetres.
- `assembly.glb`: coloured 3D reference; includes a nominal rail and PCB envelope.
  The green rectangle represents the PCB, not a complete component model.
- `generate.py`: editable parametric geometry and interference checks.
- `validation.json`: generated verification results.

## Print

Use the STL's supplied orientation: the flat ladder-frame back and slider sit on
the bed; the four standoffs point upward. PETG, 0.4 mm nozzle, 0.2 mm layers,
5 walls, 5 top/bottom layers, approximately 40% infill. Keep the latch region
solid through wall settings. No supports in the slider clearances. The guide
roof has a short 6.8 mm bridge and 0.4 mm vertical clearance above the slider.
Avoid a brim joining the two shells; compensate elephant foot if needed.

Import as **one object with two shells**, preserving their positions. Do not
auto-arrange the shells or merge away the clearances. Both shells start on the
bed; the pin, roof and end stops make the moving piece captive.

The nominal moving gaps are 0.4 mm per side. These and the 2.15 mm screw pilots
need checking on your printer. Clear any strings and gently free the slider
before attaching the electronics. The slender locking tongue is intended to
flex in the layer plane. Its elastic behaviour and life need a physical trial.

## Attach and release

1. Press the small projecting tongue toward print-coordinate -Y (away from the
   PCB). This retracts its locking tooth from the side pocket.
2. Hold it depressed and pull the slider right approximately 6 mm (toward +X).
3. Engage the fixed hook with one rail flange, then pivot the
   bracket onto the rail.
4. Push the slider left to its stop. The tongue should spring into its pocket,
   positively blocking outward travel. Check that it has engaged.
5. Release by pressing the tongue and pulling right again.

Fasten the board with four M2.5 × 8 mm screws. The 8 mm-deep blind pilots allow
about 6.4 mm engagement through the nominal 1.57 mm PCB. Form the threads gently;
do not overtighten. The PCB underside is 14 mm above the print bed, giving
10 mm clearance over the 4 mm frame ribs. Standoff contact diameter is 6 mm.
Check nearby components and any fitted SSD against the frame before tightening.

The bracket occupies approximately 82.5 × 112 × 24 mm. Its 112 mm axis follows
the board's vertical direction and its 24 mm axis follows the rail. The cooler
extends beyond the bracket. The rail sits along the bottom short edge, opposite
GPIO. Allow 6 mm of extra space to the right for slider travel and room to
operate the tab. Leave fan airflow unobstructed.

## Dimensions and provenance

NVIDIA's [P3768 A04 reference design archive](https://developer.nvidia.com/downloads/assets/embedded/secure/jetson/orin_nano/docs/jetson_orin_nano_devkit_carrier_board_reference_design_files_a04_20230320.zip/)
contains `P3768_A04_Allegro_layout_ascii_extract.zip/P3768_A04.brd.alg`.
The `MEC5` mounting-hole records specify 2.75 mm holes at native coordinates
`(0,0), (92,0), (0,58), (92,58)`; two alternative holes are at `(86,0)` and
`(86,58)`. This design uses the **92 × 58 mm** set, not the 86 mm alternatives.
The board outline is X=-4…96, Y=-17…62 in that file.

To put GPIO at the top, native coordinates `(a,b)` become print coordinates
`(62-b, a+4)`. The four standoffs are therefore at `(4,4), (62,4), (4,96),
(62,96)`, within a 79 × 100 mm rotated board envelope.

NVIDIA's [carrier-board specification, v1.3, Figure 4-1, page 29](https://developer.nvidia.com/downloads/assets/embedded/secure/jetson/orin_nano/docs/jetson_orin_nano_devkit_carrier_board_specification_sp.pdf)
confirms a 100 × 79 mm PCB, 1.57 ± 0.16 mm thickness and 4.30 mm maximum
underside component height for the illustrated board.

## Regenerate and verify

```sh
python3 -m venv .venv
.venv/bin/pip install -r mechanical/jetson-orin-din/requirements.txt
.venv/bin/python mechanical/jetson-orin-din/generate.py
```

The generator checks watertight shells, zero nominal rail/PCB interference,
61 positions over the rigid slider's 6 mm travel, positive tooth engagement,
and captive end stops. Travel checks omit the flexible tongue: they do not
simulate its deformation. These are geometry checks, not proof of printability,
rail tolerances, retention force, vibration performance or PETG heat resistance.
Trial-fit the print to the actual rail and board before putting it into service.

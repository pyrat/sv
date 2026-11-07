# Jetson Orin Nano edge-on DIN bracket

**Revised prototype — geometry checked; tunnel revision not yet physically tested.**

The original print-in-place latch seized, and the subsequent open-slot slider
could escape sideways. This revision uses a solid slider in a four-sided tunnel.
An M3 screw crosses the tunnel entrance behind the slider's tail, blocking full
withdrawal without passing through the slider. Remove the screw to remove the
slider. The locking tongue holds the closed position; the screw is a withdrawal
stop, not a clamp.

- NVIDIA P3768 developer-kit carrier, plastic base removed.
- Board vertical and **perpendicular to the rail length**. Looking at the fan,
  **GPIO is at the top and USB is on the right** (confirmed orientation).
- DIN mount on the **bottom 79 mm short edge, opposite GPIO**, centred on the
  board width. The release slider pulls toward the right.
- Standard 35 × 7.5 mm top-hat rail, nominal 1 mm steel thickness.
- Two printed parts: chassis and slide-in latch, retained by one M3 × 8 mm
  screw. No nuts, inserts or separate spring.
- Four M2.5 × 8 mm screws attach the board directly to printed pilot holes.

## Files

- `jetson-orin-din-body.stl`: chassis, in millimetres.
- `jetson-orin-din-slider.stl`: separate sliding latch, in millimetres.
- `jetson-orin-din-clip-test.stl`: clip-only test body, in millimetres;
  print with the same separate slider and retain it with an M3 × 8 mm screw.
- `assembly.glb`: coloured 3D reference; includes a nominal rail and PCB envelope.
  The green rectangle represents the PCB, not a complete component model.
- `generate.py`: editable parametric geometry and interference checks.
- `validation.json`: generated verification results.

## Print

### Quick rail-fit test

Print `jetson-orin-din-clip-test.stl` and `jetson-orin-din-slider.stl` as separate
objects, in their supplied orientations, using the settings below. The test
body is the full bracket's complete tunnel-and-hook mechanism, without the PCB
frame, and measures 73 × 17 × 24 mm. Follow the latch assembly instructions below.

On your actual rail, check that both hooks seat, the tongue locks, the slider
releases smoothly and the closed clip does not rock excessively or pull off.
With the screw installed, check that the slider stays captured sideways and
vertically throughout its travel, and cannot be pulled completely out. When
locked, it should have only small clearance movement, not the full release stroke.
This checks local rail fit, not the full bracket's stiffness under board load.

### Print settings

Use each STL's supplied orientation: the flat ladder-frame back and slider sit on
the bed; the four standoffs point upward. PETG, 0.4 mm nozzle, 0.2 mm layers,
5 walls, 5 top/bottom layers, approximately 40% infill. Keep the latch region
solid through wall settings. The empty guide roof bridges 11.8 mm and provides
0.5 mm nominal vertical clearance above and below the assembled slider. Print this bridge
without supports if your printer bridges cleanly; remove any sag or strings
before assembly. Compensate elephant foot if needed.

Import as **two separate objects** and space them apart on the bed, or print
them in separate jobs. Both files have their minimum Z at zero.

Both tunnel side walls have 0.4 mm nominal clearance to the rigid slider.
These and the screw pilots need checking on your printer. The slender locking tongue is intended to
flex in the layer plane. Its elastic behaviour and life need a physical trial.

## Assemble the latch

1. Leave the M3 retaining screw out. With the chassis in print orientation,
   bring the slider in from the right (+X), flat face down and hook facing
   away from the PCB (-Y). Feed it above the tunnel floor and push it left
   into the guide. The hook fits through the tunnel.
2. Depress the small locking tongue away from the PCB as it passes the guide
   wall; use a small flat screwdriver through the narrow top access window
   when needed. Let it spring into the pocket at the closed position.
3. Drive one **M3 × 8 mm screw (length under the head)** into the 2.5 mm pilot
   on top of the guide roof, near the entrance. Its shaft sits behind the
   solid slider tail and stops withdrawal at 6.5 mm. A rear shoulder stops
   inward travel. Seat the head gently against the roof;
   the screw must not clamp the slider. Use a pan/button head, not countersunk.
4. Check that pressing the tongue permits approximately 6 mm of smooth travel
   and that the screw prevents withdrawal. Remove the screw to replace the latch.

The chassis (or clip-test body) and slider both need reprinting; the tunnel
revision is not compatible with either earlier design. Assemble before fitting the rail
or electronics.

## Attach and release

1. Use a small flat screwdriver through the top access window to press the
   tongue toward print-coordinate -Y (away from the PCB). This retracts its
   locking tooth from the side pocket.
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

The bracket occupies approximately 90.5 × 115 × 24 mm. Its 115 mm axis follows
the board's vertical direction and its 24 mm axis follows the rail. The cooler
extends beyond the bracket. The rail sits along the bottom short edge, opposite
GPIO. Allow 6.5 mm of extra space to the right for slider travel and room to
operate the tab. Initial latch insertion needs approximately 38 mm of clearance
to the right. Leave fan airflow unobstructed.

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

The generator checks each exported STL is a single watertight shell, zero
nominal rail/PCB interference, 381 positions along the screw-free insertion
path, 66 positions over the rigid slider's full 6.5 mm travel, positive tooth engagement,
and the external screw and rear shoulder stops. It also checks sideways and
vertical blocking at every travel sample, plus screw-stop engagement at the
extremes of lateral and vertical clearance. These are rigid translation checks,
not a simulation of twisting or forced deformation.
Insertion and travel checks omit the flexible tongue: they do not
simulate its deformation. These are geometry checks, not proof of printability,
rail tolerances, retention force, vibration performance or PETG heat resistance.
Trial-fit the print to the actual rail and board before putting it into service.

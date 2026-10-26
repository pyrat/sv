# One-piece PETG demonstration drums

**Print two copies of `drum-one-piece.stl`, one per gearbox.** No separate flange
is needed. Units are millimetres; print at 100% scale, broad lower flange on the
bed and open centre upward. This model supersedes the two-piece design for the
requested unloaded demonstration.

## Shape and shaft fit

- 80 mm barrel diameter; 110 mm flange diameter; 50 mm overall height.
- Upper flange has a 45° sloping underside, growing from radius 40 mm at Z=31
  to radius 55 mm at Z=46. Both flanges are integral to the print.
- The straight barrel runs from Z=7 to Z=31: **24 mm of unobstructed cylindrical
  winding width**, versus 39 mm on the two-piece version. Rope space widens
  toward the flange rim; this is intended for a few loose demonstration wraps.
- A 64 mm central opening stays accessible for installing the clamp nuts.
- The same 14.3 mm nominal printed bore and 30 mm long split clamp fit the
  MG23-G20-D8's nominal 14 × 34 mm output shaft. Remove the loose shaft key.
- A **10.5 mm wide vertical access slot** on the bolt-head side admits the bolts,
  9 mm washers and straight Allen tool. Its rounded ends are centred at Z=13
  and Z=24, X=18, entering from -Y. The entrance flares slightly at the barrel
  surface. Move rope aside to access the bolts.

Print `shaft-fit-test.stl` first and check the actual shaft fit, nut pockets and
clamp grip. Adjust `BORE` in `generate.py` if necessary, then regenerate with
`generate_one_piece.py`. This checks the interface before committing to two
full drums; the complete drum still needs a physical grip check.

## Hardware for BOTH drums

| Hardware | Total |
| --- | ---: |
| M4 × 30 mm socket-head bolts (3 mm hex socket) | 4 |
| Standard M4 hex nuts, 7 mm AF × about 3.2 mm thick | 4 |
| M4 flat washers, about 9 mm OD × 1 mm thick | 4 |

Use standard nuts, not thicker nyloc nuts. A straight **3 mm Allen driver** or
the long leg of a 3 mm Allen key reaches the heads through the slot. Allow at
least 25 mm of slim tool reach. Needle-nose pliers or tweezers help seat the nuts.

## Printing in PETG

- Use the STL orientation: shaft axis vertical, broad lower flange on the bed.
- Starting settings: 0.4 mm nozzle, 0.2 mm layers, 6 walls, 6 top/bottom layers,
  30–40% gyroid infill; use your filament's temperature profile.
- Use a solid-infill modifier within 27 mm radius through the first 30 mm of
  height, covering the hub and clamp lugs.
- Designed for printing without supports: the upper flange has a 45° underside
  and the centre is open. Inspect the access-slot roof, transverse bolt holes
  and nut-pocket bridges in the slicer; their print quality depends on cooling
  and bridge settings. Clear sagging before fitting hardware.
- Requires a 110 × 110 mm footprint, plus any brim and bed margins.
- Deburr and smooth the access-slot lips and flange edges before adding rope.
  The access slot interrupts the winding surface; lay loose wraps gently over
  it, or cover it with smooth tape after tightening if rope catches.

## Assembly

1. Through the open centre, seat two M4 nuts in the clamp's +Y pockets (opposite
   the external slot). Hold each in place with pliers or a small piece of tape.
2. Put a washer on each M4 × 30 bolt. Feed the bolt **shaft first through the side
   slot**, then through the aligned clamp hole into its nut. Start both loosely.
   The slot passes the bolt heads and washers, avoiding awkward internal insertion.
3. Remove the gearbox shaft key and slide the drum onto the shaft. Leave about
   **3 mm clearance from the frontmost stationary gearbox face/boss**. For a
   34 mm shaft projection measured from that face, this gives 31 mm insertion,
   covering the full 30 mm clamp. Check your actual shaft shoulder and boss.
4. Tighten alternately through the side slot with the 3 mm Allen tool. Snug only
   until the drum grips; leave the clamp split visibly open.
5. Rotate by hand and check clearance around the gearbox. Add loose rope wraps
   for the demonstration. Recheck grip after initial use as PETG can settle.

For the requested **no-load demonstration**, not tensioned rope, lifting or
braking. Physical fit and grip have not been tested on your hardware.

## Regeneration and checks

Use the environment and `requirements.txt` described in `README.md`, then run:

```sh
/path/to/cad-venv/bin/python mechanical/prototype-drums/generate_one_piece.py
```

The generator checks single-solid connectivity, watertightness, winding and
positive volume, then reloads the STL. It also verifies full 9 mm diameter
bolt-head/washer insertion paths to both clamp faces, the M4 shaft paths and
the open centre above the clamp. These are geometric checks, not a print trial.

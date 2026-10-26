# Identical PETG prototype drums — MG23-G20-D8

## Recommended: one-piece version

Print **two `drum-one-piece.stl`** for the simplified, integral-flange drums.
See **[ONE-PIECE.md](ONE-PIECE.md)** for hardware, printing and assembly.
Generate with `python mechanical/prototype-drums/generate_one_piece.py` using
the environment below. Each drum only needs two M4 × 30 bolts, nuts and washers.

## Original two-piece version

Print **two `drum-body.stl` and two `outer-flange.stl`** for two identical drums.
All models use **millimetres** and are already oriented for printing. This is a
direct printed clamp design for unloaded demonstrations with loose rope.

## Dimensions and mounting

| Feature | Dimension |
| --- | --- |
| Rope barrel | 80 mm diameter |
| Flanges | 110 mm diameter |
| Clear winding width | 40 mm (39 mm beyond the lower root chamfer) |
| Assembled axial width | 50 mm |
| Barrel wall | 8 mm |
| Lower / removable flange thickness | 6 / 4 mm |
| Shaft bore | 14.3 mm, for nominal 14 mm shaft |
| Clamp engagement length | 30 mm |
| Flange fastening | Four M4 bolts on 70 mm PCD |

The BOM specifies the **MG23-G20-D8** output shaft as 14 mm diameter × 34 mm
projection. The drum clamps onto the bare shaft: **remove the loose parallel
key** first. No metal hub or key is used. Two transverse M4 bolts close a 1.6 mm
radial split in the printed hub. Bolt heads and hex nuts are accessible inside
the hollow barrel for installation before fitting the outer flange. **Tighten
the bolts from outside through the two barrel access ports**, using a straight
3 mm Allen driver or the long leg of a 3 mm Allen key. The ports have 6 mm throats
and flared entrances, aligned with the bolt heads at 13 and 24 mm above the rear
face (X=18 mm, entering from -Y). Use at least 25 mm of slim tool reach; a bulky
bit holder will not fit through the throat. Access remains possible with the
outer flange installed, provided rope is moved clear of the ports.

The outer flange has a 63.4 mm locating spigot fitting the 64 mm barrel opening.
Its flat outside face is solid. The 40 mm winding space suits a few demonstration
wraps of the project's 3–4 mm ropes; it is not sized to store the full travel rope.
Use tape to hold a loose rope tail to the barrel while winding.

## Hardware for two drums

| Hardware | Quantity | Purpose |
| --- | ---: | --- |
| M4 × 30 socket-head bolts | 4 | Shaft clamps, two per drum |
| M4 × 50 socket-head bolts | 8 | Removable flanges, four per drum |
| Standard M4 hex nuts, nominal 7 mm AF × 3.2 mm thick | 12 | All fastenings |
| M4 flat washers, about 1 mm thick and 9 mm OD | 12 | Under bolt heads |

Nut pockets are 7.3 mm across flats. Flange nuts sit in 4 mm deep recesses on
the gearbox-facing side. With the stated washers and nuts, M4 × 50 bolts engage
the nuts without projecting below the rear face. Check your actual hardware
stack; bolt ends and nuts must stay recessed. Do not substitute thicker nyloc
nuts in these pockets.

## Print the fit test first

`shaft-fit-test.stl` is a 52 mm diameter, 30 mm tall sample of the same split hub,
bolt holes and nut pockets. Print one with your intended PETG settings.

1. Remove the shaft key and check the actual shaft diameter and projection.
2. Confirm the sample slides onto the shaft by hand without forcing it.
3. Fit two M4 × 30 bolts, washers and nuts. Snug them alternately; check that the
   sample grips and releases normally. The split should remain visibly open.
4. Check that nuts seat and that your 3 mm Allen tool fits both bolt heads.

If the bore is too tight or loose, adjust `BORE` in `generate.py` and regenerate.
Change diameter in roughly 0.1 mm steps; keep the model at 100% scale so the
fastener holes and drum geometry remain correct. The fit sample checks the
interface; confirm grip again on the full drum, whose web has different stiffness.

## PETG printing

- Body: broad lower flange on the bed, open barrel upward.
- Outer flange: broad flat face on the bed, locating spigot upward.
- Fit sample: broad circular base on the bed.
- Suggested starting point: 0.4 mm nozzle, 0.2 mm layers, **6 walls**, **6 top and
  bottom layers**, 30–40% gyroid infill. Use a solid-infill modifier over the
  central clamp and lugs (within 27 mm radius, through the first 30 mm height).
- Use your filament's PETG temperature profile. Each part fits a 120 × 120 mm
  printable area without a brim; allow more space if using one.
- Main geometry needs no supports. The 4.5 mm transverse clamp holes, flared
  access ports and small nut-pocket roofs use short bridges; inspect these in
  the slicer. Clear minor
  sagging from the holes before assembly, keeping the clamp split free.
- Deburr flange edges and the seam where the removable flange meets the barrel.
  The outer flange edges have 0.8 mm chamfers.
- Smooth and round the access-port lips after printing, removing bridge sag or
  sharp seams before laying rope over them.

## Assembly, once per drum

1. Seat four M4 nuts in the recessed pockets underneath the body. Small pieces
   of tape can retain them temporarily; keep everything below the rear face.
2. Seat the two clamp nuts in the lug pockets; insert the M4 × 30 bolts and
   washers loosely from the opposite side.
3. Slide the body onto the keyless shaft. Leave **3 mm clearance** between the
   gearbox's frontmost stationary face/boss and the drum's rear face. With a
   34 mm projection measured from that face, the shaft reaches 31 mm into the
   drum, covering the full 30 mm clamp. Check the actual boss geometry and shaft
   shoulder on your gearbox; the downloaded gearbox CAD was unavailable.
4. Insert a straight 3 mm Allen driver (or the long leg of an Allen key) through
   each barrel access port and engage the corresponding bolt head. Snug the
   clamp bolts alternately from outside; no tool turning inside the drum is needed.
   Stop when the drum grips; do not force the split shut. Turn the drum by hand
   and confirm rear nuts, bolt ends and flange clear the gearbox throughout.
5. Flip the printed outer flange so its locating spigot points into the barrel.
   Align its four holes and fasten with four M4 × 50 bolts and washers from the
   outer face. Snug evenly, rather than crushing PETG under the washers.
6. Repeat for the second drum. Recheck clamp grip after the first demonstration,
   since PETG can settle under bolt preload.

These drums are for the requested **no-load demonstration**, not tensioned rope,
lifting or braking. Physical shaft fit and clamp grip require the test print;
the mesh checks below do not establish those properties.

## Source and verification

```sh
python3 -m venv /path/to/cad-venv
/path/to/cad-venv/bin/pip install -r mechanical/prototype-drums/requirements.txt
/path/to/cad-venv/bin/python mechanical/prototype-drums/generate.py
```

The editable source uses Manifold boolean solids and exports binary STL. The
generator checks each part for a single connected solid, watertight mesh,
consistent winding and positive volume, reloads the exported STLs, and checks
that the assembled flange does not intersect the body. It also checks a 4 mm
diameter straight tool path through each access port up to the bolt-head face.
Assembly coordinates:
body unchanged; rotate flange 180° about X, then translate +50 mm along Z.

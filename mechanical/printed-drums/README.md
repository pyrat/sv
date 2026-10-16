# Direct-drive rope drums — prototype v1

All STL dimensions are **millimetres**. Generate with `python3 mechanical/printed-drums/generate.py` (standard library only). These are prototype bodies, not load-rated assemblies.

## Files and assembly

| Part | Material | Assembly position along Z |
|---|---|---|
| `y-body.stl` | PETG | 0 |
| `y-removable-flange.stl` | PETG | +50 mm |
| `x-body.stl` | PETG | 0 |
| `x-removable-flange.stl` | PETG | +32 mm |
| `x-tpu-sleeve.stl` | TPU, trial 95A | +8 mm |

Y retains the old drawing's 92 mm barrel, 200 mm flange diameter and 60 mm overall envelope. It is a new printable design, not a reproduction of the machined POM drawing. Clear winding width is 48 mm, with a small chamfer at the bottom flange. Approximately five layers of neatly wound 3 mm rope provide room for 20 m plus retained wraps; actual packing must be checked.

X has a 28 mm long traction sleeve, 92 mm finished traction diameter, 108 mm flanges and 42 mm overall width. The sleeve stretches over an 88 mm rigid barrel (87.6 mm nominal sleeve ID). This is a friction-fit sleeve, not a positive torque interlock. Test for sleeve slip and wet-rope grip in both directions. Wrap routing and outgoing rope tension remain part of the installation design.

## Metal hub interface — not yet final

Intended candidate: Nexus Robot **18014**, stainless flanged hub, nominal 14 mm keyed shaft bore and 56.8 mm flange OD:
https://www.nexusrobot.com/product/14mm-stainless-steel-key-hub-18014.html

Both printed bodies have an 8 mm thick mounting web, 15 mm shaft clearance hole, and a 57.4 mm diameter x 6 mm deep flange locating recess on the inside. The hub flange seats against the web at Z=8 mm. Larger cavities above the recess accommodate the hub boss. The 15 mm printed hole does not transmit shaft torque; the metal hub does.

**The hub axial dimensions, bolt pattern, keyway and retention details have not been verified.** Confirm them before committing to full-size prints. Hub orientation and gearbox-face clearance must also be checked against the 34 mm shaft projection. Modify the profile if the actual flange thickness, boss or retaining screw requires more clearance. The removable flange has a 60 mm access opening.

Hub bolt holes are deliberately absent. Transfer the actual flange pattern onto the printed web, drill appropriate clearance holes, and bolt to the metal hub using load-spreading washers/backing plate. Confirm bolt engagement and access; do not rely on threads tapped into the print. The printed locating recess is a concentricity aid, not a press fit.

## Removable flange fastening

Use six axial M4 tie bolts through the barrel wall and both flanges, with washers and nuts. The holes are deliberately left for drilling, avoiding unsupported horizontal holes when printing the removable flange flat.

- Y: six 4.5 mm clearance holes on **82 mm PCD**, 60 mm total assembly thickness.
- X: six 4.5 mm clearance holes on **78 mm PCD**, 42 mm total assembly thickness.
- Select bolt lengths to accommodate washers/nuts beyond these thicknesses (approximately M4x70 and M4x50 respectively; confirm hardware stack).
- Included SVG templates must be printed at **100% / actual size**; verify their 50 mm scale bars.
- Seat the removable flange, mark from the template, then drill axially with suitable tooling. Keep holes aligned within the barrel wall. Metal compression sleeves can reduce creep under bolt preload; enlarging holes for sleeves requires checking remaining wall thickness.

The retaining bolts are separate from the metal hub's mounting bolts.

## Printing

- Bodies: print with the large flat flange on the bed, shaft axis vertical.
- Removable flanges: flip so the broad flat face is on the bed and the locating spigot points up.
- Sleeve: print standing on its annular end.
- Start with 6–8 perimeters, substantial top/bottom layers, and locally solid hub/bolt regions. Check slicer preview for continuous barrel walls. Settings are starting points, not strength certification.
- Round and smooth every rope-contact edge; remove seams and burrs.
- Y requires a bed that can accommodate a 200 mm part plus any brim.
- Fit-check the hub recess and flange spigots with small test prints before printing complete bodies.

The old drum's external geometry comes from `sv-docs/dawinchi/drum-115669.pdf`. Direct gearbox mounting remains the intended arrangement; shaft radial-load capacity and printed assembly strength have not been validated by these meshes.

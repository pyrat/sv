"""FIT0737 tunnel cradle; millimetres. Approximate envelope, physical fit untested."""
from pathlib import Path
import importlib.util
import json

import numpy as np
import trimesh

OUT = Path(__file__).parent
spec = importlib.util.spec_from_file_location(
    "jetson_clip", OUT.parent / "jetson-orin-din" / "generate.py")
clip = importlib.util.module_from_spec(spec)
spec.loader.exec_module(clip)
box, mesh = clip.box, clip.mesh

# Adapter length along Y; depth along X; width along Z (rail length).
LENGTH, WIDTH, DEPTH = 83.0, 26.0, 17.0
CLEARANCE = 0.7  # each side; dimensions in the source PDF are only approximate
WALL = 3.0
START = 8.0
INNER_W = WIDTH + 2 * CLEARANCE
INNER_D = DEPTH + 2 * CLEARANCE
END = START + LENGTH + 2 * CLEARANCE
OUTER_W = INNER_W + 2 * WALL


def rail():
    r = box((1, 25, 54), (-13.5, 37.5, -10))
    for y in (37.5, 61.5):
        r += box((7.5, 1, 54), (-13.5, y, -10))
    for y in (32.5, 62.5):
        r += box((1, 5, 54), (-7, y, -10))
    return r


def make_parts():
    body, slider = clip.make_clip()
    # Close the underside as well: a roof alone permits escape toward -Z.
    body += box((12, 12.8, 2), (-8, 10, -2.4))
    body += box((2.6, 12.8, 3), (-8, 10, -2.4))
    body += box((1, 12.8, 3), (3, 10, -2.4))
    # Backplate joins the existing spine with a broad, overlapping web.
    body += box((WALL, END - START, OUTER_W), (5, START, 0))
    body += box((7.5, 35.5, 24), (-1.5, 32.5, 0))
    body += box((3, 20, 7.4), (3, 14, 0))
    for z in (0, WALL + INNER_W):
        body += box((8, END - START, WALL), (5, START, z))
        # Short full-depth tunnel; open ends leave both connectors accessible.
        body += box((WALL + INNER_D, 12, WALL), (5, 44, z))
    body += box((WALL, 12, OUTER_W), (8 + INNER_D, 44, 0))
    # Two pairs of slots let ties wrap around the device AND backplate.
    # Located away from the rail spine, so ties cannot foul the latch.
    for y in (22, 78):
        for z in (6, OUTER_W - 8):
            body -= box((5, 4.5, 2), (4, y, z))
    return body, slider


def main():
    body, slider = make_parts()
    envelope = box((DEPTH, LENGTH, WIDTH),
                   (8 + CLEARANCE, START + CLEARANCE, WALL + CLEARANCE))
    for name, part in (("body", body), ("slider", slider)):
        assert part.status() == clip.m.Error.NoError, name
        assert len(part.decompose()) == 1, name
        assert mesh(part).is_volume, name
    for other in (slider, rail(), envelope):
        assert (body ^ other).volume() < 1e-6
    assert (slider ^ rail()).volume() < 1e-6
    # Pressing the flexible tooth is not simulated; check the rigid carriage.
    rigid = slider - box((4, 20, 7), (-1.6, 8, -1))
    for travel in np.linspace(0, 6, 61):
        moved = rigid.translate((0, -float(travel), 0))
        assert (body ^ moved).volume() < 1e-6, travel
        assert (rail() ^ moved).volume() < 1e-6, travel
    assert (body ^ slider.translate((0, -1, 0))).volume() > 0
    for travel in (-0.5, 6.5):
        assert (body ^ rigid.translate((0, -travel, 0))).volume() > 0
    # Pin/slot sidewalls and guide roof obstruct lateral and upward escape.
    for shift in ((1, 0, 0), (-1, 0, 0), (0, 0, 1), (0, 0, -1)):
        assert (body ^ rigid.translate(shift)).volume() > 0
    assembly = mesh(body + slider)
    assembly.apply_translation((0, 0, 2.4))
    assembly.export(OUT / "fit0737-din-print-in-place.stl")
    saved = trimesh.load_mesh(OUT / "fit0737-din-print-in-place.stl")
    assert saved.is_volume and len(saved.split()) == 2
    scene = trimesh.Scene()
    for name, part, colour in (
        ("cradle", body, [60, 100, 140, 255]),
        ("captive-latch", slider, [240, 150, 35, 255]),
        ("rail-reference", rail(), [155, 160, 165, 255]),
        ("approximate-adapter-envelope", envelope, [45, 130, 70, 150]),
    ):
        tri = mesh(part)
        tri.visual.vertex_colors = colour
        tri.apply_scale(0.001)
        scene.add_geometry(tri, node_name=name)
    scene.export(OUT / "assembly.glb")
    report = {
        "status": "prototype; physical fit and flexure performance untested",
        "adapter_envelope_mm": [LENGTH, WIDTH, DEPTH],
        "tunnel_clearance_per_side_mm": CLEARANCE,
        "print_extents_mm": assembly.extents.round(3).tolist(),
        "plastic_volume_cm3": round(assembly.volume / 1000, 2),
        "watertight_shells": 2,
        "rigid_slider_travel_mm": [0, 6],
        "travel_samples": 61,
        "nominal_rail_and_adapter_interference_mm3": 0,
        "lock_end_stops_sidewalls_and_roof_checked": True,
        "flexure_deformation_simulated": False,
        "rail_axial_retention": "requires external DIN rail end stops",
        "adapter_axial_retention": "requires cable ties; tunnel alone is open-ended",
    }
    (OUT / "validation.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps(report, indent=2))


if __name__ == "__main__":
    main()

"""Edge-on Orin Nano DIN bracket, millimetres. Prototype: physical fit untested.

Print coordinates: PCB in XY, fan side +Z, GPIO +Y, USB +X.
DIN clip on the bottom 79 mm edge (-Y); rail length parallel to Z.
"""
from pathlib import Path
import json
import manifold3d as m
import numpy as np
import trimesh

OUT = Path(__file__).parent
GAP = 0.4
PILOT_D = 2.15
PCB_Z = 14.0
HOLES = [(4, 4), (62, 4), (4, 96), (62, 96)]


def box(size, at=(0, 0, 0)):
    return m.Manifold.cube(size).translate(at)


def cyl(radius, height, at):
    return m.Manifold.cylinder(height, radius, radius, 64).translate(at)


def mesh(s):
    data = s.to_mesh()
    return trimesh.Trimesh(np.asarray(data.vert_properties)[:, :3],
                            np.asarray(data.tri_verts), process=True)


def short_edge(s):
    """Map clip-local axes to the short edge, centred at PCB X=39.5."""
    return s.rotate((0, 0, 90)).translate((89.5, -2, 0))


def make_parts():
    # Open ladder frame: 10 mm component clearance above its 4 mm ribs.
    body = box((5.6, 100, 4), (1.4, 0, 0))
    body += box((8, 100, 4), (58, 0, 0))
    body += box((75.1, 8, 4), (1.4, 0, 0))
    body += box((64.6, 8, 4), (1.4, 92, 0))
    for x, y in HOLES:
        body += cyl(3, PCB_Z, (x, y, 0))
        body -= cyl(PILOT_D / 2, 9, (x, y, PCB_Z - 8))

    clip, slider = make_clip()
    # Connect the relocated spine to the bottom frame rib. The 2 mm setback
    # keeps the moving latch clear of the bottom-right standoff and frame.
    body += box((35.5, 4.5, 4), (21.5, -0.5, 0))
    body += short_edge(clip)
    return body, short_edge(slider)


def make_clip():
    # Clip-local geometry; short_edge() rotates the whole mechanism together.

    # Rail length is Z; flange front face X=-6, edges Y=32.5 and 67.5.
    # Deep spine spreads the overturning load along 24 mm of rail.
    body = box((7.8, 35.5, 4), (-6, 32.5, 0))
    body += box((5, 35.5, 24), (-6, 32.5, 0))
    body += box((9, 3, 24), (-10, 68, 0))
    body += box((2, 5.5, 24), (-10, 65.5, 0))

    # Guide roof bridges only 6.8 mm; slider and chassis both start on bed.
    # Include the lock shoulder formerly supplied by the long frame rib.
    body += box((2.6, 23, 7.4), (1.4, 6, 0))
    body += box((2.6, 12.8, 7.4), (-8, 10, 0))
    body += box((12, 12.8, 2), (-8, 10, 5 + GAP))
    # Captive pin through longitudinal slider slot: 6 mm useful travel.
    body += box((1.6, 1.6, 7.4), (-3.2, 12.4, 0))
    # Positive latch pocket in the right guide wall, accessible at its end.
    body -= box((1.2, 4.8, 5.4), (1.4, 7.6, 0))

    slider = box((6, 21.8, 5), (-5, 10, 0))
    slider -= box((2.4, 8, 7), (-3.6, 12, -1))
    # Integral 1.2 mm-wide cantilever, 18 mm free length; press toward -X.
    # 1.4 mm flexure relief allows the tooth's required ~1 mm retraction.
    slider -= box((1.4, 18, 7), (-1.6, 10, -1))
    slider += box((2.4, 4, 5), (-0.2, 8, 0))
    # Moving lower hook; its outside elbow is clear of the guide at full travel.
    slider += box((5.2, 2.3, 5), (-10, 29.5, 0))
    slider += box((2, 5, 5), (-10, 29.5, 0))
    return body, slider


def rail():
    # Nominal 35 x 7.5 mm rail with 1 mm stock, for interference checks.
    r = box((1, 25, 44), (-13.5, 37.5, -10))
    for y in (37.5, 61.5):
        r += box((7.5, 1, 44), (-13.5, y, -10))
    for y in (32.5, 62.5):
        r += box((1, 5, 44), (-7, y, -10))
    return short_edge(r)


def main():
    body, slider = make_parts()
    for name, part in [('body', body), ('slider', slider)]:
        assert part.status() == m.Error.NoError, name
        assert len(part.decompose()) == 1, name
        assert mesh(part).is_volume, name
    assert (body ^ slider).volume() < 1e-6
    assert (body ^ rail()).volume() < 1e-6
    assert (slider ^ rail()).volume() < 1e-6
    pcb = box((79, 100, 1.57), (0, 0, PCB_Z))
    assert (body ^ pcb).volume() < 1e-6

    # Test rigid guide/hook travel; locking tongue must be pressed separately.
    rigid = slider - short_edge(box((4, 20, 7), (-1.6, 8, -1)))
    for travel in np.linspace(0, 6, 61):
        moved = rigid.translate((float(travel), 0, 0))
        assert (body ^ moved).volume() < 1e-6, travel
        assert (rail() ^ moved).volume() < 1e-6, travel
    # Locking tooth really blocks uncommanded release; captive stop blocks overtravel.
    assert (body ^ slider.translate((1, 0, 0))).volume() > 0
    assert (body ^ rigid.translate((6.5, 0, 0))).volume() > 0
    assert (body ^ rigid.translate((-0.5, 0, 0))).volume() > 0

    assembly = mesh(body + slider)
    assert len(assembly.split()) == 2
    assembly.export(OUT / 'jetson-orin-din-print-in-place.stl')
    saved = trimesh.load_mesh(OUT / 'jetson-orin-din-print-in-place.stl')
    assert saved.is_volume and len(saved.split()) == 2

    # Colour-separated geometry-only viewer: PCB envelope is a visual reference.
    scene = trimesh.Scene()
    for name, part, colour in [
        ('bracket', body, [60, 100, 140, 255]),
        ('captive-slider', slider, [240, 150, 35, 255]),
        ('DIN-rail-reference', rail(), [155, 160, 165, 255]),
        ('PCB-envelope-reference', box((79, 100, 1.57), (0, 0, PCB_Z)),
         [45, 130, 70, 130]),
    ]:
        tri = mesh(part)
        tri.visual.vertex_colors = colour
        tri.apply_scale(0.001)
        scene.add_geometry(tri, node_name=name)
    scene.export(OUT / 'assembly.glb')
    report = {
        'status': 'prototype; no physical print, fit, thermal or load testing',
        'units': 'mm (STL); metres (GLB)',
        'mount_edge': 'bottom 79 mm short edge, opposite GPIO',
        'slider_release_direction': '+X (right from fan side)',
        'print_extents_mm': assembly.extents.round(3).tolist(),
        'plastic_volume_cm3': round(assembly.volume / 1000, 2),
        'watertight_shells': 2,
        'shells': 'one chassis and one captive slider in a single print',
        'rigid_slider_travel_checked_mm': [0, 6],
        'travel_samples': 61,
        'nominal_rail_interference_mm3': 0,
        'positive_lock_and_end_stops_checked': True,
        'flexure_deformation_simulated': False,
        'pcb_holes_xy_mm': HOLES,
        'pilot_diameter_mm': PILOT_D,
        'pcb_underside_z_mm': PCB_Z,
    }
    (OUT / 'validation.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
